package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	libastrald "github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/indexing"
	indexingClient "github.com/cryptopunkscc/astrald/mod/indexing/client"
	"github.com/cryptopunkscc/astrald/mod/objects"
	objectsClient "github.com/cryptopunkscc/astrald/mod/objects/client"
	"github.com/cryptopunkscc/media/db"
	"github.com/cryptopunkscc/media/obj"
	"github.com/dhowden/tag"
	"gorm.io/gorm"
)

const indexerName = "media-indexer"

type Indexer struct {
	db      *db.DB
	objects *objectsClient.Client
	reader  objects.Repository
	logger  *log.Logger
}

type nextResult struct {
	obj astral.Object
	err error
}

func NewIndexer(db *db.DB, logger *log.Logger) *Indexer {
	client := objectsClient.Default()
	return &Indexer{
		db:      db,
		objects: client,
		reader:  &clientRepository{client: client},
		logger:  logger.AppendTag(log.Tag("indexer")),
	}
}

func (idx *Indexer) Run(ctx context.Context) error {
	backoff := time.Second

	for {
		idx.logger.Logv(1, "connecting to astrald")

		connected, err := idx.runOnce(ctx)
		if err == nil || errors.Is(err, context.Canceled) {
			return err
		}

		if connected {
			backoff = time.Second
		}
		idx.logger.Logv(1, "indexer disconnected: %v", err)
		idx.logger.Logv(1, "retrying in %v", backoff)

		if !sleepContext(ctx.Done(), backoff) {
			return context.Canceled
		}

		if backoff < time.Minute {
			backoff *= 2
			if backoff > time.Minute {
				backoff = time.Minute
			}
		}
	}
}

func (idx *Indexer) runOnce(ctx context.Context) (bool, error) {
	astralCtx := astral.
		NewContext(ctx).
		WithIdentity(libastrald.GuestID()).
		WithZone(astral.ZoneAll)

	nonce, err := indexingClient.RegisterIndexer(astralCtx, indexerName)
	if err != nil {
		return false, err
	}
	idx.logger.Logv(1, "registered indexer %v with nonce %v", indexerName, nonce)

	sub, err := indexingClient.Subscribe(astralCtx, nonce)
	if err != nil {
		return false, err
	}

	defer sub.Close()
	idx.logger.Logv(1, "subscribed to indexing changes")

	for {
		object, err := nextChange(ctx, sub, idx.logger)
		if err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, context.Canceled) {
				return true, err
			}
			return true, err
		}

		shouldFail, err := idx.handleChange(astralCtx, object)
		if err != nil {
			idx.logger.Error("change failed: %v", err)
		}

		if shouldFail {
			if err := sub.Fail(); err != nil {
				return true, err
			}
			continue
		}

		if err := sub.Ack(); err != nil {
			return true, err
		}
	}
}

func nextChange(ctx context.Context, sub *indexingClient.Subscription, logger *log.Logger) (astral.Object, error) {
	resCh := make(chan nextResult, 1)

	go func() {
		object, err := sub.Next()
		resCh <- nextResult{obj: object, err: err}
	}()

	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case res := <-resCh:
			return res.obj, res.err
		case <-ticker.C:
			logger.Logv(1, "waiting for indexing changes")
		case <-ctx.Done():
			_ = sub.Close()
			return nil, ctx.Err()
		}
	}
}

func (idx *Indexer) handleChange(ctx *astral.Context, object astral.Object) (bool, error) {
	switch msg := object.(type) {
	case *indexing.IndexMsg:
		return idx.handleIndex(ctx, msg)
	case *indexing.UnindexMsg:
		return false, nil
	default:
		return true, astral.NewErrUnexpectedObject(object)
	}
}

func (idx *Indexer) handleIndex(ctx *astral.Context, msg *indexing.IndexMsg) (bool, error) {
	_, err := idx.db.FindObject(msg.ObjectID)
	switch {
	case err == nil:
		return false, nil
	case !errors.Is(err, gorm.ErrRecordNotFound):
		return true, err
	}

	probe, err := idx.objects.Probe(ctx, msg.ObjectID, objects.RepoLocal)
	if err != nil {
		return true, err
	}
	if probe != nil && probe.Type != "" {
		idx.logger.Logv(1, "skipping typed object %v (%v)", msg.ObjectID, probe.Type)
		return false, nil
	}

	audio, err := idx.inspect(ctx, msg.ObjectID)
	if err != nil {
		if errors.Is(err, ErrNotMedia) {
			idx.logger.Logv(1, "skipping non-media object %v", msg.ObjectID)
			return false, nil
		}
		return true, err
	}

	if err := idx.db.SaveAudio(audio); err != nil {
		return true, err
	}

	if err := idx.db.SaveObject(msg.ObjectID); err != nil {
		return true, err
	}

	idx.logger.Logv(1, "indexed %v", audio)
	return false, nil
}

func (idx *Indexer) inspect(ctx *astral.Context, objectID *astral.ObjectID) (*obj.AudioFile, error) {
	reader, err := idx.reader.Read(ctx, objectID, 0, 0)
	if err != nil {
		return nil, err
	}

	rs := objects.NewReadSeeker(ctx, objectID, idx.reader, reader)
	defer rs.Close()

	audioTag, err := tag.ReadFrom(rs)
	if err != nil {
		return nil, ErrNotMedia
	}

	var pictureID *astral.ObjectID
	if p := audioTag.Picture(); p != nil {
		pictureID, err = astral.Resolve(bytes.NewReader(p.Data))
		if err != nil {
			idx.logger.Logv(1, "ignoring invalid embedded picture for %v: %v", objectID, err)
			pictureID = nil
		}
	}

	return &obj.AudioFile{
		ObjectID:  objectID,
		Format:    astral.String8(audioTag.FileType()),
		Title:     astral.String8(audioTag.Title()),
		Artist:    astral.String8(audioTag.Artist()),
		Album:     astral.String8(audioTag.Album()),
		Genre:     astral.String8(audioTag.Genre()),
		Year:      astral.Uint16(audioTag.Year()),
		PictureID: pictureID,
	}, nil
}

type clientRepository struct {
	objects.NilRepository
	client *objectsClient.Client
}

func (repo *clientRepository) Label() string {
	return "Astrald client"
}

func (repo *clientRepository) Read(ctx *astral.Context, objectID *astral.ObjectID, offset int64, limit int64) (objects.Reader, error) {
	reader, err := repo.client.Read(ctx, objectID, offset, limit)
	if err != nil {
		return nil, err
	}
	return &clientReader{ReadCloser: reader, repo: repo}, nil
}

type clientReader struct {
	io.ReadCloser
	repo objects.Repository
}

func (r *clientReader) Repo() objects.Repository {
	return r.repo
}

func sleepContext(done <-chan struct{}, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return true
	case <-done:
		return false
	}
}
