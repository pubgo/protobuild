// Package depresolver provides multi-source dependency resolution for protobuf files.
package depresolver

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/pubgo/funk/pathutil"
	"github.com/schollz/progressbar/v3"
)

// Source represents the dependency source type
type Source string

// Supported source types.
const (
	SourceAuto  Source = ""      // auto-detect
	SourceGoMod Source = "gomod" // Go module (default)
	SourceGit   Source = "git"   // Git repository
	SourceHTTP  Source = "http"  // HTTP/HTTPS (supports archives)
	SourceS3    Source = "s3"    // AWS S3
	SourceGCS   Source = "gcs"   // Google Cloud Storage
	SourceLocal Source = "local" // Local path
)

// DisplayName returns a human-readable name for the source type.
func (s Source) DisplayName() string {
	switch s {
	case SourceGoMod:
		return "Go Module"
	case SourceGit:
		return "Git"
	case SourceHTTP:
		return "HTTP"
	case SourceS3:
		return "AWS S3"
	case SourceGCS:
		return "Google Cloud Storage"
	case SourceLocal:
		return "Local"
	default:
		return "Auto"
	}
}

// Dependency represents a proto dependency configuration
type Dependency struct {
	Name     string  `yaml:"name"`
	Source   Source  `yaml:"source,omitempty"` // default: auto-detect -> gomod
	URL      string  `yaml:"url"`
	Path     string  `yaml:"path,omitempty"`     // subdirectory within the source
	Version  *string `yaml:"version,omitempty"`  // version for gomod
	Ref      string  `yaml:"ref,omitempty"`      // for git: tag/branch/commit
	Optional *bool   `yaml:"optional,omitempty"` // skip if not found
}

// ResolveResult contains the result of dependency resolution
type ResolveResult struct {
	LocalPath string // local path to the resolved dependency
	Version   string // resolved version
	Changed   bool   // whether the dependency was updated
}

// ResolveError provides detailed error information for dependency resolution failures
type ResolveError struct {
	Dependency *Dependency
	Source     Source
	URL        string
	Operation  string // "download", "resolve", "validate"
	Err        error
}

func (e *ResolveError) Error() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("\n❌ Failed to %s dependency: %s\n", e.Operation, e.Dependency.Name))
	sb.WriteString(fmt.Sprintf("   Source:  %s\n", e.Source.DisplayName()))
	sb.WriteString(fmt.Sprintf("   URL:     %s\n", e.URL))
	if e.Dependency.Ref != "" {
		sb.WriteString(fmt.Sprintf("   Ref:     %s\n", e.Dependency.Ref))
	}
	if e.Dependency.Path != "" {
		sb.WriteString(fmt.Sprintf("   Path:    %s\n", e.Dependency.Path))
	}
	sb.WriteString(fmt.Sprintf("   Error:   %s\n", e.Err.Error()))
	sb.WriteString("\n💡 Suggestions:\n")

	// Add helpful suggestions based on source type and error
	switch e.Source {
	case SourceGit:
		sb.WriteString("   • Check if the repository URL is correct and accessible\n")
		sb.WriteString("   • Verify the ref (tag/branch/commit) exists\n")
		sb.WriteString("   • Ensure you have proper authentication (SSH key or token)\n")
	case SourceHTTP:
		sb.WriteString("   • Check if the URL is correct and the file exists\n")
		sb.WriteString("   • Verify your network connection\n")
		sb.WriteString("   • Check if authentication is required\n")
	case SourceS3:
		sb.WriteString("   • Check AWS credentials (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)\n")
		sb.WriteString("   • Verify the bucket and path exist\n")
		sb.WriteString("   • Check bucket permissions\n")
	case SourceGCS:
		sb.WriteString("   • Check Google Cloud credentials (GOOGLE_APPLICATION_CREDENTIALS)\n")
		sb.WriteString("   • Verify the bucket and path exist\n")
		sb.WriteString("   • Check bucket permissions\n")
	case SourceGoMod:
		sb.WriteString("   • Check if the module path is correct\n")
		sb.WriteString("   • Verify the version exists in the module\n")
		sb.WriteString("   • Run 'go mod tidy' to update dependencies\n")
	case SourceLocal:
		sb.WriteString("   • Check if the local path exists\n")
		sb.WriteString("   • Verify read permissions\n")
	}

	return sb.String()
}

func (e *ResolveError) Unwrap() error {
	return e.Err
}

// Manager manages dependency resolution
type Manager struct {
	cacheDir  string
	gomodPath string // $GOPATH/pkg/mod
}

