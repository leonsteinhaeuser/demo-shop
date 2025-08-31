package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

// StopSignalHandler listens for termination signals and executes the provided functions.
// The processFunc is executed in a separate goroutine.
// The main goroutine waits for a signal and then calls the onShutdownFunc.
func StopSignalHandler(processFunc, onShutdownFunc func(ctx context.Context)) {
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go processFunc(context.Background())
	<-done
	onShutdownFunc(context.Background())
}
