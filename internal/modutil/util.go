// Package modutil provides Go module utilities.
package modutil

import (
	"os"
	"path/filepath"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/pathutil"
	"github.com/pubgo/funk/v2/result"
	"github.com/samber/lo"
	"golang.org/x/mod/modfile"

	"github.com/pubgo/protobuild/internal/shutil"
)

func getFileByRecursion(file, path string) string {
	filePath := filepath.Join(path, file)
	if pathutil.IsExist(filePath) {
		return filePath
	}

	if path == string(os.PathSeparator) {
		return ""
	}

	return getFileByRecursion(file, filepath.Dir(path))
}

// GoModPath returns the path to go.mod file by searching recursively.
func GoModPath() string {
	pwd := assert.Must1(os.Getwd())
	return getFileByRecursion("go.mod", pwd)
}

// LoadVersionGraph loads the module version graph from 'go mod graph'.
func LoadVersionGraph() map[string]string {
	modList := strings.Split(result.Wrap(shutil.GoModGraph()).Unwrap(), "\n")
	modSet := mapset.NewSet[string]()
	for _, m := range modList {
		for _, v := range strings.Split(m, " ") {
			modSet.Add(strings.TrimSpace(v))
		}
	}

	modMap := make(map[string][]*ver.Version)
	modSet.Each(func(s string) bool {
		ver2 := strings.Split(s, "@")
		if len(ver2) != 2 {
			return false
		}

		if !strings.HasPrefix(ver2[1], "v") {
			return false
		}

		modMap[ver2[0]] = append(modMap[ver2[0]], ver.Must(ver.NewSemver(ver2[1])))
		return false
	})

	return lo.MapValues(modMap, func(versions []*ver.Version, _ string) string {
		return "v" + maxVersion(versions).String()
	})
}

// LoadVersions loads module versions from go.mod file.
func LoadVersions() map[string]string {
	path := GoModPath()
	assert.Assert(path == "", "go.mod not exists")

	modBytes := assert.Must1(os.ReadFile(path))

	a, err := modfile.Parse("in", modBytes, nil)
	assert.Must(err, "go.mod 解析失败")

	versions := make(map[string]string)

	for i := range a.Require {
		mod := a.Require[i].Mod
		versions[mod.Path] = mod.Version
	}

	for i := range a.Replace {
		mod := a.Replace[i].New
		versions[mod.Path] = mod.Version
	}

	return versions
}

func maxVersion(versions []*ver.Version) *ver.Version {
	return lo.MaxBy(versions, func(a, b *ver.Version) bool { return a.GreaterThan(b) })
}
