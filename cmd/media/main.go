package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/cryptopunkscc/astrald/astral"
	_ "github.com/cryptopunkscc/astrald/astral/log/views"
	"github.com/cryptopunkscc/astrald/lib/apps"
	libastrald "github.com/cryptopunkscc/astrald/lib/astrald"
	_ "github.com/cryptopunkscc/media/all"
	"github.com/cryptopunkscc/media/app"
	"github.com/cryptopunkscc/media/db"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	configureLogger()

	id := libastrald.GuestID()
	logger := newLogger(id)

	db, err := db.Open()
	if err != nil {
		logger.Error("open media database: %v", err)
		os.Exit(1)
	}

	err = db.AutoMigrate()
	if err != nil {
		logger.Error("migrate media database: %v", err)
		os.Exit(1)
	}

	logger.Log("starting media app")
	astralCtx := astral.
		NewContext(ctx).
		WithIdentity(id).
		WithZone(astral.ZoneAll)

	router := app.NewRouter(db, logger)

	indexer := app.NewIndexer(db, logger)
	var (
		wg       sync.WaitGroup
		once     sync.Once
		firstErr error
	)

	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := indexer.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
			once.Do(func() {
				firstErr = err
				stop()
			})
		}
	}()

	go func() {
		defer wg.Done()
		if err := apps.Serve(astralCtx, router); err != nil && !errors.Is(err, context.Canceled) {
			once.Do(func() {
				firstErr = err
				stop()
			})
		}
	}()

	wg.Wait()

	if firstErr != nil {
		logger.Error("media app stopped: %v", firstErr)
		os.Exit(1)
	}
}
