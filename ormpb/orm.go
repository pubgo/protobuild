package ormpb

import (
	"database/sql"
	"database/sql/driver"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func New(t time.Time) *Timestamp {
	return &Timestamp{Timestamp: timestamppb.New(t)}
}

func Now() *Timestamp {
	return &Timestamp{Timestamp: timestamppb.Now()}
}

func (x *Timestamp) Scan(src interface{}) error {
	nullTime := &sql.NullTime{}
	if err := nullTime.Scan(src); err != nil {
		return err
	}

	x.Timestamp = timestamppb.New(nullTime.Time)
	return nil
}

// Value implements driver.Valuer interface and returns string format of Time.
func (x *Timestamp) Value() (driver.Value, error) {
	return x.Timestamp.AsTime(), nil
}

// MarshalJSON implements json.Marshaler to convert Time to json serialization.
func (x *Timestamp) MarshalJSON() ([]byte, error) {
	return x.Timestamp.AsTime().MarshalJSON()
}

// UnmarshalJSON implements json.Unmarshaler to deserialize json data.
func (x *Timestamp) UnmarshalJSON(data []byte) error {
	// ignore null
	if string(data) == "null" {
		return nil
	}

	var t = time.Time{}
	if err := t.UnmarshalJSON(data); err != nil {
		return err
	}

	x.Timestamp = timestamppb.New(t)
	return nil
}
