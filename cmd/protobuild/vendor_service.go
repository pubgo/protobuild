package protobuild

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/pathutil"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/protobuild/internal/depresolver"
	"github.com/schollz/progressbar/v3"
)

// VendorService handles dependency vendoring operations.
type VendorService struct {
	resolver *depresolver.Manager
	config   *Config
}

// NewVendorService creates a new VendorService.
func NewVendorService(config *Config) *VendorService {
	return &VendorService{
		resolver: depresolver.NewManager("", ""),
		config:   config,
	}
}

// VendorResult contains the result of a vendor operation.
type VendorResult struct {
	ResolvedPaths map[string]string // dep.Name -> localPath
	FailedDeps    []string
	Changed       bool
}

// ResolveDependencies resolves all configured dependencies.
func (s *VendorService) ResolveDependencies(ctx context.Context, update bool) (*VendorResult, error) {
	result := &VendorResult{
		ResolvedPaths: make(map[string]string),
	}

	// Filter valid dependencies
	validDeps := s.filterValidDeps()
	if len(validDeps) == 0 {
		return result, nil
	}

	fmt.Printf("\n🔍 Resolving %d dependencies...\n\n", len(validDeps))

	// Clean cache if update flag is set
	if update {
		fmt.Println("🗑️  Cleaning dependency cache...")
		_ = s.resolver.CleanCache()
		fmt.Println()
	}

	// Resolve each dependency
	for i, dep := range s.config.Depends {
		if dep.Name == "" || dep.Url == "" {
			continue
		}

		resolverDep := s.toResolverDep(dep)
		resolved, err := s.resolver.Resolve(ctx, resolverDep)

		if err != nil {
			if dep.Optional != nil && *dep.Optional {
				fmt.Printf("  ⚠️  [optional] %s - skipped\n", dep.Name)
				continue
			}
			fmt.Print(err.Error())
			result.FailedDeps = append(result.FailedDeps, dep.Name)
			continue
		}

		if resolved.LocalPath == "" {
			continue
		}

		if resolved.Changed {
			result.Changed = true
			fmt.Printf("  ✅ %s (downloaded)\n", dep.Name)
		} else {
			fmt.Printf("  ✅ %s (cached)\n", dep.Name)
		}

		// Update version in config if resolved
		if resolved.Version != "" {
			s.config.Depends[i].Version = &resolved.Version
		}

		result.ResolvedPaths[dep.Name] = resolved.LocalPath
	}

	return result, nil
}

// CopyToVendor copies resolved dependencies to the vendor directory.
func (s *VendorService) CopyToVendor(resolvedPaths map[string]string) (int, error) {
	fmt.Printf("\n📁 Updating vendor directory: %s\n", s.config.Vendor)
	_ = os.RemoveAll(s.config.Vendor)

	totalFiles := CountProtoFiles(resolvedPaths)

	bar := progressbar.NewOptions(totalFiles,
		progressbar.OptionSetDescription("  📋 Copying proto files"),
		progressbar.OptionShowCount(),
		progressbar.OptionSetWidth(30),
		progressbar.OptionOnCompletion(func() { fmt.Println() }),
	)

	var copiedFiles int
	for name, localPath := range resolvedPaths {
		if pathutil.IsNotExist(localPath) {
			fmt.Printf("  ⚠️  Path not found: %s (%s)\n", name, localPath)
			continue
		}

		newUrl := filepath.Join(s.config.Vendor, name)
		err := filepath.Walk(localPath, func(path string, info fs.FileInfo, err error) (gErr error) {
			if err != nil {
				return err
			}

			defer recovery.Err(&gErr, func(err error) error {
				return errors.WrapTag(err,
					errors.T("path", path),
					errors.T("name", info.Name()),
				)
			})

			if info.IsDir() || !strings.HasSuffix(info.Name(), ".proto") {
				return nil
			}

			newPath := filepath.Join(newUrl, strings.TrimPrefix(path, localPath))
			assert.Must(pathutil.IsNotExistMkDir(filepath.Dir(newPath)))
			assert.Must1(copyFile(newPath, path))
			copiedFiles++
			_ = bar.Add(1)

			return nil
		})
		if err != nil {
			return copiedFiles, err
		}
	}

	_ = bar.Finish()
	return copiedFiles, nil
}

// CleanCache cleans the dependency cache.
func (s *VendorService) CleanCache() error {
	return s.resolver.CleanCache()
}

// CacheDir returns the cache directory path.
func (s *VendorService) CacheDir() string {
	return s.resolver.CacheDir()
}

// filterValidDeps returns dependencies with valid name and url.
func (s *VendorService) filterValidDeps() []*depend {
	var valid []*depend
	for _, dep := range s.config.Depends {
		if dep.Name != "" && dep.Url != "" {
			valid = append(valid, dep)
		}
	}
	return valid
}

// toResolverDep converts a config depend to depresolver.Dependency.
func (s *VendorService) toResolverDep(dep *depend) *depresolver.Dependency {
	return &depresolver.Dependency{
		Name:     dep.Name,
		Source:   depresolver.Source(dep.Source),
		URL:      dep.Url,
		Path:     dep.Path,
		Version:  dep.Version,
		Ref:      dep.Ref,
		Optional: dep.Optional,
	}
}

// copyFile copies a file from src to dst.
func copyFile(dstFilePath, srcFilePath string) (written int64, err error) {
	srcFile, err := os.Open(srcFilePath)
	if err != nil {
		return 0, fmt.Errorf("open source file %s: %w", srcFilePath, err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstFilePath)
	if err != nil {
		return 0, fmt.Errorf("create dest file %s: %w", dstFilePath, err)
	}
	defer dstFile.Close()

	return io.Copy(dstFile, srcFile)
}
