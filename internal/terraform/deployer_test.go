package terraform

import (
	"context"
	"testing"
	"time"

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
	deployer, err := NewDeployer(ctx, log, tfc, &Config{
		Organization: "organization",
		Workspace:    "workspace",
	})
	require.NoError(t, err)
	require.NotNil(t, deployer)
}

func TestDeployer_Deploy(t *testing.T) {
	tcs := []struct {
		name   string
		config *Config
		mock   func(*mocks.MockWorkspaces, *mocks.MockVariables, *mocks.MockRuns)
		expect func(*testing.T, error)
	}{
		{
			name: "run applied",
			config: &Config{
				Organization: "organization",
				Workspace:    "workspace",
				WaitDelay:    1 * time.Millisecond,
			},
			mock: func(mw *mocks.MockWorkspaces, mv *mocks.MockVariables, mr *mocks.MockRuns) {
				wsp := &tfe.Workspace{
					ID: "workspace-id",
				}
				mw.EXPECT().Read(gomock.Any(), "organization", "workspace").Return(wsp, nil)

				mv.EXPECT().List(gomock.Any(), "workspace-id", nil).Return(&tfe.VariableList{
					Items: []*tfe.Variable{
						{
							ID:    "var-id",
							Key:   "var-key",
							Value: "var-value",
						},
					},
				}, nil)
				updateOpts := tfe.VariableUpdateOptions{
					Value: stringPtr("new-var-value"),
				}
				mv.EXPECT().Update(gomock.Any(), "workspace-id", "var-id", updateOpts).Return(nil, nil)

				mr.EXPECT().Create(gomock.Any(), tfe.RunCreateOptions{
					Message:   stringPtr("commit"),
					Workspace: wsp,
					AutoApply: boolPtr(true),
				}).Return(&tfe.Run{
					ID:     "run-id",
					Status: tfe.RunPending,
				}, nil)
				mr.EXPECT().Read(gomock.Any(), "run-id").Return(&tfe.Run{
					ID:     "run-id",
					Status: tfe.RunApplying,
				}, nil)
				mr.EXPECT().Read(gomock.Any(), "run-id").Return(&tfe.Run{
					ID:     "run-id",
					Status: tfe.RunApplied,
				}, nil)
			},
			expect: func(t *testing.T, err error) {
				require.NoError(t, err)
			},
		},
		{
			name: "run canceled",
			config: &Config{
				Organization: "organization",
				Workspace:    "workspace",
				WaitDelay:    1 * time.Millisecond,
			},
			mock: func(mw *mocks.MockWorkspaces, mv *mocks.MockVariables, mr *mocks.MockRuns) {
				wsp := &tfe.Workspace{
					ID: "workspace-id",
				}
				mw.EXPECT().Read(gomock.Any(), "organization", "workspace").Return(wsp, nil)

				mv.EXPECT().List(gomock.Any(), "workspace-id", nil).Return(&tfe.VariableList{
					Items: []*tfe.Variable{
						{
							ID:    "var-id",
							Key:   "var-key",
							Value: "var-value",
						},
					},
				}, nil)
				updateOpts := tfe.VariableUpdateOptions{
					Value: stringPtr("new-var-value"),
				}
				mv.EXPECT().Update(gomock.Any(), "workspace-id", "var-id", updateOpts).Return(nil, nil)

				mr.EXPECT().Create(gomock.Any(), tfe.RunCreateOptions{
					Message:   stringPtr("commit"),
					Workspace: wsp,
					AutoApply: boolPtr(true),
				}).Return(&tfe.Run{
					ID:     "run-id",
					Status: tfe.RunPending,
				}, nil)
				mr.EXPECT().Read(gomock.Any(), "run-id").Return(&tfe.Run{
					ID:     "run-id",
					Status: tfe.RunApplying,
				}, nil)
				mr.EXPECT().Read(gomock.Any(), "run-id").Return(&tfe.Run{
					ID:     "run-id",
					Status: tfe.RunCanceled,
				}, nil)
			},
			expect: func(t *testing.T, err error) {
				require.EqualError(t, err, "waiting on run: run canceled with status \"canceled\"")
			},
		},
		{
			name: "wait timeout",
			config: &Config{
				Organization: "organization",
				Workspace:    "workspace",
				WaitDelay:    1 * time.Millisecond,
				WaitTimeout:  10 * time.Millisecond,
			},
			mock: func(mw *mocks.MockWorkspaces, mv *mocks.MockVariables, mr *mocks.MockRuns) {
				wsp := &tfe.Workspace{
					ID: "workspace-id",
				}
				mw.EXPECT().Read(gomock.Any(), "organization", "workspace").Return(wsp, nil)

				mv.EXPECT().List(gomock.Any(), "workspace-id", nil).Return(&tfe.VariableList{
					Items: []*tfe.Variable{
						{
							ID:    "var-id",
							Key:   "var-key",
							Value: "var-value",
						},
					},
				}, nil)
				updateOpts := tfe.VariableUpdateOptions{
					Value: stringPtr("new-var-value"),
				}
				mv.EXPECT().Update(gomock.Any(), "workspace-id", "var-id", updateOpts).Return(nil, nil)

				mr.EXPECT().Create(gomock.Any(), tfe.RunCreateOptions{
					Message:   stringPtr("commit"),
					Workspace: wsp,
					AutoApply: boolPtr(true),
				}).Return(&tfe.Run{
					ID:     "run-id",
					Status: tfe.RunPending,
				}, nil)
				mr.EXPECT().Read(gomock.Any(), "run-id").Return(&tfe.Run{
					ID:     "run-id",
					Status: tfe.RunApplying,
				}, nil).AnyTimes()
			},
			expect: func(t *testing.T, err error) {
				require.EqualError(t, err, "waiting on run: timeout waiting for run")
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log, err := zap.NewDevelopment()
			require.NoError(t, err)
			ctx := context.Background()
			tfc, err := tfe.NewClient(&tfe.Config{
				Token: "token",
			})
			require.NoError(t, err)

			ctrl := gomock.NewController(t)

			workspaces := mocks.NewMockWorkspaces(ctrl)
			tfc.Workspaces = workspaces

			variables := mocks.NewMockVariables(ctrl)
			tfc.Variables = variables

			runs := mocks.NewMockRuns(ctrl)
			tfc.Runs = runs

			tc.mock(workspaces, variables, runs)

			deployer, err := NewDeployer(ctx, log, tfc, tc.config)
			require.NoError(t, err)
			require.NotNil(t, deployer)

			err = deployer.Deploy("var-key", "new-var-value", "commit")
			tc.expect(t, err)
		})
	}
}

func stringPtr(v string) *string {
	return &v
}
