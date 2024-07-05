package main

import (
	"context"
	"os"
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
		panic(err)
	}
}
