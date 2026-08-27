package generator

import (
	"context"
	"io/fs"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const transportGenerationTimeout = 2 * time.Minute

var transportGenerationSemaphore = make(chan struct{}, 1)

type OutputRenderer interface {
	Render(context.Context, Model) ([]PreviewFile, error)
}
type TransportGenerator interface {
	Generate(context.Context, Model, PreviewFile) ([]PreviewFile, error)
}

type CanonicalRenderer struct{ transport TransportGenerator }

func NewCanonicalRenderer(transport TransportGenerator) (*CanonicalRenderer, error) {
	if transport == nil {
		return nil, ErrInvalid
	}
	return &CanonicalRenderer{transport: transport}, nil
}
func (renderer *CanonicalRenderer) Render(ctx context.Context, model Model) ([]PreviewFile, error) {
	files, err := renderBase(model)
	if err != nil {
		return nil, err
	}
	contractPath := "contracts/openapi/modules/" + model.Module + ".yaml"
	var contract PreviewFile
	for _, file := range files {
		if file.Path == contractPath {
			contract = file
			break
		}
	}
	if contract.Path == "" {
		return nil, ErrInternal
	}
	generated, err := renderer.transport.Generate(ctx, model, contract)
	if err != nil {
		return nil, err
	}
	return append(files, generated...), nil
}

type CanonicalTransportGenerator struct{ repositoryRoot string }

func NewCanonicalTransportGenerator(repositoryRoot string) (*CanonicalTransportGenerator, error) {
	root, err := canonicalDirectory(repositoryRoot)
	if err != nil {
		return nil, ErrInvalid
	}
	for _, path := range []string{"scripts/contracts/cli.mjs", "contracts/openapi/components"} {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(path))); err != nil {
			return nil, ErrInvalid
		}
	}
	return &CanonicalTransportGenerator{repositoryRoot: root}, nil
}

func (generator *CanonicalTransportGenerator) Generate(ctx context.Context, model Model, contract PreviewFile) ([]PreviewFile, error) {
	select {
	case transportGenerationSemaphore <- struct{}{}:
		defer func() { <-transportGenerationSemaphore }()
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	}
	ctx, cancel := context.WithTimeout(ctx, transportGenerationTimeout)
	defer cancel()
	temporary, err := os.MkdirTemp("", "go-admin-generator-transport-")
	if err != nil {
		return nil, ErrInternal
	}
	defer os.RemoveAll(temporary)
	contractPath := filepath.Join(temporary, filepath.FromSlash(contract.Path))
	if err := os.MkdirAll(filepath.Dir(contractPath), 0o750); err != nil {
		return nil, ErrInternal
	}
	if err := os.WriteFile(contractPath, []byte(contract.Content), 0o640); err != nil {
		return nil, ErrInternal
	}
	if err := copyDirectory(filepath.Join(generator.repositoryRoot, "contracts/openapi/components"), filepath.Join(temporary, "contracts/openapi/components")); err != nil {
		return nil, ErrInternal
	}
	output := filepath.Join(temporary, "output")
	if err := os.Mkdir(output, 0o750); err != nil {
		return nil, ErrInternal
	}
	cliURL := (&url.URL{Scheme: "file", Path: filepath.Join(generator.repositoryRoot, "scripts/contracts/cli.mjs")}).String()
	script := `const { generate, lintContracts } = await import(process.argv[1]); lintContracts([process.argv[3]]); generate(process.argv[2], [process.argv[3]]);`
	nodeExecutable, err := resolveToolExecutable("node")
	if err != nil {
		return nil, ErrGateFailed
	}
	command := exec.CommandContext(ctx, nodeExecutable, "--input-type=module", "-e", script, cliURL, output, contractPath)
	command.Dir = generator.repositoryRoot
	safeHome := filepath.Join(temporary, "home")
	if err := os.Mkdir(safeHome, 0o700); err != nil {
		return nil, ErrInternal
	}
	command.Env = minimalCommandEnvironment(safeHome)
	if outputBytes, err := command.CombinedOutput(); err != nil {
		_ = outputBytes
		if context.Cause(ctx) != nil {
			return nil, context.Cause(ctx)
		}
		return nil, ErrGateFailed
	}
	paths := []string{
		"go-admin-plus/internal/modules/" + model.Module + "/transport/openapi.gen.go",
		"go-admin-plus/internal/modules/" + model.Module + "/transport/openapi.json",
		"go-admin-plus/internal/modules/" + model.Module + "/transport/openapi.manifest.json",
		"go-admin-plus-ui/packages/domains/" + model.Module + "/src/generated/schema.ts",
		"go-admin-plus-ui/packages/domains/" + model.Module + "/src/generated/client.ts",
	}
	files := make([]PreviewFile, 0, len(paths))
	for _, path := range paths {
		content, err := os.ReadFile(filepath.Join(output, filepath.FromSlash(path)))
		if err != nil {
			return nil, ErrInternal
		}
		files = append(files, PreviewFile{Path: path, Content: string(content)})
	}
	return files, nil
}

func copyDirectory(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if entry.Type()&os.ModeSymlink != 0 {
			return ErrInvalid
		}
		if entry.IsDir() {
			return os.MkdirAll(target, 0o750)
		}
		if !entry.Type().IsRegular() {
			return ErrInvalid
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, content, 0o640)
	})
}
