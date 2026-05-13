package app

import (
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/media/db"
)

type Root struct{}

type AudioOps struct {
	db       *db.DB
	logger   *log.Logger
	searcher *audioSearcher
}

func NewRouter(db *db.DB, logger *log.Logger) *routing.App {
	app := routing.NewApp(&Root{})
	app.Add("audio", &AudioOps{
		db:       db,
		logger:   logger.AppendTag(log.Tag("audio")),
		searcher: newAudioSearcher(db),
	})
	return app
}
