/*
All Rights Reversed (ɔ)
*/

package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/redpierrot/bestow/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	cmd.Execute(ctx)
}
