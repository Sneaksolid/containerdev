package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"os/signal"
)

func run(f func(ctx context.Context) error) {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		<-sig
		cancel()
	}()

	if err := f(ctx); err != nil {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) {
			os.Exit(exitError.ExitCode())
			return
		}

		panic(err)
	}
}
