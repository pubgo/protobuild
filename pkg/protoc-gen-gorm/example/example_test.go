package example

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestName(t *testing.T) {
	pb := &Example{WithNewTags: "true", Test_5: timestamppb.New(time.Now().UTC())}
	dd, err := json.Marshal(pb)
	assert.Must(err)
	// 2022-10-19T03:49:42.240649Z
	log.Print(string(dd))

	mm, _ := json.Marshal(pb.Test_5.AsTime())
	log.Print(string(mm))

	d2, err := json.Marshal(pb.ToModel())
	assert.Must(err)
	// 2022-10-19T03:49:42.240649Z
	log.Print(string(d2))

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

func TestName2(t *testing.T) {
	req := new(AllSrvReq)
	req.Req = append(req.Req, &Req{Req: &Req_Req1{}})

	rsp := new(AllSrvRsp)
	for i := range rsp.Rsp {
		rsp.Rsp[i].GetRsp()
	}
}
