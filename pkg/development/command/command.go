package command

import (
	"context"

	"k8s.io/klog/v2"
	"k8s.io/utils/exec"
)

type baseCommand struct {
	executor exec.Interface
	dir      string
}

func newBaseCommand() *baseCommand {
	return &baseCommand{executor: exec.New()}
}

func (c *baseCommand) run(ctx context.Context, cmdStr string, args ...string) (string, error) {
	cmd := c.executor.CommandContext(ctx, cmdStr, args...)
	if c.dir != "" {
		cmd.SetDir(c.dir)
	}

	output, err := cmd.CombinedOutput()
	klog.Info("command output: \n", string(output))
	if err != nil {
		klog.Error("run command error, ", err, ", ", cmdStr)
	}
	return string(output), err
}

func (c *baseCommand) withDir(dir string) *baseCommand {
	c.dir = dir
	return c
}

type checkChart struct {
	baseCommand
}

func newCheckChart() *checkChart {
	return &checkChart{
		*newBaseCommand(),
	}
}
