package app

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/media/db"
)

var _ objects.Searcher = (*audioSearcher)(nil)

type audioSearcher struct {
	db *db.DB
}

func newAudioSearcher(db *db.DB) *audioSearcher {
	return &audioSearcher{db: db}
}

func (s *audioSearcher) SearchObject(ctx *astral.Context, query objects.SearchQuery) (<-chan *objects.SearchResult, error) {
	if err := query.RequiredTagsIn(knownAudioTags...); err != nil {
		return nil, err
	}

	rows, err := s.db.SearchAudio(parseAudioSearch(query))
	if err != nil {
		return nil, err
	}

	results := make(chan *objects.SearchResult)
	go func() {
		defer close(results)

		for _, row := range rows {
			select {
			case results <- &objects.SearchResult{
				SourceID: ctx.Identity(),
				ObjectID: row.ObjectID,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()

	return results, nil
}

var knownAudioTags = []string{"artist", "album", "title", "genre", "year", "format"}

var knownAudioTagSet = func() map[string]bool {
	m := make(map[string]bool, len(knownAudioTags))
	for _, t := range knownAudioTags {
		m[t] = true
	}
	return m
}()

func parseAudioSearch(query objects.SearchQuery) db.AudioSearch {
	search := db.AudioSearch{
		Text:    strings.ToLower(strings.TrimSpace(string(query.Query))),
		Include: make(map[string][]string),
		Exclude: make(map[string][]string),
	}

	for _, tag := range query.Tags {
		name := string(tag.Name)
		if !knownAudioTagSet[name] {
			continue
		}

		value := string(tag.Value)
		switch tag.Mod {
		case objects.TagModRequire:
			search.Include[name] = append(search.Include[name], value)
		case objects.TagModExclude:
			search.Exclude[name] = append(search.Exclude[name], value)
		}
	}

	return search
}
