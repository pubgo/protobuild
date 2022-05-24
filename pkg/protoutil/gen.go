package protoutil

import (
	"bytes"
	"strings"
	"unicode"

	pongo "github.com/flosch/pongo2/v5"
)

// Camel2Case
// 驼峰式写法转为下划线写法
func Camel2Case(name string) string {
	name = trim(name)
	buf := new(bytes.Buffer)
	for i, r := range name {
		if !unicode.IsUpper(r) {
			buf.WriteRune(r)
			continue
		}

		if i != 0 {
			buf.WriteRune('-')
		}
		buf.WriteRune(unicode.ToLower(r))
	}
	return strings.NewReplacer(".", "-", "_", "-", "--", "-").Replace(buf.String())
}

func trim(s string) string {
	return strings.Trim(strings.TrimSpace(s), ".-_/")
}

type Context = pongo.Context
