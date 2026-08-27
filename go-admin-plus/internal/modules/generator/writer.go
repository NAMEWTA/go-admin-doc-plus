package generator

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"go/format"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
)

var outputCommitLock sync.Mutex

type OutputGate interface {
	Check(context.Context, string, Preview) error
}

// CompleteOutputGate proves contract generation, Go/TypeScript compilation, tests and package boundaries.
// Implementations are application-owned and cannot be selected or parameterized by an HTTP request.
type CompleteOutputGate interface {
	OutputGate
	CompleteOutputGate()
}

type AtomicWriter struct {
	root  string
	gates []OutputGate
}

func NewAtomicWriter(root string, gates ...OutputGate) (*AtomicWriter, error) {
	canonical, err := canonicalDirectory(root)
	if err != nil {
		return nil, ErrInvalid
	}
	complete := false
	for _, gate := range gates {
		if _, ok := gate.(CompleteOutputGate); ok {
			complete = true
		}
	}
	if !complete {
		return nil, ErrInvalid
	}
	return &AtomicWriter{root: canonical, gates: append([]OutputGate{SyntaxBoundaryGate{}}, gates...)}, nil
}

func (writer *AtomicWriter) Write(ctx context.Context, preview Preview, beforeCommit ...func(context.Context) error) (result WriteResult, returnedError error) {
	outputCommitLock.Lock()
	defer outputCommitLock.Unlock()
	if len(beforeCommit) > 1 || (len(beforeCommit) == 1 && beforeCommit[0] == nil) {
		return WriteResult{}, ErrInvalid
	}
	root, err := canonicalDirectory(writer.root)
	if err != nil || root != writer.root || !modulePattern.MatchString(preview.Module) || len(preview.Token) != sha256.Size*2 {
		return WriteResult{}, ErrInvalid
	}
	if err := verifyPreview(preview); err != nil {
		return WriteResult{}, err
	}
	rootHandle, err := os.OpenRoot(root)
	if err != nil {
		return WriteResult{}, ErrInternal
	}
	defer rootHandle.Close()
	if !rootStillAnchored(rootHandle, writer.root) {
		return WriteResult{}, ErrInvalid
	}
	targetName := preview.Module + "-" + preview.Digest[:12]
	if _, err := rootHandle.Lstat(targetName); err == nil {
		return WriteResult{}, ErrConflict
	} else if !errors.Is(err, fs.ErrNotExist) {
		return WriteResult{}, ErrInternal
	}
	stagingName, stagingHandle, err := createAnchoredStaging(rootHandle)
	if err != nil {
		return WriteResult{}, ErrInternal
	}
	defer stagingHandle.Close()
	defer func() { _ = rootHandle.RemoveAll(stagingName) }()
	staging := filepath.Join(root, stagingName)
	for _, file := range preview.Files {
		_, pathErr := safeJoin(staging, file.Path)
		if pathErr != nil {
			return WriteResult{}, pathErr
		}
		relative := filepath.FromSlash(file.Path)
		if err := stagingHandle.MkdirAll(filepath.Dir(relative), 0o750); err != nil {
			return WriteResult{}, ErrInternal
		}
		handle, err := stagingHandle.OpenFile(relative, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
		if err != nil {
			return WriteResult{}, ErrInternal
		}
		_, writeErr := handle.WriteString(file.Content)
		closeErr := handle.Close()
		if writeErr != nil || closeErr != nil {
			return WriteResult{}, ErrInternal
		}
	}
	for _, gate := range writer.gates {
		if gate == nil {
			return WriteResult{}, ErrInvalid
		}
		if err := gate.Check(ctx, staging, preview); err != nil {
			if context.Cause(ctx) != nil {
				return WriteResult{}, context.Cause(ctx)
			}
			return WriteResult{}, ErrGateFailed
		}
	}
	if err := verifyAnchoredStaging(stagingHandle, preview); err != nil {
		return WriteResult{}, ErrGateFailed
	}
	if err := context.Cause(ctx); err != nil {
		return WriteResult{}, err
	}
	if len(beforeCommit) == 1 {
		if err := beforeCommit[0](ctx); err != nil {
			if context.Cause(ctx) != nil {
				return WriteResult{}, context.Cause(ctx)
			}
			return WriteResult{}, err
		}
	}
	if !rootStillAnchored(rootHandle, writer.root) {
		return WriteResult{}, ErrInvalid
	}
	if err := rootHandle.Rename(stagingName, targetName); err != nil {
		if _, statErr := rootHandle.Lstat(targetName); statErr == nil {
			return WriteResult{}, ErrConflict
		}
		return WriteResult{}, ErrInternal
	}
	if !rootStillAnchored(rootHandle, writer.root) {
		_ = rootHandle.RemoveAll(targetName)
		return WriteResult{}, ErrInvalid
	}
	paths := make([]string, 0, len(preview.Files))
	for _, file := range preview.Files {
		paths = append(paths, file.Path)
	}
	return WriteResult{Token: preview.Token, Directory: targetName, Files: paths}, nil
}

func createAnchoredStaging(root *os.Root) (string, *os.Root, error) {
	for range 8 {
		token, err := randomToken()
		if err != nil {
			return "", nil, err
		}
		name := ".generator-stage-" + token[:24]
		if err := root.Mkdir(name, 0o700); err != nil {
			if errors.Is(err, fs.ErrExist) {
				continue
			}
			return "", nil, err
		}
		handle, err := root.OpenRoot(name)
		if err != nil {
			_ = root.RemoveAll(name)
			return "", nil, err
		}
		return name, handle, nil
	}
	return "", nil, ErrInternal
}

func rootStillAnchored(root *os.Root, configuredPath string) bool {
	anchored, err := root.Stat(".")
	if err != nil {
		return false
	}
	current, err := os.Lstat(configuredPath)
	return err == nil && current.IsDir() && current.Mode()&os.ModeSymlink == 0 && os.SameFile(anchored, current)
}

type SyntaxBoundaryGate struct{}

func (SyntaxBoundaryGate) Check(_ context.Context, root string, preview Preview) error {
	for _, file := range preview.Files {
		path, err := safeJoin(root, file.Path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil || string(content) != file.Content {
			return fmt.Errorf("generated file drift")
		}
		switch filepath.Ext(file.Path) {
		case ".go":
			formatted, formatErr := format.Source(content)
			if formatErr != nil || string(formatted) != string(content) {
				return fmt.Errorf("Go source is not canonical")
			}
		case ".json":
			var value any
			if json.Unmarshal(content, &value) != nil {
				return fmt.Errorf("JSON is invalid")
			}
		}
		lower := strings.ToLower(file.Path + "\n" + file.Content)
		for _, forbidden := range []string{"gorm.io/", "casbin", "redis", "tenant_id", "/src/../", "gorm:\""} {
			if strings.Contains(lower, forbidden) {
				return fmt.Errorf("forbidden generated architecture pattern")
			}
		}
	}
	return nil
}

func canonicalDirectory(path string) (string, error) {
	if path == "" {
		return "", ErrInvalid
	}
	absolute, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	info, err := os.Lstat(absolute)
	if err != nil || !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
		return "", ErrInvalid
	}
	return filepath.EvalSymlinks(absolute)
}

func safeJoin(root, relativePath string) (string, error) {
	if relativePath == "" || filepath.IsAbs(relativePath) || strings.Contains(relativePath, `\`) || filepath.ToSlash(filepath.Clean(relativePath)) != relativePath {
		return "", ErrInvalid
	}
	for _, segment := range strings.Split(relativePath, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return "", ErrInvalid
		}
	}
	path := filepath.Join(root, filepath.FromSlash(relativePath))
	location, err := filepath.Rel(root, path)
	if err != nil || location == ".." || strings.HasPrefix(location, ".."+string(filepath.Separator)) {
		return "", ErrInvalid
	}
	return path, nil
}

func verifyPreview(preview Preview) error {
	files := append([]PreviewFile(nil), preview.Files...)
	sort.Slice(files, func(i, j int) bool { return files[i].Path < files[j].Path })
	digest := sha256.New()
	seen := map[string]struct{}{}
	for _, file := range files {
		if _, duplicate := seen[file.Path]; duplicate {
			return ErrInvalid
		}
		seen[file.Path] = struct{}{}
		if _, err := safeJoin("/generator-root", file.Path); err != nil {
			return err
		}
		hash := sha256.Sum256([]byte(file.Content))
		if file.SHA256 != hex.EncodeToString(hash[:]) {
			return ErrPreviewStale
		}
		digest.Write([]byte(file.Path))
		digest.Write([]byte{0})
		digest.Write(hash[:])
	}
	if preview.Digest != hex.EncodeToString(digest.Sum(nil)) {
		return ErrPreviewStale
	}
	return nil
}

func verifyAnchoredStaging(root *os.Root, preview Preview) error {
	expected := make(map[string]PreviewFile, len(preview.Files))
	for _, file := range preview.Files {
		expected[file.Path] = file
	}
	seen := map[string]struct{}{}
	err := fs.WalkDir(root.FS(), ".", func(filePath string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if filePath == "." {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return ErrGateFailed
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return ErrGateFailed
		}
		file, exists := expected[filepath.ToSlash(filePath)]
		if !exists {
			return ErrGateFailed
		}
		content, err := root.ReadFile(filePath)
		if err != nil || string(content) != file.Content {
			return ErrGateFailed
		}
		hash := sha256.Sum256(content)
		if hex.EncodeToString(hash[:]) != file.SHA256 {
			return ErrGateFailed
		}
		seen[filepath.ToSlash(filePath)] = struct{}{}
		return nil
	})
	if err != nil || len(seen) != len(expected) {
		return ErrGateFailed
	}
	return nil
}
