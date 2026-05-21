package app

import (
	"errors"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type describeAudioArgs struct {
	ID  *astral.ObjectID `query:"key:id"`
	Out string           `query:"optional"`
}

func (ops *AudioOps) Describe(ctx *astral.Context, q *routing.IncomingQuery, args describeAudioArgs) error {
	ch := q.Accept(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	descriptors, err := ops.describer.DescribeObject(ctx, args.ID)
	if err != nil {
		if errors.Is(err, objects.ErrNotFound) {
			return ch.Send(astral.Err(objects.ErrNotFound))
		}
		return ch.Send(astral.Err(err))
	}

	for descriptor := range descriptors {
		descriptor.SourceID = q.Target()
		if err := ch.Send(descriptor); err != nil {
			return err
		}
	}

	return ch.Send(&astral.EOS{})
}
