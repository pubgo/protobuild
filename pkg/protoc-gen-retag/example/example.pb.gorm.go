package example

import (
	"bytes"
	"encoding/json"
	"github.com/golang/protobuf/jsonpb"
)

// ExampleJSONMarshaler describes the default jsonpb.Marshaler used by all
// instances of Example. This struct is safe to replace or modify but
// should not be done so concurrently.
var ExampleJSONMarshaler = new(jsonpb.Marshaler)

// MarshalJSON satisfies the encoding/json Marshaler interface. This method
// uses the more correct jsonpb package to correctly marshal the message.
func (m *Example) MarshalJSON() ([]byte, error) {
	if m == nil {
		return json.Marshal(nil)
	}
	buf := &bytes.Buffer{}
	if err := ExampleJSONMarshaler.Marshal(buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ json.Marshaler = (*Example)(nil)

// ExampleJSONUnmarshaler describes the default jsonpb.Unmarshaler used by all
// instances of Example. This struct is safe to replace or modify but
// should not be done so concurrently.
var ExampleJSONUnmarshaler = new(jsonpb.Unmarshaler)

// UnmarshalJSON satisfies the encoding/json Unmarshaler interface. This method
// uses the more correct jsonpb package to correctly unmarshal the message.
func (m *Example) UnmarshalJSON(b []byte) error {
	return ExampleJSONUnmarshaler.Unmarshal(bytes.NewReader(b), m)
}

var _ json.Unmarshaler = (*Example)(nil)

// SecondMessageJSONMarshaler describes the default jsonpb.Marshaler used by all
// instances of SecondMessage. This struct is safe to replace or modify but
// should not be done so concurrently.
var SecondMessageJSONMarshaler = new(jsonpb.Marshaler)

// MarshalJSON satisfies the encoding/json Marshaler interface. This method
// uses the more correct jsonpb package to correctly marshal the message.
func (m *SecondMessage) MarshalJSON() ([]byte, error) {
	if m == nil {
		return json.Marshal(nil)
	}
	buf := &bytes.Buffer{}
	if err := SecondMessageJSONMarshaler.Marshal(buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ json.Marshaler = (*SecondMessage)(nil)

// SecondMessageJSONUnmarshaler describes the default jsonpb.Unmarshaler used by all
// instances of SecondMessage. This struct is safe to replace or modify but
// should not be done so concurrently.
var SecondMessageJSONUnmarshaler = new(jsonpb.Unmarshaler)

// UnmarshalJSON satisfies the encoding/json Unmarshaler interface. This method
// uses the more correct jsonpb package to correctly unmarshal the message.
func (m *SecondMessage) UnmarshalJSON(b []byte) error {
	return SecondMessageJSONUnmarshaler.Unmarshal(bytes.NewReader(b), m)
}

var _ json.Unmarshaler = (*SecondMessage)(nil)

// ThirdExampleJSONMarshaler describes the default jsonpb.Marshaler used by all
// instances of ThirdExample. This struct is safe to replace or modify but
// should not be done so concurrently.
var ThirdExampleJSONMarshaler = new(jsonpb.Marshaler)

// MarshalJSON satisfies the encoding/json Marshaler interface. This method
// uses the more correct jsonpb package to correctly marshal the message.
func (m *ThirdExample) MarshalJSON() ([]byte, error) {
	if m == nil {
		return json.Marshal(nil)
	}
	buf := &bytes.Buffer{}
	if err := ThirdExampleJSONMarshaler.Marshal(buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ json.Marshaler = (*ThirdExample)(nil)

// ThirdExampleJSONUnmarshaler describes the default jsonpb.Unmarshaler used by all
// instances of ThirdExample. This struct is safe to replace or modify but
// should not be done so concurrently.
var ThirdExampleJSONUnmarshaler = new(jsonpb.Unmarshaler)

// UnmarshalJSON satisfies the encoding/json Unmarshaler interface. This method
// uses the more correct jsonpb package to correctly unmarshal the message.
func (m *ThirdExample) UnmarshalJSON(b []byte) error {
	return ThirdExampleJSONUnmarshaler.Unmarshal(bytes.NewReader(b), m)
}

var _ json.Unmarshaler = (*ThirdExample)(nil)

// ThirdExample_InnerExampleJSONMarshaler describes the default jsonpb.Marshaler used by all
// instances of ThirdExample_InnerExample. This struct is safe to replace or modify but
// should not be done so concurrently.
var ThirdExample_InnerExampleJSONMarshaler = new(jsonpb.Marshaler)

// MarshalJSON satisfies the encoding/json Marshaler interface. This method
// uses the more correct jsonpb package to correctly marshal the message.
func (m *ThirdExample_InnerExample) MarshalJSON() ([]byte, error) {
	if m == nil {
		return json.Marshal(nil)
	}
	buf := &bytes.Buffer{}
	if err := ThirdExample_InnerExampleJSONMarshaler.Marshal(buf, m); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

var _ json.Marshaler = (*ThirdExample_InnerExample)(nil)

// ThirdExample_InnerExampleJSONUnmarshaler describes the default jsonpb.Unmarshaler used by all
// instances of ThirdExample_InnerExample. This struct is safe to replace or modify but
// should not be done so concurrently.
var ThirdExample_InnerExampleJSONUnmarshaler = new(jsonpb.Unmarshaler)

// UnmarshalJSON satisfies the encoding/json Unmarshaler interface. This method
// uses the more correct jsonpb package to correctly unmarshal the message.
func (m *ThirdExample_InnerExample) UnmarshalJSON(b []byte) error {
	return ThirdExample_InnerExampleJSONUnmarshaler.Unmarshal(bytes.NewReader(b), m)
}

var _ json.Unmarshaler = (*ThirdExample_InnerExample)(nil)
