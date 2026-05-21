package app

import (
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/media/db"
)

type Root struct{}

type AudioOps struct {
	db        *db.DB
	logger    *log.Logger
	searcher  objects.Searcher
	describer objects.Describer
}

func NewRouter(db *db.DB, logger *log.Logger, searcher objects.Searcher, describer objects.Describer) *routing.App {
	app := routing.NewApp(&Root{})
	app.Add("audio", &AudioOps{
		db:        db,
		logger:    logger.AppendTag(log.Tag("audio")),
		searcher:  searcher,
		describer: describer,
	})
	return app
}

func NewObjectSearcher(db *db.DB) objects.Searcher {
	return newAudioSearcher(db)
}

func NewObjectDescriber(db *db.DB) objects.Describer {
	return newAudioDescriber(db)
}
