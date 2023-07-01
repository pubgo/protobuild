package version

import (
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/assert"
)

var Version = "v0.0.1-dev"

func init() {
	if Version == "" {
		Version = "v0.0.1-dev"
	}

	if _, err := ver.NewVersion(Version); err != nil {
		assert.Exit(err, Version)
	}
}
