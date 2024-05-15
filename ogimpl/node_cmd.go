package ogimpl

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"github.com/symphony09/ograph"
	"github.com/symphony09/ograph/ogcore"
)

var CmdNodeFactory = func() ogcore.Node {
	return &CmdNode{}
}

type CmdNode struct {
	ograph.BaseNode
	*slog.Logger

	Cmd  []string
	Env  []string
	Dir  string
	Path string
}

func (node *CmdNode) Run(ctx context.Context, state ogcore.State) error {
	if node.Logger == nil {
		node.Logger = slog.Default()
	}

	if len(node.Cmd) == 0 {
		return errors.New("cmd is empty")
	}

	if !node.CmdIsAllowed() {
		node.Logger.Error("exec cmd is not allowed", "Cmd", node.Cmd)
		return errors.New("cmd is not allowed")
	}

	var args []string
	if len(node.Cmd) > 1 {
		args = node.Cmd[1:]
	}

	cmd := exec.Command(node.Cmd[0], args...)
	cmd.Env = node.Env
	cmd.Dir = node.Dir

	if node.Path != "" {
		cmd.Path = node.Path
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		node.Logger.Error("exec cmd failed", "Cmd", strings.Join(node.Cmd, " "), "Output", string(output))
		return fmt.Errorf("exec cmd failed, err: %w", err)
	}

	node.Logger.Info("exec cmd succeed", "Cmd", strings.Join(node.Cmd, " "), "Output", string(output))

	return nil
}

func (node *CmdNode) CmdIsAllowed() bool {
	if len(node.Cmd) == 0 {
		return false
	}

	allowCmdList := os.Getenv("OGRAPH_ALLOW_CMD_LIST")

	if allowCmdList == "" {
		return true
	}

	for _, allowCmd := range strings.Split(allowCmdList, ",") {
		allowCmd = strings.TrimSpace(allowCmd)

		if node.Cmd[0] == allowCmd {
			return true
		}
	}

	return false
}
