package main

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	logviews "github.com/cryptopunkscc/astrald/mod/log/views"
)

const appLogTag = "media"

func configureLogger() {
	logviews.UseEntryView()
	logviews.UseIdentityView()
}

func newLogger(id *astral.Identity) *log.Logger {
	return log.New(id).Tag(log.Tag(appLogTag))
}
