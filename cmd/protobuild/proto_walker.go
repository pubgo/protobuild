package protobuild

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/pathutil"
	"github.com/samber/lo"
)

// ProtoWalker provides utilities for walking proto directories.
type ProtoWalker struct {
	roots    []string
	excludes []string
}

// NewProtoWalker creates a new ProtoWalker.
func NewProtoWalker(roots, excludes []string) *ProtoWalker {
	return &ProtoWalker{
		roots:    roots,
		excludes: excludes,
	}
}

// WalkDirs walks all directories containing proto files.
// Returns a map of directory path to list of proto files in that directory.
func (w *ProtoWalker) WalkDirs() map[string][]string {
	result := make(map[string][]string)

	for _, root := range w.roots {
		if pathutil.IsNotExist(root) {
			continue
		}

		_ = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				return nil
			}

			// Check excludes
			for _, e := range w.excludes {
				if strings.HasPrefix(path, e) {
					return filepath.SkipDir
				}
			}

			// Get proto files in this directory
			protoFiles := w.GetProtoFiles(path)
			if len(protoFiles) > 0 {
				result[path] = protoFiles
			}

			return nil
		})
	}

	return result
}

// GetProtoFiles returns all .proto files in a directory (non-recursive).
func (w *ProtoWalker) GetProtoFiles(dir string) []string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".proto") {
			files = append(files, filepath.Join(dir, entry.Name()))
		}
	}
	return files
}

// GetAllProtoDirs returns all directories that contain proto files.
func (w *ProtoWalker) GetAllProtoDirs() []string {
	dirs := w.WalkDirs()
	return lo.Keys(dirs)
}

// HasProtoFiles checks if a directory contains proto files.
func (w *ProtoWalker) HasProtoFiles(dir string) bool {
	return len(w.GetProtoFiles(dir)) > 0
}

// CollectPluginConfigs collects plugin configurations for each proto directory.
func (w *ProtoWalker) CollectPluginConfigs(baseCfg *Config, pluginCfgName string) map[string]*Config {
	pluginMap := make(map[string]*Config)

	for _, root := range w.roots {
		if pathutil.IsNotExist(root) {
			continue
		}

		assert.Must(filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}

			if !d.IsDir() {
				return nil
			}

			// Check for directory-specific plugin config
			var pluginCfg *Config
			pluginCfgPath := filepath.Join(path, pluginCfgName)
			if pathutil.IsExist(pluginCfgPath) {
				pluginCfg = parsePluginConfig(pluginCfgPath)
			} else {
				// Inherit from parent directory
				for dir, v := range pluginMap {
					if strings.HasPrefix(path, dir) {
						pluginCfg = v
						break
					}
				}
			}

			// Merge with base config
			pluginCfg = mergePluginConfig(baseCfg, pluginCfg)

			// Check excludes
			for _, e := range pluginCfg.Excludes {
				if strings.HasPrefix(path, e) {
					return filepath.SkipDir
				}
			}

			pluginMap[path] = pluginCfg
			return nil
		}))
	}

	return pluginMap
}

// CountProtoFiles counts total proto files in resolved paths.
func CountProtoFiles(paths map[string]string) int {
	var total int
	for _, localPath := range paths {
		_ = filepath.Walk(localPath, func(path string, info fs.FileInfo, err error) error {
			if err == nil && !info.IsDir() && strings.HasSuffix(info.Name(), ".proto") {
				total++
			}
			return nil
		})
	}
	return total
}
