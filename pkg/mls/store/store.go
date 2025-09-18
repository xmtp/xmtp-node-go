package store

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/metrics"
	"github.com/xmtp/xmtp-node-go/pkg/utils"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
	migrations "github.com/xmtp/xmtp-node-go/pkg/migrations/mls"
	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"github.com/xmtp/xmtp-node-go/pkg/mlsvalidate"
	identity "github.com/xmtp/xmtp-node-go/pkg/proto/identity/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/proto/identity/associations"
	mlsv1 "github.com/xmtp/xmtp-node-go/pkg/proto/mls/api/v1"
	"github.com/xmtp/xmtp-node-go/pkg/types"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

const maxPageSize = 100

type Store struct {
	log       *zap.Logger
	db        *bun.DB
	queries   *queries.Queries
	readstore *ReadStore
}

// The ReadMlsStore interface is used for all queries, and will work whether connected to the primary DB or
// a read replica
type ReadMlsStore interface {
	GetInboxLogs(
		ctx context.Context,
		req *identity.GetIdentityUpdatesRequest,
	) (*identity.GetIdentityUpdatesResponse, error)
	GetInboxIds(
		ctx context.Context,
		req *identity.GetInboxIdsRequest,
	) (*identity.GetInboxIdsResponse, error)
	FetchKeyPackages(
		ctx context.Context,
		installationIds [][]byte,
	) ([]queries.FetchKeyPackagesRow, error)
	QueryGroupMessagesV1(
		ctx context.Context,
		query *mlsv1.QueryGroupMessagesRequest,
	) (*mlsv1.QueryGroupMessagesResponse, error)
	QueryWelcomeMessagesV1(
		ctx context.Context,
		query *mlsv1.QueryWelcomeMessagesRequest,
	) (*mlsv1.QueryWelcomeMessagesResponse, error)
	QueryCommitLog(
		ctx context.Context,
		query *mlsv1.QueryCommitLogRequest,
	) (*mlsv1.QueryCommitLogResponse, error)
	GetNewestGroupMessage(
		ctx context.Context,
		groupIds [][]byte,
	) ([]*queries.GetNewestGroupMessageRow, error)
	GetNewestGroupMessageMetadata(
		ctx context.Context,
		groupIds [][]byte,
	) ([]*queries.GetNewestGroupMessageMetadataRow, error)
	Queries() *queries.Queries
}

// The ReadWriteMlsStore is a superset of the ReadMlsStore interface and is used for all writes.
// It must be connected to the primary DB, and will use the primary for both reads and writes.
type ReadWriteMlsStore interface {
	ReadMlsStore
	CreateOrUpdateInstallation(ctx context.Context, installationId []byte, keyPackage []byte) error
	PublishIdentityUpdate(
		ctx context.Context,
		req *identity.PublishIdentityUpdateRequest,
		validationService mlsvalidate.MLSValidationService,
	) (*identity.PublishIdentityUpdateResponse, error)
	InsertGroupMessage(
		ctx context.Context,
		groupId []byte,
		data []byte,
		senderHmac []byte,
		shouldPush bool,
		isCommit bool,
	) (*queries.GroupMessage, error)
	InsertWelcomeMessage(
		ctx context.Context,
		installationId []byte,
		data []byte,
		hpkePublicKey []byte,
		algorithm types.WrapperAlgorithm,
		welcomeMetadata []byte,
	) (*queries.WelcomeMessage, error)
	InsertWelcomePointerMessage(
		ctx context.Context,
		installationKey []byte,
		welcomePointerData []byte,
		hpkePublicKey []byte,
		wrapperAlgorithm types.WrapperAlgorithm,
	) (*queries.WelcomeMessage, error)
	InsertCommitLog(
		ctx context.Context,
		groupId []byte,
		serialized_entry []byte,
		serialized_signature []byte,
	) (queries.CommitLogV2, error)
}

func New(ctx context.Context, log *zap.Logger, db *bun.DB) (*Store, error) {
	s := &Store{
		log:     log.Named("mlsstore"),
		db:      db,
		queries: queries.New(db.DB),
		// Create the read store with the same database connection as the write store
		readstore: NewReadStore(log, db),
	}

	if err := s.migrate(ctx); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Store) Queries() *queries.Queries {
	return s.queries
}

