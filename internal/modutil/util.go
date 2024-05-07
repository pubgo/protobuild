package modutil

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/pathutil"
	"golang.org/x/mod/modfile"
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

func GoModPath() string {
	pwd := assert.Must1(os.Getwd())
	return getFileByRecursion("go.mod", pwd)
}

func LoadVersions() map[string]string {
	path := GoModPath()
	assert.Assert(path == "", "go.mod not exists")

	modBytes := assert.Must1(ioutil.ReadFile(path))

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
