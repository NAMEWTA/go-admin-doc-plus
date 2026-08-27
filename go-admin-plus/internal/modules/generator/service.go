package generator

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"sync"
	"time"
)

type Generator struct {
	metadata   MetadataSource
	writer     *AtomicWriter
	authorizer Authorizer
	store      ConfigStore
	renderer   OutputRenderer
	ttl        time.Duration
	now        func() time.Time
	mu         sync.Mutex
	previews   map[string]pendingPreview
	order      []string
}

type Authorizer interface {
	Require(context.Context, string, string) error
}

type pendingPreview struct {
	actorID string
	preview Preview
	bytes   int
}

func New(metadata MetadataSource, writer *AtomicWriter, authorizer Authorizer, store ConfigStore, renderer OutputRenderer, ttl time.Duration) (*Generator, error) {
	if metadata == nil || writer == nil || authorizer == nil || store == nil || renderer == nil || ttl < time.Minute || ttl > 30*time.Minute {
		return nil, ErrInvalid
	}
	return &Generator{metadata: metadata, writer: writer, authorizer: authorizer, store: store, renderer: renderer, ttl: ttl, now: time.Now, previews: make(map[string]pendingPreview)}, nil
}

func (generator *Generator) Tables(ctx context.Context, actorID string) ([]TableRef, error) {
	if err := generator.authorize(ctx, actorID, PermissionMetadataRead); err != nil {
		return nil, err
	}
	return generator.metadata.Tables(ctx)
}

func (generator *Generator) Describe(ctx context.Context, actorID string, ref TableRef) (Table, error) {
	if err := generator.authorize(ctx, actorID, PermissionMetadataRead); err != nil {
		return Table{}, err
	}
	return generator.metadata.Describe(ctx, ref)
}

func (generator *Generator) Config(ctx context.Context, actorID, module string) (Draft, string, error) {
	if err := generator.authorize(ctx, actorID, PermissionPreview); err != nil {
		return Draft{}, "", err
	}
	return generator.store.Get(ctx, actorID, module)
}

func (generator *Generator) Preview(ctx context.Context, actorID string, draft Draft) (Preview, error) {
	if err := generator.authorize(ctx, actorID, PermissionPreview); err != nil {
		return Preview{}, err
	}
	table, err := generator.metadata.Describe(ctx, draft.Table)
	if err != nil {
		return Preview{}, err
	}
	model, err := normalize(table, draft)
	if err != nil {
		return Preview{}, err
	}
	files, err := generator.renderer.Render(ctx, model)
	if err != nil {
		return Preview{}, err
	}
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	if len(files) > 64 {
		return Preview{}, ErrInvalid
	}
	digest := sha256.New()
	totalBytes := 0
	for index := range files {
		if len(files[index].Content) > 1_000_000 {
			return Preview{}, ErrInvalid
		}
		totalBytes += len(files[index].Path) + len(files[index].Content)
		if totalBytes > 8_000_000 {
			return Preview{}, ErrInvalid
		}
		contentHash := sha256.Sum256([]byte(files[index].Content))
		files[index].SHA256 = hex.EncodeToString(contentHash[:])
		digest.Write([]byte(files[index].Path))
		digest.Write([]byte{0})
		digest.Write(contentHash[:])
	}
	now := generator.now().UTC()
	token, err := randomToken()
	if err != nil {
		return Preview{}, ErrInternal
	}
	preview := Preview{Token: token, Digest: hex.EncodeToString(digest.Sum(nil)), Module: model.Module, CreatedAt: now, ExpiresAt: now.Add(generator.ttl), Files: files}
	if err := generator.store.Save(ctx, actorID, draft, preview); err != nil {
		return Preview{}, err
	}
	generator.mu.Lock()
	defer generator.mu.Unlock()
	generator.order = append(generator.order, preview.Token)
	generator.previews[preview.Token] = pendingPreview{actorID: actorID, preview: clonePreview(preview), bytes: totalBytes}
	for len(generator.order) > 16 || generator.pendingBytes() > 32_000_000 {
		delete(generator.previews, generator.order[0])
		generator.order = generator.order[1:]
	}
	return clonePreview(preview), nil
}

func (generator *Generator) pendingBytes() int {
	total := 0
	for _, value := range generator.previews {
		total += value.bytes
	}
	return total
}

func (generator *Generator) Write(ctx context.Context, actorID, previewToken string, confirmed bool) (WriteResult, error) {
	if !confirmed || len(previewToken) != sha256.Size*2 {
		return WriteResult{}, ErrInvalid
	}
	if err := generator.authorize(ctx, actorID, PermissionWrite); err != nil {
		return WriteResult{}, err
	}
	generator.mu.Lock()
	pending, exists := generator.previews[previewToken]
	if exists && pending.actorID == actorID {
		delete(generator.previews, previewToken)
	}
	generator.mu.Unlock()
	if !exists || pending.actorID != actorID || !generator.now().UTC().Before(pending.preview.ExpiresAt) {
		return WriteResult{}, ErrPreviewStale
	}
	result, err := generator.writer.Write(ctx, clonePreview(pending.preview), func(ctx context.Context) error {
		return generator.authorize(ctx, actorID, PermissionWrite)
	})
	if err != nil {
		return WriteResult{}, err
	}
	return result, nil
}

func (generator *Generator) authorize(ctx context.Context, actorID, permission string) error {
	if actorID == "" {
		return ErrDenied
	}
	if err := generator.authorizer.Require(ctx, actorID, permission); err != nil {
		if context.Cause(ctx) != nil {
			return context.Cause(ctx)
		}
		return ErrDenied
	}
	return nil
}

func randomToken() (string, error) {
	value := make([]byte, sha256.Size)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}

func clonePreview(preview Preview) Preview {
	preview.Files = append([]PreviewFile(nil), preview.Files...)
	return preview
}
