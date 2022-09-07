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

type deployer struct {
	ctx context.Context
	log *zap.Logger
	tfe *tfe.Client
	wsp *tfe.Workspace
}

func NewDeployer(ctx context.Context, log *zap.Logger, tfc *tfe.Client, organization, workspace string) (*deployer, error) {
	wsp, err := tfc.Workspaces.Read(ctx, organization, workspace)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("getting workspace %s/%s", organization, workspace))
	}

	return &deployer{
		ctx: ctx,
		log: log,
		tfe: tfc,
		wsp: wsp,
	}, nil
}

func (d *deployer) Deploy(varKey, varValue, gitCommit string) error {
	msg := fmt.Sprintf("triggered from commit %s", gitCommit)
	out, err := exec.Command("git", "log", "--oneline", "-n 1").Output()
	if err != nil {
		d.log.Error("getting git commit message", zap.Error(err))
	} else {
		msg = string(out)
	}

	err = d.updateVar(varKey, varValue)
	if err != nil {
		return errors.Wrap(err, "updating variable")
	}

	run, err := d.tfe.Runs.Create(d.ctx, tfe.RunCreateOptions{
		Message:   &msg,
		Workspace: d.wsp,
		AutoApply: boolPtr(true),
	})
	if err != nil {
		return errors.Wrap(err, "creating run")
	}

	err = d.runWait(run.ID)
	if err != nil {
		return errors.Wrap(err, "waiting on run")
	}

	return nil
}

func (d *deployer) updateVar(name string, value string) error {
	vars, err := d.tfe.Variables.List(d.ctx, d.wsp.ID, nil)
	if err != nil {
		return errors.Wrap(err, "listing workspace vars")
	}

	for _, v := range vars.Items {
		if v.Key == name {
			_, err := d.tfe.Variables.Update(d.ctx, d.wsp.ID, v.ID, tfe.VariableUpdateOptions{Value: &value})
			if err != nil {
				return errors.Wrap(err, fmt.Sprintf("updating variable %s", name))
			}
			return nil
		}
	}
	return fmt.Errorf("variable %s not found", name)
}

func (d *deployer) runWait(runID string) error {
	started := time.Now()
	for {
		run, err := d.tfe.Runs.Read(d.ctx, runID)
		if err != nil {
			return errors.Wrap(err, "reading run")
		}

		switch run.Status {
		case tfe.RunApplied:
			d.log.Info("success", zap.String("status", string(run.Status)))
			return nil
		case tfe.RunErrored:
			d.log.Info("failed", zap.String("status", string(run.Status)))
			return fmt.Errorf("run failed with status %q", run.Status)
		case tfe.RunDiscarded:
			d.log.Info("canceled", zap.String("status", string(run.Status)))
			return fmt.Errorf("run canceled with status %q", run.Status)
		case tfe.RunCanceled:
			d.log.Info("canceled", zap.String("status", string(run.Status)))
			return fmt.Errorf("run canceled with status %q", run.Status)
		default:
			d.log.Info("waiting", zap.String("status", string(run.Status)))
		}

		if time.Since(started) > runWaitTimeout {
			return fmt.Errorf("timeout waiting for run")
		}

		time.Sleep(runWaitDelay)
	}
}

func boolPtr(v bool) *bool {
	return &v
}
