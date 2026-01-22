package githubclient

import (
	"regexp"
	"strings"
)

type Asset struct {
	Name string
	URL  string
	Size int64
	OS   string
	Arch string
}

func (a Asset) IsChecksumFile() bool {
	return strings.HasSuffix(a.Name, ".sha256") || strings.HasSuffix(a.Name, ".md5")
}

// GetAssets extracts and parses assets from a release.
func GetAssets(release *RepositoryRelease) []Asset {
	var assets []Asset
	for _, a := range release.Assets {
		os, arch := parseAssetPlatform(a.Name)
		assets = append(assets, Asset{
			Name: a.Name,
			URL:  a.BrowserDownloadURL,
			Size:  a.Size,
			OS:    os,
			Arch:  arch,
		})
	}
	return assets
}

// parseAssetPlatform parses OS and architecture from asset filename.
// Supports common naming patterns like:
// - protobuild-darwin-amd64
// - protobuild-linux-arm64
// - protobuild-windows-amd64.exe
// - protobuild-v1.0.0-darwin-arm64.tar.gz
func parseAssetPlatform(filename string) (os, arch string) {
	filename = strings.ToLower(filename)
	
	// Common OS patterns
	osPatterns := map[string]string{
		"darwin":  "darwin",
		"macos":   "darwin",
		"mac":     "darwin",
		"linux":   "linux",
		"windows": "windows",
		"win":     "windows",
	}
	
	// Common arch patterns
	archPatterns := map[string]string{
		"amd64":   "amd64",
		"x86_64":  "amd64",
		"x64":     "amd64",
		"386":     "386",
		"i386":    "386",
		"x86":     "386",
		"arm64":   "arm64",
		"aarch64": "arm64",
		"arm":     "arm",
		"armv6":   "arm",
		"armv7":   "arm",
	}
	
	// Try to find OS
	for pattern, osName := range osPatterns {
		if strings.Contains(filename, pattern) {
			os = osName
			break
		}
	}
	
	// Try to find arch
	for pattern, archName := range archPatterns {
		if strings.Contains(filename, pattern) {
			arch = archName
			break
		}
	}
	
	// If not found, try regex patterns
	if os == "" || arch == "" {
		// Pattern: {name}-{os}-{arch} or {name}-{version}-{os}-{arch}
		re := regexp.MustCompile(`(darwin|linux|windows|macos|mac|win)[-_](amd64|x86_64|x64|386|i386|x86|arm64|aarch64|arm|armv6|armv7)`)
		matches := re.FindStringSubmatch(filename)
		if len(matches) >= 3 {
			if os == "" {
				osName := strings.ToLower(matches[1])
				if osName == "macos" || osName == "mac" {
					os = "darwin"
				} else if osName == "win" {
					os = "windows"
				} else {
					os = osName
				}
			}
			if arch == "" {
				archName := strings.ToLower(matches[2])
				if archName == "x86_64" || archName == "x64" {
					arch = "amd64"
				} else if archName == "i386" || archName == "x86" {
					arch = "386"
				} else if archName == "aarch64" {
					arch = "arm64"
				} else if archName == "armv6" || archName == "armv7" {
					arch = "arm"
				} else {
					arch = archName
				}
			}
		}
	}
	
	return
}
