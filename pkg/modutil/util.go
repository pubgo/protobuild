package modutil

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/pubgo/protobuild/pkg/utils"
	"github.com/pubgo/xerror"
	"golang.org/x/mod/modfile"
)

func getFileByRecursion(file string, path string) string {
	filePath := filepath.Join(path, file)
	if utils.FileExists(filePath) {
		return filePath
	}

	if path == string(os.PathSeparator) {
		return ""
	}

	return getFileByRecursion(file, filepath.Dir(path))
}

func GoModPath() string {
	var pwd = xerror.PanicStr(os.Getwd())
	return getFileByRecursion("go.mod", pwd)
}

func LoadVersions() map[string]string {
	var path = GoModPath()
	xerror.Assert(path == "", "go.mod not exists")

	var modBytes = xerror.PanicBytes(ioutil.ReadFile(path))

	var a, err = modfile.Parse("in", modBytes, nil)
	xerror.Panic(err, "go.mod 解析失败")

	var versions = make(map[string]string)

	for i := range a.Require {
		var mod = a.Require[i].Mod
		versions[mod.Path] = mod.Version
	}

	for i := range a.Replace {
		var mod = a.Replace[i].New
		versions[mod.Path] = mod.Version
	}

	return versions
}

func LoadMod() *modfile.File {
	var path = GoModPath()
	var modBytes = xerror.PanicBytes(ioutil.ReadFile(path))

	var a, err = modfile.Parse("in", modBytes, nil)
	xerror.Panic(err, "go.mod 解析失败")
	return a
}
