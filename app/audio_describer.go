package app

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/media/db"
	"gorm.io/gorm"
)

var _ objects.Describer = (*audioDescriber)(nil)

type audioDescriber struct {
	db *db.DB
}

func newAudioDescriber(db *db.DB) *audioDescriber {
	return &audioDescriber{db: db}
}

func (d *audioDescriber) DescribeObject(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.Descriptor, error) {
	if objectID == nil || objectID.IsZero() {
		return nil, objects.ErrNotFound
	}

	row, err := d.db.FindAudio(objectID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, objects.ErrNotFound
		}
		return nil, err
	}

	descriptor := &objects.Descriptor{
		SourceID: ctx.Identity(),
		ObjectID: objectID,
		Data:     row.ToAudioFile(),
	}

	out := make(chan *objects.Descriptor, 1)
	go func() {
		defer close(out)

		select {
		case out <- descriptor:
		case <-ctx.Done():
		}
	}()

	return out, nil
}
