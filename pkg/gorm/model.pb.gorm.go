package ormpb

import (
	"bytes"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
)

// ModelJSONMarshaler describes the default jsonpb.Marshaler used by all
// instances of Model. This struct is safe to replace or modify but
// should not be done so concurrently.
var ModelJSONMarshaler = new(jsonpb.Marshaler)

// MarshalJSON satisfies the encoding/json Marshaler interface. This method
// uses the more correct jsonpb package to correctly marshal the message.
func (m *Model) MarshalJSON() ([]byte, error) {
	if m == nil {
		return json.Marshal(nil)
	}
	buf := &bytes.Buffer{}
	if err := ModelJSONMarshaler.Marshal(buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ json.Marshaler = (*Model)(nil)

// ModelJSONUnmarshaler describes the default jsonpb.Unmarshaler used by all
// instances of Model. This struct is safe to replace or modify but
// should not be done so concurrently.
var ModelJSONUnmarshaler = new(jsonpb.Unmarshaler)

// UnmarshalJSON satisfies the encoding/json Unmarshaler interface. This method
// uses the more correct jsonpb package to correctly unmarshal the message.
func (m *Model) UnmarshalJSON(b []byte) error {
	return ModelJSONUnmarshaler.Unmarshal(bytes.NewReader(b), m)
}

var _ json.Unmarshaler = (*Model)(nil)
