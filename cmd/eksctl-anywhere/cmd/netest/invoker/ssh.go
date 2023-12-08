package invoker

import (
	"context"
	"os/exec"
	"strings"

	"github.com/go-logr/logr"
)

// SSH is an Invoker that executes commands over an SSH channel.
type SSH struct {
	Host string
	Log  logr.Logger
}

// Invoke staisfies Invoker.
func (s SSH) Invoke(ctx context.Context, args ...string) Outcome {
	s.Log.Info("Running", "args", args)
	cmd := exec.CommandContext(ctx, "ssh", append([]string{s.Host}, args...)...)

	res := Outcome{Cmd: strings.Join(cmd.Args, " ")}
	cmd.Stdout = &res.Stdout
	cmd.Stderr = &res.Stderr

	if err := cmd.Run(); err != nil {
		res.Error = err
	}

	return res
}
