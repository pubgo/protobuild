// Package depresolver provides multi-source dependency resolution for protobuf files.
package depresolver

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hashicorp/go-getter"
	"github.com/pubgo/funk/v2/pathutil"
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

const defaultGitShallowDepth = 1

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
	Version  *string `yaml:"version,omitempty"`  // module version; for git it is used as tag/branch/commit
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
	if version := dependencyVersion(e.Dependency); version != "" {
		sb.WriteString(fmt.Sprintf("   Version: %s\n", version))
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
		sb.WriteString("   • Verify the git version (tag/branch/commit) exists\n")
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
	if dep == nil {
		return nil, fmt.Errorf("dependency is nil")
	}

	// Detect source if not specified
	source := m.detectSource(dep)
	dep.Source = source
	m.normalizeVersion(dep)

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

// normalizeVersion trims and normalizes the version field.
func (m *Manager) normalizeVersion(dep *Dependency) {
	if dep == nil {
		return
	}

	if dep.Version == nil {
		return
	}

	version := strings.TrimSpace(*dep.Version)
	if version == "" {
		dep.Version = nil
		return
	}

	if version != *dep.Version {
		v := version
		dep.Version = &v
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
	// Generate cache path from normalized source URL to maximize cache reuse.
	cachePath := m.cachePathForDependency(dep, source)

	// Check if we need to download
	changed := false
	if pathutil.IsNotExist(cachePath) {
		changed = true

		displayName := strings.TrimSpace(dep.Name)
		if displayName == "" {
			displayName = strings.TrimSpace(dep.URL)
		}

		getterURL := m.buildGetterURL(dep, source)
		fmt.Printf("  📥 [%s] %s\n", source.DisplayName(), displayName)
		fmt.Printf("     URL: %s\n", getterURL)
		if version := dependencyVersion(dep); version != "" {
			fmt.Printf("     Version: %s\n", version)
		}
		if dep.Path != "" {
			fmt.Printf("     Path: %s\n", dep.Path)
		}
		fmt.Printf("     Cache: %s\n", cachePath)

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
		Version:   dependencyVersion(dep),
		Changed:   changed,
	}, nil
}

// cachePathForDependency builds a stable cache path for getter-based sources.
func (m *Manager) cachePathForDependency(dep *Dependency, source Source) string {
	cacheKey := m.cacheKeyForDependency(dep, source)
	return filepath.Join(m.cacheDir, string(source), cacheKey)
}

// cacheKeyForDependency returns a hash key from a normalized dependency seed.
func (m *Manager) cacheKeyForDependency(dep *Dependency, source Source) string {
	return hashString(m.cacheSeedForDependency(dep, source))
}

// cacheSeedForDependency normalizes dependency coordinates for stable caching.
//
// Notes:
//   - dep.Path is intentionally excluded so multiple subpaths share one downloaded source.
//   - For getter-based sources, URL normalization follows buildGetterURL behavior.
func (m *Manager) cacheSeedForDependency(dep *Dependency, source Source) string {
	if dep == nil {
		return string(source)
	}

	trimmedURL := strings.TrimSpace(dep.URL)
	trimmedVersion := dependencyVersion(dep)

	// For getter-based sources, use canonical getter URL as the cache seed.
	if source != SourceLocal && source != SourceGoMod {
		normalized := *dep
		normalized.URL = trimmedURL
		if trimmedVersion == "" {
			normalized.Version = nil
		} else {
			v := trimmedVersion
			normalized.Version = &v
		}
		return fmt.Sprintf("%s|%s", source, m.buildGetterURL(&normalized, source))
	}

	if trimmedVersion == "" {
		return fmt.Sprintf("%s|%s", source, trimmedURL)
	}

	return fmt.Sprintf("%s|%s|%s", source, trimmedURL, trimmedVersion)
}

// downloadWithGetter uses go-getter to download dependencies
func (m *Manager) downloadWithGetter(ctx context.Context, dep *Dependency, source Source, destPath string) error {
	// Build go-getter URL with appropriate prefix and query parameters
	getterURL := m.buildGetterURL(dep, source)
	displayName := strings.TrimSpace(dep.Name)
	if displayName == "" {
		displayName = strings.TrimSpace(dep.URL)
	}

	// Create progress bar with getter-backed byte tracking
	bar := progressbar.NewOptions64(-1,
		progressbar.OptionSetDescription(fmt.Sprintf("  ↓ [%s] %s", source.DisplayName(), displayName)),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionShowBytes(true),
		progressbar.OptionShowTotalBytes(true),
		progressbar.OptionSetElapsedTime(true),
		progressbar.OptionSetPredictTime(true),
		progressbar.OptionShowElapsedTimeOnFinish(),
		progressbar.OptionSetWidth(30),
		progressbar.OptionThrottle(100*time.Millisecond),
		progressbar.OptionOnCompletion(func() { fmt.Println() }),
		progressbar.OptionSetRenderBlankState(true),
	)
	tracker := newGetterProgressTracker(bar, source, displayName)

	// Fallback spinner updates for getters that don't emit byte callbacks (e.g. some git transports).
	fallbackDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(180 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-fallbackDone:
				return
			case <-ctx.Done():
				return
			case <-ticker.C:
				if tracker.HasStreamCallback() {
					return
				}
				_ = bar.Add64(1)
			}
		}
	}()

	// Create go-getter client
	client := &getter.Client{
		Ctx:              ctx,
		Src:              getterURL,
		Dst:              destPath,
		Mode:             getter.ClientModeAny,
		ProgressListener: tracker,
		Options:          []getter.ClientOption{},
	}

	// Execute download
	err := client.Get()
	close(fallbackDone)
	tracker.Finish()

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
		// Add ref query parameter for git tag/branch/commit from version
		version := dependencyVersion(dep)
		if version != "" {
			if strings.Contains(url, "?") {
				url += "&ref=" + version
			} else {
				url += "?ref=" + version
			}
		}

		// Default to shallow clone for faster downloads, except commit SHA refs.
		if shouldUseGitShallowClone(version) && !hasGetterQueryParam(url, "depth") {
			if strings.Contains(url, "?") {
				url += "&depth=" + strconv.Itoa(defaultGitShallowDepth)
			} else {
				url += "?depth=" + strconv.Itoa(defaultGitShallowDepth)
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

func dependencyVersion(dep *Dependency) string {
	if dep == nil || dep.Version == nil {
		return ""
	}
	return strings.TrimSpace(*dep.Version)
}

func shouldUseGitShallowClone(version string) bool {
	if version == "" {
		return true
	}
	return !isLikelyGitCommit(version)
}

func isLikelyGitCommit(ref string) bool {
	if len(ref) < 7 || len(ref) > 40 {
		return false
	}
	for _, r := range ref {
		if (r < '0' || r > '9') && (r < 'a' || r > 'f') && (r < 'A' || r > 'F') {
			return false
		}
	}
	return true
}

func hasGetterQueryParam(url, key string) bool {
	return strings.Contains(url, "?"+key+"=") || strings.Contains(url, "&"+key+"=")
}

type getterProgressTracker struct {
	bar        *progressbar.ProgressBar
	label      string
	startedAt  time.Time
	bytesRead  atomic.Int64
	hasStream  atomic.Bool
	finishOnce sync.Once
}

func newGetterProgressTracker(bar *progressbar.ProgressBar, source Source, displayName string) *getterProgressTracker {
	return &getterProgressTracker{
		bar:       bar,
		label:     fmt.Sprintf("[%s] %s", source.DisplayName(), displayName),
		startedAt: time.Now(),
	}
}

// TrackProgress adapts go-getter stream callbacks into progressbar updates.
func (t *getterProgressTracker) TrackProgress(_ string, currentSize, totalSize int64, stream io.ReadCloser) io.ReadCloser {
	t.hasStream.Store(true)

	if currentSize >= 0 {
		t.bytesRead.Store(currentSize)
		_ = t.bar.Set64(currentSize)
	}

	if totalSize > 0 {
		t.bar.ChangeMax64(totalSize)
	}

	return &trackingReadCloser{
		ReadCloser: stream,
		onRead: func(n int) {
			t.bytesRead.Add(int64(n))
			_ = t.bar.Add64(int64(n))
		},
		onClose: t.Finish,
	}
}

func (t *getterProgressTracker) HasStreamCallback() bool {
	return t.hasStream.Load()
}

func (t *getterProgressTracker) Finish() {
	t.finishOnce.Do(func() {
		_ = t.bar.Finish()

		elapsed := time.Since(t.startedAt)
		if elapsed <= 0 {
			elapsed = time.Millisecond
		}

		bytes := t.bytesRead.Load()
		if bytes <= 0 {
			fmt.Printf("     ✅ Download complete: %s (elapsed %s)\n", t.label, elapsed.Round(100*time.Millisecond))
			return
		}

		rateBytes := int64(float64(bytes) / elapsed.Seconds())
		fmt.Printf("     ✅ Download complete: %s, %s in %s (avg %s/s)\n",
			t.label,
			formatBinaryBytes(bytes),
			elapsed.Round(100*time.Millisecond),
			formatBinaryBytes(rateBytes),
		)
	})
}

type trackingReadCloser struct {
	io.ReadCloser
	onRead  func(n int)
	onClose func()
	closed  sync.Once
}

func (r *trackingReadCloser) Read(p []byte) (int, error) {
	n, err := r.ReadCloser.Read(p)
	if n > 0 && r.onRead != nil {
		r.onRead(n)
	}
	if err == io.EOF {
		r.closeOnce()
	}
	return n, err
}

func (r *trackingReadCloser) Close() error {
	err := r.ReadCloser.Close()
	r.closeOnce()
	return err
}

func (r *trackingReadCloser) closeOnce() {
	r.closed.Do(func() {
		if r.onClose != nil {
			r.onClose()
		}
	})
}

func formatBinaryBytes(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	v := float64(n) / 1024
	unit := 0
	for v >= 1024 && unit < len(units)-1 {
		v /= 1024
		unit++
	}

	if v >= 10 {
		return fmt.Sprintf("%.0f %s", v, units[unit])
	}
	return fmt.Sprintf("%.1f %s", v, units[unit])
}
