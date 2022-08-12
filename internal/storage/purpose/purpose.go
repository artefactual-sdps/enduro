package purpose

import (
	"database/sql/driver"
	"encoding/json"
	"strings"
)

type LocationPurpose uint

const (
	LocationPurposeUnspecified LocationPurpose = iota
	LocationPurposeAIPStore
)

func NewLocationPurpose(purpose string) LocationPurpose {
	var p LocationPurpose

	switch strings.ToLower(purpose) {
	case "aip_store":
		p = LocationPurposeAIPStore
	default:
		p = LocationPurposeUnspecified
	}

	return p
}

func (p LocationPurpose) String() string {
	switch p {
	case LocationPurposeAIPStore:
		return "aip_store"
	}
	return "unspecified"
}

func (p LocationPurpose) Values() []string {
	return []string{
		LocationPurposeUnspecified.String(),
		LocationPurposeAIPStore.String(),
	}
}

// Value provides the DB a string from int.
func (p LocationPurpose) Value() (driver.Value, error) {
	return p.String(), nil
}

// Scan tells our code how to read the enum into our type.
func (p *LocationPurpose) Scan(val interface{}) error {
	var s string
	switch v := val.(type) {
	case nil:
		return nil
	case string:
		s = v
	case []uint8:
		s = string(v)
	}

	*p = NewLocationPurpose(s)

	return nil
}

func (p LocationPurpose) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.String())
}

func (p *LocationPurpose) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	*p = NewLocationPurpose(s)

	return nil
}
