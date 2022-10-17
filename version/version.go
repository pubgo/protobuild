package version

import (
	ver "github.com/hashicorp/go-version"
	"github.com/pubgo/funk/assert"
)

var CommitID = ""
var BuildTime = ""
var Data = ""
var Domain = ""
var Version = "v0.0.1-dev"
var Tag = ""

func init() {
	if Version == "" {
		Version = "v0.0.1-dev"
	}

	if _, err := ver.NewVersion(Version); err != nil {
		assert.Exit(err, Version)
	}
}