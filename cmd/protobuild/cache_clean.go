package protobuild

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// cacheEntry is one cached artifact (a downloaded dependency checkout).
type cacheEntry struct {
	name    string // display name: <source>/<key>
	path    string
	size    int64
	modTime time.Time
}

// parseCleanDuration parses durations like "30d", "12h", "45m".
// The "d" (day) suffix is not covered by time.ParseDuration and is the most
// natural unit for cache retention, so it is handled explicitly.
func parseCleanDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0, fmt.Errorf("空的时长")
	}

	if strings.HasSuffix(s, "d") {
		days, err := strconv.ParseFloat(strings.TrimSuffix(s, "d"), 64)
		if err != nil || days <= 0 {
			return 0, fmt.Errorf("无效的时长 %q", s)
		}
		return time.Duration(days * 24 * float64(time.Hour)), nil
	}

	d, err := time.ParseDuration(s)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("无效的时长 %q", s)
	}
	return d, nil
}

// parseCleanSize parses sizes like "500MB", "1GB", "10GiB", "1024KB".
// Decimal (KB/MB/GB/TB) and binary (KiB/MiB/GiB/TiB) units are supported,
// plain bytes ("1048576", "512B") as well.
func parseCleanSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))
	if s == "" {
		return 0, fmt.Errorf("空的大小")
	}

	units := []struct {
		suffix string
		multi  int64
	}{
		{"KIB", 1 << 10},
		{"MIB", 1 << 20},
		{"GIB", 1 << 30},
		{"TIB", 1 << 40},
		{"KB", 1e3},
		{"MB", 1e6},
		{"GB", 1e9},
		{"TB", 1e12},
		{"B", 1},
	}

	for _, u := range units {
		if !strings.HasSuffix(s, u.suffix) {
			continue
		}

		numStr := strings.TrimSpace(strings.TrimSuffix(s, u.suffix))
		num, err := strconv.ParseFloat(numStr, 64)
		if err != nil || num <= 0 {
			return 0, fmt.Errorf("无效的大小 %q", s)
		}
		return int64(num * float64(u.multi)), nil
	}

	// plain number -> bytes
	if num, err := strconv.ParseFloat(s, 64); err == nil && num > 0 {
		return int64(num), nil
	}

	return 0, fmt.Errorf("无效的大小 %q", s)
}

// collectCacheEntries lists cached artifacts: directories two levels under
// the cache root (<cache>/<source>/<key>), with their size and mtime.
func collectCacheEntries(cacheDir string) []cacheEntry {
	var entries []cacheEntry

	sourceDirs, err := os.ReadDir(cacheDir)
	if err != nil {
		return nil
	}

	for _, sourceDir := range sourceDirs {
		if !sourceDir.IsDir() {
			continue
		}

		sourcePath := filepath.Join(cacheDir, sourceDir.Name())
		keyDirs, err := os.ReadDir(sourcePath)
		if err != nil {
			continue
		}

		for _, keyDir := range keyDirs {
			entryPath := filepath.Join(sourcePath, keyDir.Name())
			info, err := os.Stat(entryPath)
			if err != nil {
				continue
			}

			entries = append(entries, cacheEntry{
				name:    sourceDir.Name() + "/" + keyDir.Name(),
				path:    entryPath,
				size:    dirSize(entryPath),
				modTime: info.ModTime(),
			})
		}
	}

	return entries
}

// dirSize walks a directory and sums file sizes.
func dirSize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if info, err := d.Info(); err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

// planCacheRemoval selects entries to delete:
//   - everything older than cutoff (when cutoff is non-zero)
//   - then, still above maxSize, the oldest entries first
func planCacheRemoval(entries []cacheEntry, cutoff time.Time, maxSize int64) []cacheEntry {
	var remove []cacheEntry

	var kept []cacheEntry
	var keptSize int64
	for _, e := range entries {
		if !cutoff.IsZero() && e.modTime.Before(cutoff) {
			remove = append(remove, e)
			continue
		}
		kept = append(kept, e)
		keptSize += e.size
	}

	if maxSize > 0 && keptSize > maxSize {
		sort.Slice(kept, func(i, j int) bool { return kept[i].modTime.Before(kept[j].modTime) })
		for _, e := range kept {
			if keptSize <= maxSize {
				break
			}
			remove = append(remove, e)
			keptSize -= e.size
		}
	}

	sort.Slice(remove, func(i, j int) bool { return remove[i].name < remove[j].name })
	return remove
}
