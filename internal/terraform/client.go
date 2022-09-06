package terraform

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/hashicorp/go-tfe"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

var (
	runWaitTimeout = 10 * time.Minute
	runWaitDelay   = 5 * time.Second
)

type client struct {
	*tfe.Client
	ctx context.Context
	wsp *tfe.Workspace
	log *zap.Logger
}

type Config struct {
	Token        string
	Workspace    string
	Organization string
}

func NewClient(log *zap.Logger, config *Config) (*client, error) {
	tfc, err := tfe.NewClient(&tfe.Config{
		Token: config.Token,
	})
	if err != nil {
		return nil, errors.Wrap(err, "creating client")
	}

	ctx := context.Background()
	wsp, err := tfc.Workspaces.Read(ctx, config.Organization, config.Workspace)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("getting workspace %s/%s", config.Organization, config.Workspace))
	}

	c := &client{
		Client: tfc,
		ctx:    ctx,
		wsp:    wsp,
	}

	return c, nil
}

func (c *client) UpdateVar(name string, value string) error {
	vars, err := c.Variables.List(c.ctx, c.wsp.ID, nil)
	if err != nil {
		return errors.Wrap(err, "listing workspace vars")
	}

	for _, v := range vars.Items {
		if v.Key == name {
			_, err := c.Variables.Update(c.ctx, c.wsp.ID, v.ID, tfe.VariableUpdateOptions{Value: &value})
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("updating variable %s", name))
			}
			return nil
		}
	}
	return fmt.Errorf("variable %s not found", name)
}

func (c client) StartRun(commit string, apply bool) error {
	msg := fmt.Sprintf("triggered from commit %s", commit)
	out, err := exec.Command("git", "log", "--oneline", "-n 1").Output()
	if err == nil {
		msg = string(out)
	}

	run, err := c.Runs.Create(c.ctx, tfe.RunCreateOptions{
		Message:   &msg,
		Workspace: c.wsp,
		AutoApply: &apply,
	})
	if err != nil {
		return errors.Wrap(err, "creating run")
	}

	err = c.runWait(run.ID)
	if err != nil {
		return errors.Wrap(err, "waiting on run")
	}

	return nil
}

func (c client) runWait(runID string) error {
	started := time.Now()
	for {
		run, err := c.Runs.Read(c.ctx, runID)
		if err != nil {
			return errors.Wrap(err, "reading run")
		}

		switch run.Status {
		case tfe.RunApplied:
			c.log.Info("success", zap.String("status", string(run.Status)))
			return nil
		case tfe.RunErrored:
			c.log.Info("failed", zap.String("status", string(run.Status)))
			return fmt.Errorf("run failed with status %q", run.Status)
		case tfe.RunDiscarded:
			c.log.Info("canceled", zap.String("status", string(run.Status)))
			return fmt.Errorf("run canceled with status %q", run.Status)
		case tfe.RunCanceled:
			c.log.Info("canceled", zap.String("status", string(run.Status)))
			return fmt.Errorf("run canceled with status %q", run.Status)
		default:
			c.log.Info("waiting", zap.String("status", string(run.Status)))
		}

		if time.Since(started) > runWaitTimeout {
			return fmt.Errorf("timeout waiting for run")
		}

		time.Sleep(runWaitDelay)
	}
}
