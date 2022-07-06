package utils

import (
	"go/format"
	"os"
	"strings"

	"github.com/pubgo/funk"
)

func CodeFormat(data ...string) string {
	var str = ""
	for i := range data {
		str += strings.TrimSpace(data[i]) + "\n"
	}
	str = strings.TrimSpace(str)
	return string(funk.Must1(format.Source([]byte(str))))
}

func DotJoin(str ...string) string {
	return strings.Join(str, ".")
}

// DirExists function to check if directory exists?
func DirExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && stat.IsDir() {
		// path is a directory
		return true
	}
	return false
}

// FileExists function to check if file exists?
func FileExists(path string) bool {
	if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
		return true
	} else if os.IsNotExist(err) {
		return false
	} else {
		return false
	}
}

func FirstFnNotEmpty(fx ...func() string) string {
	for i := range fx {
		if s := fx[i](); s != "" {
			return s
		}
	}
	return ""
}

func FirstNotEmpty(strs ...string) string {
	for i := range strs {
		if s := strs[i]; s != "" {
			return s
		}
	}
	return ""
}

func IfEmpty(str string, fx func()) {
	if str == "" {
		fx()
	}
}
