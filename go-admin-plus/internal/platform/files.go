package platform

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrInvalidFileKey = errors.New("invalid file key")
	ErrFileNotFound   = errors.New("file not found")
)

// FileStore owns files addressed by relative, slash-separated keys. Keys must
// never escape the configured root, including through symbolic links.
type FileStore interface {
	Put(ctx context.Context, key string, source io.Reader) error
	Open(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, key string) error
}

// LocalFileStore uses an os.Root handle so concurrent path replacement cannot
// redirect an operation outside the configured data root.
type LocalFileStore struct {
	root *os.Root
}

// NewLocalFileStore creates or opens root. Close releases the root handle.
func NewLocalFileStore(root string) (*LocalFileStore, error) {
	root = strings.TrimSpace(root)
	if root == "" {
		return nil, errors.New("file store root is required")
	}
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("create file store root: %w", err)
	}
	rootHandle, err := os.OpenRoot(root)
	if err != nil {
		return nil, fmt.Errorf("open file store root: %w", err)
	}
	return &LocalFileStore{root: rootHandle}, nil
}

func (s *LocalFileStore) Put(ctx context.Context, key string, source io.Reader) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	if source == nil {
		return errors.New("file source is required")
	}
	relative, err := normalizeFileKey(key)
	if err != nil {
		return err
	}
	parent := filepath.Dir(relative)
	if err := s.root.MkdirAll(parent, 0o700); err != nil {
		return fileParentError(err)
	}
	temporaryName, temporary, err := s.createTemporary(parent)
	if err != nil {
		return err
	}
	defer s.root.Remove(temporaryName)
	_, copyErr := io.Copy(temporary, &contextReader{ctx: ctx, source: source})
	closeErr := temporary.Close()
	if copyErr != nil || closeErr != nil {
		return errors.Join(
			wrapFileError("write file", copyErr),
			wrapFileError("close temporary file", closeErr),
		)
	}
	if err := s.root.Rename(temporaryName, relative); err != nil {
		if _, statErr := s.root.Stat(relative); statErr != nil {
			return fileOperationError("replace file", err)
		}
		recoveryName, nameErr := randomFileName(parent, ".go-admin-recover-")
		if nameErr != nil {
			return fileOperationError("replace file", errors.Join(err, nameErr))
		}
		if preserveErr := s.root.Rename(relative, recoveryName); preserveErr != nil {
			return fileOperationError("preserve replaced file", errors.Join(err, preserveErr))
		}
		if retryErr := s.root.Rename(temporaryName, relative); retryErr != nil {
			restoreErr := s.root.Rename(recoveryName, relative)
			return fileOperationError("replace file", errors.Join(retryErr, restoreErr))
		}
		if removeErr := s.root.Remove(recoveryName); removeErr != nil {
			return fileOperationError("remove replaced file recovery copy", removeErr)
		}
	}
	return nil
}

func (s *LocalFileStore) Open(ctx context.Context, key string) (io.ReadCloser, error) {
	if err := contextError(ctx); err != nil {
		return nil, err
	}
	relative, err := normalizeFileKey(key)
	if err != nil {
		return nil, err
	}
	file, err := s.root.Open(relative)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrFileNotFound
	}
	if err != nil {
		return nil, fileOperationError("open file", err)
	}
	return file, nil
}

func (s *LocalFileStore) Delete(ctx context.Context, key string) error {
	if err := contextError(ctx); err != nil {
		return err
	}
	relative, err := normalizeFileKey(key)
	if err != nil {
		return err
	}
	if err := s.root.Remove(relative); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fileOperationError("delete file", err)
	}
	return nil
}

// Close releases the directory handle. Existing opened readers remain owned
// by their callers and must be closed separately.
func (s *LocalFileStore) Close() error {
	return s.root.Close()
}

func (s *LocalFileStore) createTemporary(parent string) (string, *os.File, error) {
	for attempt := 0; attempt < 10; attempt++ {
		name, err := randomFileName(parent, ".go-admin-write-")
		if err != nil {
			return "", nil, err
		}
		file, err := s.root.OpenFile(name, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
		if errors.Is(err, os.ErrExist) {
			continue
		}
		if err != nil {
			return "", nil, fileOperationError("create temporary file", err)
		}
		return name, file, nil
	}
	return "", nil, errors.New("create temporary file: unique name attempts exhausted")
}

func randomFileName(parent, prefix string) (string, error) {
	random := make([]byte, 16)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("create temporary file name: %w", err)
	}
	return filepath.Join(parent, prefix+hex.EncodeToString(random)), nil
}

func normalizeFileKey(key string) (string, error) {
	key = strings.TrimSpace(key)
	if key == "" || strings.ContainsRune(key, '\x00') || filepath.IsAbs(key) || filepath.VolumeName(key) != "" {
		return "", ErrInvalidFileKey
	}
	key = filepath.FromSlash(key)
	cleaned := filepath.Clean(key)
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", ErrInvalidFileKey
	}
	return cleaned, nil
}

func fileOperationError(operation string, err error) error {
	if errors.Is(err, os.ErrInvalid) || strings.Contains(err.Error(), "escapes") {
		return fmt.Errorf("%s: %w", operation, ErrInvalidFileKey)
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func fileParentError(err error) error {
	// Linux reports EEXIST when MkdirAll reaches a symlink or non-directory
	// component; other platforms report an explicit root escape instead.
	if errors.Is(err, os.ErrExist) {
		return fmt.Errorf("create file parent: %w", ErrInvalidFileKey)
	}
	return fileOperationError("create file parent", err)
}

func wrapFileError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fileOperationError(operation, err)
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	return ctx.Err()
}

type contextReader struct {
	ctx    context.Context
	source io.Reader
}

func (r *contextReader) Read(buffer []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	return r.source.Read(buffer)
}
