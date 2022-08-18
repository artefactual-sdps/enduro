package types

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

type LocationSource uint

const (
	LocationSourceUnspecified LocationSource = iota
	LocationSourceMinIO
)

func NewLocationSource(source string) LocationSource {
	var s LocationSource

	switch strings.ToLower(source) {
	case "minio":
		s = LocationSourceMinIO
	default:
		s = LocationSourceUnspecified
	}

	return s
}

func (s LocationSource) String() string {
	switch s {
	case LocationSourceMinIO:
		return "minio"
	}
	return "unspecified"
}

func (s LocationSource) Values() []string {
	return []string{
		LocationSourceUnspecified.String(),
		LocationSourceMinIO.String(),
	}
}

// Value provides the DB a string from int.
func (s LocationSource) Value() (driver.Value, error) {
	return s.String(), nil
}

// Scan tells our code how to read the enum into our type.
func (s *LocationSource) Scan(val interface{}) error {
	var str string
	switch v := val.(type) {
	case nil:
		return nil
	case string:
		str = v
	case []uint8:
		str = string(v)
	}

	*s = NewLocationSource(str)

	return nil
}

func (s LocationSource) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.String())
}

func (s *LocationSource) UnmarshalJSON(b []byte) error {
	var str string
	if err := json.Unmarshal(b, &str); err != nil {
		return err
	}

	*s = NewLocationSource(str)

	return nil
}
