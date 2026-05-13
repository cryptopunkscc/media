package app

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type searchAudioArgs struct {
	Query string `query:"key:q"`
	Out   string `query:"optional"`
}

func (ops *AudioOps) Search(ctx *astral.Context, q *routing.IncomingQuery, args searchAudioArgs) error {
	ch := q.Accept(channel.WithOutputFormat(args.Out))
	defer ch.Close()

	var searchQuery objects.SearchQuery
	if err := searchQuery.UnmarshalText([]byte(args.Query)); err != nil {
		return ch.Send(astral.Err(err))
	}

	results, err := ops.searcher.SearchObject(ctx, searchQuery)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	for result := range results {
		result.SourceID = q.Target()
		if err := ch.Send(result); err != nil {
			return err
		}
	}

	return ch.Send(&astral.EOS{})
}
