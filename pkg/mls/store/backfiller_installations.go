package store

import (
	"context"
	"database/sql"
	"sync"
	"time"

	"github.com/xmtp/xmtp-node-go/pkg/mls/store/queries"
	"go.uber.org/zap"
)

// InstallationsBackfiller is responsible for:
// - Appending to key_packages installations that have not been appended yet.
// - Updating the is_appended status of installations that have not been appended yet.
//
// The backfill process runs in a loop and periodically:
//  1. Starts a transaction
//  2. Selects up to N unprocessed installations (e.g., where is_appended IS NULL)
//     Skipping locked rows, to guarantee that no two workers process the same row.
//  3. Inserts the key package into the key_packages table
//  4. Updates the is_appended status of the installation
//
// If no work is available, it sleeps briefly before retrying.
//
// Because of how installations table works, HA is guaranteed:
// - Rows are inserted into installations via CreateOrUpdateInstallation, which
// has been modified accordingly to update the is_appended status to true, and to insert
// the key package into the key_packages table.
// This guarantees any new or updated installations makes its way into the key_packages table.
// - Meanwhile, the backfiller works by selecting installations with a NULL is_appended,
// inserting them into key_packages and updating the is_appended status to true.
// - InsertKeyPackage guarantees that on conflict with (installation_id, key_package) nothing happens.
// - UpdateIsAppendedStatus guarantees that the is_appended status is updated to true.
//
// Important!
// Final ordering in key_packages is not needed, as this design guarantees that the newest
// key package is always inserted in last position in key_packages.
type InstallationsBackfiller struct {
	ctx               context.Context
	cancel            context.CancelFunc
	wg                sync.WaitGroup
	log               *zap.Logger
	db                *sql.DB
}

var _ Backfiller = (*InstallationsBackfiller)(nil)

func NewInstallationsBackfiller(ctx context.Context, db *sql.DB,
	log *zap.Logger,
) *InstallationsBackfiller {
	ctx, cancel := context.WithCancel(ctx)

	return &InstallationsBackfiller{
		ctx:               ctx,
		cancel:            cancel,
		log:               log,
		db:                db,
	}
}

func (b *InstallationsBackfiller) Run() {
	b.wg.Add(1)
	go func() {
		defer b.wg.Done()
		for {
			select {
			case <-b.ctx.Done():
				return
			default:
				foundMessages := true
				err := RunInTx(
					b.ctx,
					b.db,
					nil,
					func(ctx context.Context, querier *queries.Queries) error {
						installations, err := querier.SelectInstallationsToBackfill(ctx)
						if err != nil {
							b.log.Error(
								"could not select next batch for installations backfill processing",
								zap.Error(err),
							)
							return err
						}

						if len(installations) == 0 {
							b.log.Info("No installations to backfill")
							foundMessages = false
							return nil
						}

						for _, installation := range installations {
								err := RunInTx(ctx, b.db, nil, func(ctx context.Context, querier *queries.Queries) error {
									err = querier.InsertKeyPackage(ctx, queries.InsertKeyPackageParams{
										InstallationID: installation.ID,
										KeyPackage:     installation.KeyPackage,
									})
									if err != nil {
										b.log.Error("error inserting key package", zap.Error(err))
										return err
									}

									err = querier.UpdateIsAppendedStatus(ctx, queries.UpdateIsAppendedStatusParams{
										IsAppended: sql.NullBool{
											Bool:  true,
											Valid: true,
										},
										ID: installation.ID,
									})
									if err != nil {
										b.log.Error("error updating is_appended status", zap.Error(err))
										return err
									}

									return nil
								})
								if err != nil {
									b.log.Error("error backfilling installation", zap.Error(err))
									return err
								}
						}

						return nil
					},
				)
				if err != nil {
					b.log.Error("Failed to execute installations backfill cycle", zap.Error(err))
					time.Sleep(1 * time.Second)
				} else if !foundMessages {
					time.Sleep(1 * time.Minute)
				}
			}
		}
	}()
}

func (b *InstallationsBackfiller) Close() {
	b.cancel()
	b.wg.Wait()
}
