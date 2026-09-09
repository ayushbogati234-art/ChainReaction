package main

import (
	"context"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/app"
	"github.com/ayushbogati234-art/ChainReaction/backend/internal/config"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	if err := app.RunAPI(ctx, config.Load()); err != nil {
		log.Fatal(err)
	}
}
