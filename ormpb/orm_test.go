package ormpb

import (
	"testing"

	"github.com/pubgo/xerror"
)

func TestName(t *testing.T) {
	var n = Now()
	var d, err = n.MarshalJSON()
	xerror.Panic(err)
	t.Log(string(d))

	var n1 = Now()
	xerror.Panic(n1.UnmarshalJSON(d))
	t.Log(n.Timestamp.AsTime())
	xerror.Assert(n.String() != n1.String(), "")
}
