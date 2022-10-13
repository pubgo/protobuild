package retagpb

import (
	"bytes"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
)

// TagJSONMarshaler describes the default jsonpb.Marshaler used by all
// instances of Tag. This struct is safe to replace or modify but
// should not be done so concurrently.
var TagJSONMarshaler = new(jsonpb.Marshaler)

// MarshalJSON satisfies the encoding/json Marshaler interface. This method
// uses the more correct jsonpb package to correctly marshal the message.
func (m *Tag) MarshalJSON() ([]byte, error) {
	if m == nil {
		return json.Marshal(nil)
	}
	buf := &bytes.Buffer{}
	if err := TagJSONMarshaler.Marshal(buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ json.Marshaler = (*Tag)(nil)

// TagJSONUnmarshaler describes the default jsonpb.Unmarshaler used by all
// instances of Tag. This struct is safe to replace or modify but
// should not be done so concurrently.
var TagJSONUnmarshaler = new(jsonpb.Unmarshaler)

// UnmarshalJSON satisfies the encoding/json Unmarshaler interface. This method
// uses the more correct jsonpb package to correctly unmarshal the message.
func (m *Tag) UnmarshalJSON(b []byte) error {
	return TagJSONUnmarshaler.Unmarshal(bytes.NewReader(b), m)
}

var _ json.Unmarshaler = (*Tag)(nil)
