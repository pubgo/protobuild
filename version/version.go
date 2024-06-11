package version

import (
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/assert"
)

var Version = "v0.0.20"

func init() {
	if Version == "" {
		Version = "v0.0.1-dev"
	}

	assert.Exit1(ver.NewVersion(Version))
}
