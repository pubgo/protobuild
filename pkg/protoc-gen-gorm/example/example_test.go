package example

import (
	"encoding/json"
	"fmt"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/logx"
	"google.golang.org/protobuf/types/known/timestamppb"
	"testing"
	"time"
)

func TestName(t *testing.T) {
	var pb = &Example{WithNewTags: "true", Test_5: timestamppb.New(time.Now().UTC())}
	var dd, err = json.Marshal(pb)
	assert.Must(err)
	// 2022-10-19T03:49:42.240649Z
	logx.Info(string(dd))

	mm, _ := json.Marshal(pb.Test_5.AsTime())
	logx.Info(string(mm))

	d2, err := json.Marshal(pb.ToModel())
	assert.Must(err)
	// 2022-10-19T03:49:42.240649Z
	logx.Info(string(d2))

	var pb1 *Example
	assert.Must(json.Unmarshal(dd, &pb1))
	fmt.Println(pb1.Test_5.AsTime().String())

	vvv, err := json.Marshal(pb1.ToModel())
	assert.Must(err)
	fmt.Println(string(vvv))

	t1, err := time.Parse(time.RFC3339, pb1.ToModel().Test_5.String())
	assert.Must(err)
	fmt.Println(t1.String())
}
