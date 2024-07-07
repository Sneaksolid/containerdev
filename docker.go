package main

import (
	"fmt"
	"os"
	"os/exec"

	"context"
)

type RunOptions struct {
	Name   string
	Image  string
	Stdin  bool
	Tty    bool
	AsUser bool

	Volumes map[string]string
	WorkDir string

	EntryPoint string
	Cmd        []string
}

func (o *RunOptions) args() []string {
	opts := []string{
		"run",
		"--rm",
	}

	if o.AsUser {
		uid := os.Geteuid()
		gid := os.Getegid()

		opts = append(opts,
			"-v", "/etc/passwd:/etc/passwd:ro",
			"-v", "/etc/group:/etc/group:ro",
			"-u", fmt.Sprintf("%d:%d", uid, gid))
	}

	if o.Name != "" {
		opts = append(opts, "--name", o.Name)
	}

	if o.Stdin {
		opts = append(opts, "-i")
	}

	if o.Tty {
		opts = append(opts, "-t")
	}

	if o.EntryPoint != "" {
		opts = append(opts, "--entrypoint", o.EntryPoint)
	}

	for src, dst := range o.Volumes {
		opts = append(opts, "-v", fmt.Sprintf("%s:%s", src, dst))
	}

	if o.WorkDir != "" {
		opts = append(opts, "-w", o.WorkDir)
	}

	opts = append(opts, o.Image)
	if len(o.Cmd) > 0 {
		opts = append(opts, o.Cmd...)
	}

	return opts
}

func Run(ctx context.Context, opts RunOptions) error {
	cmd := exec.CommandContext(ctx, "docker", opts.args()...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		return err
	}

	return nil
}
