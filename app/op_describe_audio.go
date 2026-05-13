package app

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"gorm.io/gorm"
)

type describeAudioArgs struct {
	ID  *astral.ObjectID `query:"key:id"`
	Out string           `query:"optional"`
}

func (ops *AudioOps) Describe(_ *astral.Context, q *routing.IncomingQuery, args describeAudioArgs) error {
	ch := q.Accept(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	row, err := ops.db.FindAudio(args.ID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ch.Send(astral.Err(objects.ErrNotFound))
		}
		return ch.Send(astral.Err(err))
	}

	if err := ch.Send(&objects.Descriptor{
		SourceID: q.Target(),
		ObjectID: args.ID,
		Data:     row.ToAudioFile(),
	}); err != nil {
		return err
	}

	return ch.Send(&astral.EOS{})
}