// NewManager creates a new dependency manager
func NewManager(cacheDir, gomodPath string) *Manager {
	if cacheDir == "" {
		home, _ := os.UserHomeDir()
		if home == "" {
			home = ".local"
		}
		cacheDir = filepath.Join(home, ".cache", "protobuild", "deps")
	}
	if gomodPath == "" {
		gomodPath = filepath.Join(os.Getenv("GOPATH"), "pkg", "mod")
	}

	return &Manager{
		cacheDir:  cacheDir,
		gomodPath: gomodPath,
	}
}

// Resolve resolves a dependency
func (m *Manager) Resolve(ctx context.Context, dep *Dependency) (*ResolveResult, error) {
	// Detect source if not specified
	source := m.detectSource(dep)
	dep.Source = source

	switch source {
	case SourceLocal:
		return m.resolveLocal(dep)
	case SourceGoMod:
		return m.resolveGoMod(ctx, dep)
	default:
		// Use go-getter for git, http, s3, gcs sources
		return m.resolveWithGetter(ctx, dep, source)
	}
}

// detectSource auto-detects the source type based on URL patterns
func (m *Manager) detectSource(dep *Dependency) Source {
	if dep.Source != SourceAuto {
		return dep.Source
	}
	return DetectSource(dep.URL)
}

// DetectSource auto-detects the source type based on URL patterns (public API)
func DetectSource(url string) Source {
	// Local path detection
	if filepath.IsAbs(url) {
		return SourceLocal
	}
	if pathutil.IsExist(url) {
		return SourceLocal
	}

	// S3
	if strings.HasPrefix(url, "s3://") || strings.HasPrefix(url, "s3::") {
		return SourceS3
	}

	// GCS
	if strings.HasPrefix(url, "gcs://") || strings.HasPrefix(url, "gs://") {
		return SourceGCS
	}

	// Git (explicit .git or git:: prefix or git@)
	if strings.HasSuffix(url, ".git") || strings.HasPrefix(url, "git::") || strings.HasPrefix(url, "git@") {
		return SourceGit
	}

	// HTTP archives
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return SourceHTTP
	}

	// Default to gomod for Go module-like paths (e.g., github.com/user/repo)
	return SourceGoMod
}

// resolveLocal resolves a local dependency
func (m *Manager) resolveLocal(dep *Dependency) (*ResolveResult, error) {
	url := os.ExpandEnv(dep.URL)

	localPath := url
	if dep.Path != "" {
		localPath = filepath.Join(url, dep.Path)
	}

	absPath, err := filepath.Abs(localPath)
	if err != nil {
		return nil, &ResolveError{
			Dependency: dep,
			Source:     SourceLocal,
			URL:        url,
			Operation:  "resolve",
			Err:        fmt.Errorf("failed to get absolute path: %w", err),
		}
	}

	if pathutil.IsNotExist(absPath) {
		if dep.Optional != nil && *dep.Optional {
			return &ResolveResult{LocalPath: "", Changed: false}, nil
		}
		return nil, &ResolveError{
			Dependency: dep,
			Source:     SourceLocal,
			URL:        absPath,
			Operation:  "resolve",
			Err:        fmt.Errorf("path does not exist"),
		}
	}

	return &ResolveResult{
		LocalPath: absPath,
		Changed:   false,
	}, nil
}

// resolveWithGetter resolves dependencies using go-getter (supports git, http, s3, gcs)
func (m *Manager) resolveWithGetter(ctx context.Context, dep *Dependency, source Source) (*ResolveResult, error) {
	// Generate cache path based on source type and URL hash
	cacheKey := hashString(fmt.Sprintf("%s@%s", dep.URL, dep.Ref))
	cachePath := filepath.Join(m.cacheDir, string(source), cacheKey)

	// Check if we need to download
	changed := false
	if pathutil.IsNotExist(cachePath) {
		changed = true

		// Ensure cache directory exists
		if err := os.MkdirAll(filepath.Dir(cachePath), 0o755); err != nil {
			return nil, &ResolveError{
				Dependency: dep,
				Source:     source,
				URL:        dep.URL,
				Operation:  "resolve",
				Err:        fmt.Errorf("failed to create cache directory: %w", err),
			}
		}

		// Download using go-getter
		if err := m.downloadWithGetter(ctx, dep, source, cachePath); err != nil {
			if dep.Optional != nil && *dep.Optional {
				return &ResolveResult{LocalPath: "", Changed: false}, nil
			}
			return nil, err
		}
	}

	// Build final path with subdirectory
	localPath := cachePath
	if dep.Path != "" {
		localPath = filepath.Join(cachePath, dep.Path)
	}

	// Verify path exists
	if pathutil.IsNotExist(localPath) {
		if dep.Optional != nil && *dep.Optional {
			return &ResolveResult{LocalPath: "", Changed: false}, nil
		}
		return nil, &ResolveError{
			Dependency: dep,
			Source:     source,
			URL:        dep.URL,
			Operation:  "validate",
			Err:        fmt.Errorf("subdirectory '%s' not found in downloaded content", dep.Path),
		}
	}

	return &ResolveResult{
		LocalPath: localPath,
		Version:   dep.Ref,
		Changed:   changed,
	}, nil
}

