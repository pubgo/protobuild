// Package upgradecmd implements the upgrade command for protobuild
package upgradecmd

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/hashicorp/go-getter"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/protobuild/internal/githubclient"
	"github.com/pubgo/redant"
)

func New() *redant.Command {
	const repo = "pubgo/protobuild"
	var tag, assetName, output string
	var dryRun, force bool

	return &redant.Command{
		Use:   "upgrade",
		Short: "Upgrade protobuild from GitHub release",
		Options: redant.OptionSet{
			{
				Flag:        "tag",
				Description: "Release tag (default: latest)",
				Default:     "latest",
				Value:       redant.StringOf(&tag),
			},
			{
				Flag:        "asset",
				Description: "Asset name to download (auto-detect if empty)",
				Value:       redant.StringOf(&assetName),
			},
			{
				Flag:        "output",
				Description: "Output file path (default: current executable)",
				Value:       redant.StringOf(&output),
			},
			{
				Flag:        "dry-run",
				Description: "Show release info only, do not download",
				Value:       redant.BoolOf(&dryRun),
			},
			{
				Flag:        "force",
				Description: "Force overwrite output file",
				Value:       redant.BoolOf(&force),
			},
		},
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			parts := strings.Split(repo, "/")
			if len(parts) != 2 {
				return errors.New("invalid --repo format, should be owner/repo")
			}
			owner, repoName := parts[0], parts[1]

			client := githubclient.NewPublicRelease(owner, repoName)
			var release *githubclient.RepositoryRelease
			var err error
			if tag != "" {
				release, err = client.GetByTag(ctx, tag)
			} else {
				release, err = client.Latest(ctx)
			}
			if err != nil {
				return errors.Wrap(err, "failed to get release")
			}

			assets := githubclient.GetAssets(release)
			var asset githubclient.Asset
			if assetName != "" {
				found := false
				for _, a := range assets {
					if a.Name == assetName || strings.Contains(a.URL, assetName) {
						asset = a
						found = true
						break
					}
				}
				if !found {
					return errors.New("asset not found in release")
				}
			} else {
				for _, a := range assets {
					if a.OS == runtime.GOOS && a.Arch == runtime.GOARCH && !a.IsChecksumFile() {
						asset = a
						break
					}
				}
				if asset.URL == "" {
					return errors.New("no matching asset for current platform")
				}
			}

			fmt.Printf("Release: %s\nAsset: %s (%s)\nURL: %s\n", release.TagName, asset.Name, githubclient.GetSizeFormat(asset.Size), asset.URL)
			if dryRun {
				return nil
			}

			tmpFile := filepath.Join(os.TempDir(), asset.Name+".download")
			c := &getter.Client{
				Ctx:              ctx,
				Src:              asset.URL,
				Dst:              tmpFile,
				Mode:             getter.ClientModeFile,
				ProgressListener: defaultProgressBar,
			}
			if err := c.Get(); err != nil {
				return errors.Wrap(err, "download failed")
			}
			defer os.Remove(tmpFile) // Clean up on error

			// Extract if needed (tar.gz, zip, etc.)
			extractedFile := tmpFile
			if strings.HasSuffix(asset.Name, ".tar.gz") || strings.HasSuffix(asset.Name, ".tgz") {
				var err error
				extractedFile, err = extractTarGz(tmpFile)
				if err != nil {
					return errors.Wrap(err, "failed to extract tar.gz")
				}
				defer os.Remove(extractedFile)
			} else if strings.HasSuffix(asset.Name, ".zip") {
				var err error
				extractedFile, err = extractZip(tmpFile)
				if err != nil {
					return errors.Wrap(err, "failed to extract zip")
				}
				defer os.Remove(extractedFile)
			}

			// Determine output path
			if output == "" {
				execPath, err := os.Executable()
				if err != nil {
					// Fallback to os.Args[0]
					execPath = os.Args[0]
					if !filepath.IsAbs(execPath) {
						// Try to resolve relative path
						cwd, _ := os.Getwd()
						execPath = filepath.Join(cwd, execPath)
					}
				}
				output = execPath
			}

			// Check if output file exists and is writable
			if !force {
				if fi, err := os.Stat(output); err == nil {
					if fi.Mode()&0200 == 0 {
						return errors.New("output file is not writable, use --force to override")
					}
				}
			}

			// Copy extracted file to output location
			if err := copyFile(extractedFile, output); err != nil {
				return errors.Wrap(err, "failed to replace output file")
			}

			// Set executable permission on Unix-like systems
			if runtime.GOOS != "windows" {
				if err := os.Chmod(output, 0755); err != nil {
					return errors.Wrap(err, "failed to set executable permission")
				}
			}

			fmt.Printf("✅ Upgraded successfully: %s\n", output)
			return nil
		},
	}
}

// extractTarGz extracts a tar.gz file and returns the path to the extracted binary.
func extractTarGz(tarGzPath string) (string, error) {
	file, err := os.Open(tarGzPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return "", err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)
	extractDir := filepath.Dir(tarGzPath)
	var binaryPath string

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}

		// Skip directories and non-regular files
		if header.Typeflag != tar.TypeReg {
			continue
		}

		// Look for executable files (common binary names)
		name := filepath.Base(header.Name)
		if strings.Contains(name, "protobuild") ||
			(runtime.GOOS == "windows" && strings.HasSuffix(name, ".exe")) ||
			(runtime.GOOS != "windows" && !strings.Contains(name, ".")) {
			target := filepath.Join(extractDir, name)
			outFile, err := os.Create(target)
			if err != nil {
				return "", err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return "", err
			}
			outFile.Close()
			binaryPath = target
			break
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("no binary found in archive")
	}
	return binaryPath, nil
}

// extractZip extracts a zip file and returns the path to the extracted binary.
func extractZip(zipPath string) (string, error) {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", err
	}
	defer r.Close()

	extractDir := filepath.Dir(zipPath)
	var binaryPath string

	for _, f := range r.File {
		// Skip directories
		if f.FileInfo().IsDir() {
			continue
		}

		// Look for executable files
		name := filepath.Base(f.Name)
		if strings.Contains(name, "protobuild") ||
			(runtime.GOOS == "windows" && strings.HasSuffix(name, ".exe")) ||
			(runtime.GOOS != "windows" && !strings.Contains(name, ".")) {
			rc, err := f.Open()
			if err != nil {
				continue
			}

			target := filepath.Join(extractDir, name)
			outFile, err := os.Create(target)
			if err != nil {
				rc.Close()
				continue
			}

			if _, err := io.Copy(outFile, rc); err != nil {
				rc.Close()
				outFile.Close()
				continue
			}

			rc.Close()
			outFile.Close()
			binaryPath = target
			break
		}
	}

	if binaryPath == "" {
		return "", fmt.Errorf("no binary found in archive")
	}
	return binaryPath, nil
}

// copyFile copies a file from src to dst, replacing dst if it exists.
func copyFile(src, dst string) error {
	// On Windows, we need to remove the destination file first
	if runtime.GOOS == "windows" {
		if _, err := os.Stat(dst); err == nil {
			if err := os.Remove(dst); err != nil {
				return fmt.Errorf("failed to remove existing file: %w", err)
			}
		}
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return err
	}

	return dstFile.Sync()
}
