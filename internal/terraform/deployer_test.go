package terraform

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/hashicorp/go-tfe"
	"github.com/hashicorp/go-tfe/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDeployer_New(t *testing.T) {
	log, err := zap.NewDevelopment()
	require.NoError(t, err)
	ctx := context.Background()
	tfc, err := tfe.NewClient(&tfe.Config{
		Token: "token",
	})
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	m := mocks.NewMockWorkspaces(ctrl)
	tfc.Workspaces = m
	require.NoError(t, err)
	m.EXPECT().Read(ctx, "organization", "workspace").Return(&tfe.Workspace{}, nil)
	client, err := NewDeployer(ctx, log, tfc, "organization", "workspace")
	require.NoError(t, err)
	require.NotNil(t, client)
}

func TestDeployer_Deploy(t *testing.T) {
	// TODO: expect wait until applied/canceled/failed
}
