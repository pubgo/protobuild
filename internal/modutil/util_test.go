package modutil

import (
	"strings"
	"testing"

	mapset "github.com/deckarep/golang-set/v2"
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/pretty"
	"github.com/pubgo/funk/v2/result"
	"github.com/samber/lo"

	"github.com/pubgo/protobuild/internal/shutil"
)

func TestName(t *testing.T) {
	versions := LoadVersions()
	if versions == nil {
		t.Log("LoadVersions returned nil (go.mod may not exist)")
	}
	pretty.Println(versions)

	modList := strings.Split(result.Wrap(shutil.GoModGraph()).Must(), "\n")
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

	for k, v := range modMap {
		pretty.Println(k, maxVersion(v).String(), minVersion(v).String())
	}
}

func minVersion(versions []*ver.Version) *ver.Version {
	return lo.MaxBy(versions, func(a, b *ver.Version) bool { return !a.GreaterThan(b) })
}