// downloadWithGetter uses go-getter to download dependencies
func (m *Manager) downloadWithGetter(ctx context.Context, dep *Dependency, source Source, destPath string) error {
	// Build go-getter URL with appropriate prefix and query parameters
	getterURL := m.buildGetterURL(dep, source)

	// Create progress bar for visual feedback
	bar := progressbar.NewOptions(-1,
		progressbar.OptionSetDescription(fmt.Sprintf("  ↓ [%s] %s", source.DisplayName(), dep.Name)),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionShowBytes(false),
		progressbar.OptionSetWidth(30),
		progressbar.OptionThrottle(100),
		progressbar.OptionOnCompletion(func() { fmt.Println() }),
		progressbar.OptionSetRenderBlankState(true),
	)

	// Start spinner in background
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			case <-ctx.Done():
				return
			default:
				_ = bar.Add(1)
			}
		}
	}()

	// Create go-getter client
	client := &getter.Client{
		Ctx:     ctx,
		Src:     getterURL,
		Dst:     destPath,
		Mode:    getter.ClientModeAny,
		Options: []getter.ClientOption{},
	}

	// Execute download
	err := client.Get()
	close(done)
	_ = bar.Finish()

	if err != nil {
		return &ResolveError{
			Dependency: dep,
			Source:     source,
			URL:        getterURL,
			Operation:  "download",
			Err:        err,
		}
	}
	return nil
}

// buildGetterURL constructs the go-getter URL with appropriate prefix and query parameters
func (m *Manager) buildGetterURL(dep *Dependency, source Source) string {
	url := dep.URL

	switch source {
	case SourceGit:
		// Add git:: prefix if not present
		if !strings.HasPrefix(url, "git::") {
			// Handle git@ URLs (SSH)
			if strings.HasPrefix(url, "git@") {
				url = "git::" + url
			} else if !strings.Contains(url, "://") {
				// Add https:// for bare domain paths
				url = "git::https://" + url
			} else {
				url = "git::" + url
			}
		}
		// Add ref query parameter for version/branch/tag
		if dep.Ref != "" {
			if strings.Contains(url, "?") {
				url += "&ref=" + dep.Ref
			} else {
				url += "?ref=" + dep.Ref
			}
		}

	case SourceS3:
		// go-getter supports s3:// directly
		// Format: s3::https://s3.amazonaws.com/bucket/key or s3://bucket/key
		if !strings.HasPrefix(url, "s3::") && !strings.HasPrefix(url, "s3://") {
			url = "s3::" + url
		}

	case SourceGCS:
		// go-getter supports gcs:// directly
		// Format: gcs::https://www.googleapis.com/storage/v1/bucket or gs://bucket/path
		if strings.HasPrefix(url, "gs://") {
			// Convert gs:// to gcs:// format
			url = "gcs://" + strings.TrimPrefix(url, "gs://")
		} else if !strings.HasPrefix(url, "gcs://") && !strings.HasPrefix(url, "gcs::") {
			url = "gcs::" + url
		}

	case SourceHTTP:
		// HTTP URLs work directly, go-getter handles archive extraction automatically
		// No modification needed

	case SourceLocal:
		// Local paths work directly
		// Ensure it's an absolute path
		if !filepath.IsAbs(url) {
			absPath, err := filepath.Abs(url)
			if err == nil {
				url = absPath
			}
		}
	}

	return url
}

// CacheDir returns the cache directory
func (m *Manager) CacheDir() string {
	return m.cacheDir
}

// CleanCache removes all cached dependencies
func (m *Manager) CleanCache() error {
	return os.RemoveAll(m.cacheDir)
}

// hashString returns a short hash of the input string
func hashString(s string) string {
	h := sha256.Sum256([]byte(s))
	return fmt.Sprintf("%x", h[:12])
}