func (s *Store) PublishIdentityUpdate(
	ctx context.Context,
	req *identity.PublishIdentityUpdateRequest,
	validationService mlsvalidate.MLSValidationService,
) (*identity.PublishIdentityUpdateResponse, error) {
	newUpdate := req.GetIdentityUpdate()
	if newUpdate == nil {
		return nil, errors.New("IdentityUpdate is required")
	}

	var numBytes int

	if err := s.RunInRepeatableReadTx(ctx, 3, func(ctx context.Context, txQueries *queries.Queries) error {
		inboxId := newUpdate.GetInboxId()
		// We use a pg_advisory_lock to lock the inbox_id instead of SELECT FOR UPDATE
		// This allows the lock to be enforced even when there are no existing `inbox_log`s
		if err := txQueries.LockInboxLog(ctx, inboxId); err != nil {
			return err
		}

		log := s.log.With(zap.String("inbox_id", inboxId))
		inboxLogEntries, err := txQueries.GetAllInboxLogs(ctx, inboxId)
		if err != nil {
			return err
		}

		if len(inboxLogEntries) >= 256 {
			return errors.New("inbox log is full")
		}

		updates := make([]*associations.IdentityUpdate, 0, len(inboxLogEntries))
		for _, log := range inboxLogEntries {
			identityUpdate := &associations.IdentityUpdate{}
			if err := proto.Unmarshal(log.IdentityUpdateProto, identityUpdate); err != nil {
				return err
			}
			updates = append(updates, identityUpdate)
		}

		state, err := validationService.GetAssociationState(ctx, updates, []*associations.IdentityUpdate{newUpdate})
		if err != nil {
			return err
		}

		protoBytes, err := proto.Marshal(newUpdate)
		if err != nil {
			return err
		}

		sequence_id, err := txQueries.InsertInboxLog(ctx, queries.InsertInboxLogParams{
			InboxID:             inboxId,
			ServerTimestampNs:   nowNs(),
			IdentityUpdateProto: protoBytes,
		})

		log.Info("Inserted inbox log", zap.Any("sequence_id", sequence_id))

		if err != nil {
			return err
		}

		for _, new_member := range state.StateDiff.NewMembers {
			log.Info("New member", zap.Any("member", new_member))
			if address, ok := new_member.Kind.(*associations.MemberIdentifier_EthereumAddress); ok {
				_, err = txQueries.InsertAddressLog(ctx, queries.InsertAddressLogParams{
					Address:               address.EthereumAddress,
					InboxID:               inboxId,
					AssociationSequenceID: sql.NullInt64{Valid: true, Int64: sequence_id},
					RevocationSequenceID:  sql.NullInt64{Valid: false},
				})
				if err != nil {
					return err
				}
			}
		}

		for _, removed_member := range state.StateDiff.RemovedMembers {
			log.Info("Removed member", zap.Any("member", removed_member))
			if address, ok := removed_member.Kind.(*associations.MemberIdentifier_EthereumAddress); ok {
				err = txQueries.RevokeAddressFromLog(ctx, queries.RevokeAddressFromLogParams{
					Address:              address.EthereumAddress,
					InboxID:              inboxId,
					RevocationSequenceID: sql.NullInt64{Valid: true, Int64: sequence_id},
				})
				if err != nil {
					return err
				}
			}
		}

		err = txQueries.TouchInbox(ctx, inboxId)
		if err != nil {
			return err
		}

		numBytes = len(protoBytes)

		return nil
	}); err != nil {
		return nil, err
	}

	metrics.EmitMLSSentIdentityUpdate(ctx, s.log, numBytes)

	return &identity.PublishIdentityUpdateResponse{}, nil
}

