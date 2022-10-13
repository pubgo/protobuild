package ormpb

import (
	"bytes"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
)

// TimestampJSONMarshaler describes the default jsonpb.Marshaler used by all
// instances of Timestamp. This struct is safe to replace or modify but
// should not be done so concurrently.
var TimestampJSONMarshaler = new(jsonpb.Marshaler)

// MarshalJSON satisfies the encoding/json Marshaler interface. This method
// uses the more correct jsonpb package to correctly marshal the message.
func (m *Timestamp) MarshalJSON() ([]byte, error) {
	if m == nil {
		return json.Marshal(nil)
	}
	buf := &bytes.Buffer{}
	if err := TimestampJSONMarshaler.Marshal(buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ json.Marshaler = (*Timestamp)(nil)

// TimestampJSONUnmarshaler describes the default jsonpb.Unmarshaler used by all
// instances of Timestamp. This struct is safe to replace or modify but
// should not be done so concurrently.
var TimestampJSONUnmarshaler = new(jsonpb.Unmarshaler)

// UnmarshalJSON satisfies the encoding/json Unmarshaler interface. This method
// uses the more correct jsonpb package to correctly unmarshal the message.
func (m *Timestamp) UnmarshalJSON(b []byte) error {
	return TimestampJSONUnmarshaler.Unmarshal(bytes.NewReader(b), m)
}

var _ json.Unmarshaler = (*Timestamp)(nil)

// ProtobufJSONMarshaler describes the default jsonpb.Marshaler used by all
// instances of Protobuf. This struct is safe to replace or modify but
// should not be done so concurrently.
var ProtobufJSONMarshaler = new(jsonpb.Marshaler)

// MarshalJSON satisfies the encoding/json Marshaler interface. This method
// uses the more correct jsonpb package to correctly marshal the message.
func (m *Protobuf) MarshalJSON() ([]byte, error) {
	if m == nil {
		return json.Marshal(nil)
	}
	buf := &bytes.Buffer{}
	if err := ProtobufJSONMarshaler.Marshal(buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ json.Marshaler = (*Protobuf)(nil)

// ProtobufJSONUnmarshaler describes the default jsonpb.Unmarshaler used by all
// instances of Protobuf. This struct is safe to replace or modify but
// should not be done so concurrently.
var ProtobufJSONUnmarshaler = new(jsonpb.Unmarshaler)

// UnmarshalJSON satisfies the encoding/json Unmarshaler interface. This method
// uses the more correct jsonpb package to correctly unmarshal the message.
func (m *Protobuf) UnmarshalJSON(b []byte) error {
	return ProtobufJSONUnmarshaler.Unmarshal(bytes.NewReader(b), m)
}

var _ json.Unmarshaler = (*Protobuf)(nil)