// Creates the installation and last resort key package
func (s *Store) CreateOrUpdateInstallation(
	ctx context.Context,
	installationId []byte,
	keyPackage []byte,
) error {
	now := nowNs()

	err := RunInTx(ctx, s.db.DB, nil, func(ctx context.Context, querier *queries.Queries) error {
		err := querier.CreateOrUpdateInstallation(ctx, queries.CreateOrUpdateInstallationParams{
			ID:         installationId,
			CreatedAt:  now,
			UpdatedAt:  now,
			KeyPackage: keyPackage,
			IsAppended: sql.NullBool{
				Bool:  true,
				Valid: true,
			},
		})
		if err != nil {
			s.log.Error("error creating or updating installation", zap.Error(err))
			return err
		}

		err = querier.InsertKeyPackage(ctx, queries.InsertKeyPackageParams{
			InstallationID: installationId,
			KeyPackage:     keyPackage,
			CreatedAt:      now,
		})
		if err != nil {
			s.log.Error("error inserting key package", zap.Error(err))
			return err
		}

		return nil
	})

	return err
}

func (s *Store) InsertGroupMessage(
	ctx context.Context,
	groupId []byte,
	data []byte,
	senderHmac []byte,
	shouldPush bool,
	isCommit bool,
) (*queries.GroupMessage, error) {
	dataHash := sha256.Sum256(append(groupId, data...))
	message, err := s.queries.InsertGroupMessage(ctx, queries.InsertGroupMessageParams{
		GroupID:         groupId,
		Data:            data,
		GroupIDDataHash: dataHash[:],
		IsCommit:        isCommit,
		SenderHmac:      senderHmac,
		ShouldPush:      shouldPush,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, NewAlreadyExistsError(err)
		}
		return nil, err
	}

	return &message, nil
}

func (s *Store) InsertWelcomeMessage(
	ctx context.Context,
	installationId []byte,
	data []byte,
	hpkePublicKey []byte,
	wrapperAlgorithm types.WrapperAlgorithm,
	welcomeMetadata []byte,
) (*queries.WelcomeMessage, error) {
	dataHash := sha256.Sum256(append(installationId, data...))
	message, err := s.queries.InsertWelcomeMessage(ctx, queries.InsertWelcomeMessageParams{
		InstallationKey:         installationId,
		Data:                    data,
		InstallationKeyDataHash: dataHash[:],
		HpkePublicKey:           hpkePublicKey,
		WrapperAlgorithm:        int16(wrapperAlgorithm),
		WelcomeMetadata:         welcomeMetadata,
	})
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, NewAlreadyExistsError(err)
		}
		return nil, err
	}

	return &message, nil
}

func (s *Store) InsertWelcomePointerMessage(
	ctx context.Context,
	installationKey []byte,
	welcomePointerData []byte,
	hpkePublicKey []byte,
	wrapperAlgorithm types.WrapperAlgorithm,
) (*queries.WelcomeMessage, error) {
	if len(hpkePublicKey) == 0 {
		return nil, errors.New("hpke public key is required")
	}

	h := sha256.New()
	_, err := h.Write(installationKey)
	if err != nil {
		return nil, err
	}
	_, err = h.Write(welcomePointerData)
	if err != nil {
		return nil, err
	}
	var dataHash [32]byte
	copy(dataHash[:], h.Sum(nil))
	message, err := s.queries.InsertWelcomePointerMessage(
		ctx,
		queries.InsertWelcomePointerMessageParams{
			InstallationKey:         installationKey,
			WelcomePointerData:      welcomePointerData,
			InstallationKeyDataHash: dataHash[:],
			HpkePublicKey:           hpkePublicKey,
			WrapperAlgorithm:        int16(wrapperAlgorithm),
		},
	)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
			return nil, NewAlreadyExistsError(err)
		}
		return nil, err
	}

	return &message, nil
}

func (s *Store) GetInboxIds(
	ctx context.Context,
	req *identity.GetInboxIdsRequest,
) (*identity.GetInboxIdsResponse, error) {
	return s.readstore.GetInboxIds(ctx, req)
}

func (s *Store) GetInboxLogs(
	ctx context.Context,
	batched_req *identity.GetIdentityUpdatesRequest,
) (*identity.GetIdentityUpdatesResponse, error) {
	return s.readstore.GetInboxLogs(ctx, batched_req)
}

func (s *Store) FetchKeyPackages(
	ctx context.Context,
	installationIds [][]byte,
) ([]queries.FetchKeyPackagesRow, error) {
	return s.readstore.FetchKeyPackages(ctx, installationIds)
}

func (s *Store) QueryGroupMessagesV1(
	ctx context.Context,
	req *mlsv1.QueryGroupMessagesRequest,
) (*mlsv1.QueryGroupMessagesResponse, error) {
	return s.readstore.QueryGroupMessagesV1(ctx, req)
}

func (s *Store) QueryWelcomeMessagesV1(
	ctx context.Context,
	req *mlsv1.QueryWelcomeMessagesRequest,
) (*mlsv1.QueryWelcomeMessagesResponse, error) {
	return s.readstore.QueryWelcomeMessagesV1(ctx, req)
}

func (s *Store) GetNewestGroupMessageMetadata(
	ctx context.Context,
	groupIds [][]byte,
) ([]*queries.GetNewestGroupMessageMetadataRow, error) {
	return s.readstore.GetNewestGroupMessageMetadata(ctx, groupIds)
}

func (s *Store) GetNewestGroupMessage(
	ctx context.Context,
	groupIds [][]byte,
) ([]*queries.GetNewestGroupMessageRow, error) {
	return s.readstore.GetNewestGroupMessage(ctx, groupIds)
}

func (s *Store) InsertCommitLog(
	ctx context.Context,
	groupId []byte,
	serialized_entry []byte,
	serialized_signature []byte,
) (queries.CommitLogV2, error) {
	return s.queries.InsertCommitLogV2(ctx, queries.InsertCommitLogV2Params{
		GroupID:             groupId,
		SerializedEntry:     serialized_entry,
		SerializedSignature: serialized_signature,
	})
}

func (s *Store) migrate(ctx context.Context) error {
	migrator := migrate.NewMigrator(s.db, migrations.Migrations)
	err := migrator.Init(ctx)
	if err != nil {
		return err
	}

	group, err := migrator.Migrate(ctx)
	if err != nil {
		return err
	}

	if group.IsZero() {
		s.log.Info("No new migrations to run")
	}

	return nil
}

func nowNs() int64 {
	return time.Now().UTC().UnixNano()
}

type IdentityUpdateKind int

const (
	Create IdentityUpdateKind = iota
	Revoke
)

type IdentityUpdate struct {
	Kind               IdentityUpdateKind
	InstallationKey    []byte
	CredentialIdentity []byte
	TimestampNs        uint64
}

// Add the required methods to make a valid sort.Sort interface
type IdentityUpdateList []IdentityUpdate

func (a IdentityUpdateList) Len() int {
	return len(a)
}

func (a IdentityUpdateList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a IdentityUpdateList) Less(i, j int) bool {
	return a[i].TimestampNs < a[j].TimestampNs
}

type AlreadyExistsError struct {
	Err error
}

func (e *AlreadyExistsError) Error() string {
	return e.Err.Error()
}

func NewAlreadyExistsError(err error) *AlreadyExistsError {
	return &AlreadyExistsError{err}
}

func IsAlreadyExistsError(err error) bool {
	_, ok := err.(*AlreadyExistsError)
	return ok
}

func (s *Store) RunInRepeatableReadTx(
	ctx context.Context,
	numRetries int,
	fn func(ctx context.Context, txQueries *queries.Queries) error,
) error {
	var err error
	for i := 0; i < numRetries; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			err = RunInTx(ctx, s.db.DB, &sql.TxOptions{Isolation: sql.LevelRepeatableRead}, fn)
			if err == nil {
				return nil
			}
			s.log.Warn("Error in tx", zap.Error(err))
			utils.RandomSleep(20)
		}
	}
	return err
}

func (s *Store) QueryCommitLog(
	ctx context.Context,
	query *mlsv1.QueryCommitLogRequest,
) (*mlsv1.QueryCommitLogResponse, error) {
	return s.readstore.QueryCommitLog(ctx, query)
}
