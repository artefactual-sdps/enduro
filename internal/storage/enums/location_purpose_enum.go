// Code generated by go-enum DO NOT EDIT.
// Version: 0.6.0
// Revision: 919e61c0174b91303753ee3898569a01abb32c97
// Build Date: 2023-12-18T15:54:43Z
// Built By: goreleaser

package enums

import (
	"fmt"
	"strings"
)

const (
	LocationPurposeUnspecified LocationPurpose = "unspecified"
	LocationPurposeAipStore    LocationPurpose = "aip_store"
)

var ErrInvalidLocationPurpose = fmt.Errorf("not a valid LocationPurpose, try [%s]", strings.Join(_LocationPurposeNames, ", "))

var _LocationPurposeNames = []string{
	string(LocationPurposeUnspecified),
	string(LocationPurposeAipStore),
}

// LocationPurposeNames returns a list of possible string values of LocationPurpose.
func LocationPurposeNames() []string {
	tmp := make([]string, len(_LocationPurposeNames))
	copy(tmp, _LocationPurposeNames)
	return tmp
}

// String implements the Stringer interface.
func (x LocationPurpose) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x LocationPurpose) IsValid() bool {
	_, err := ParseLocationPurpose(string(x))
	return err == nil
}

var _LocationPurposeValue = map[string]LocationPurpose{
	"unspecified": LocationPurposeUnspecified,
	"aip_store":   LocationPurposeAipStore,
}

// ParseLocationPurpose attempts to convert a string to a LocationPurpose.
func ParseLocationPurpose(name string) (LocationPurpose, error) {
	if x, ok := _LocationPurposeValue[name]; ok {
		return x, nil
	}
	return LocationPurpose(""), fmt.Errorf("%s is %w", name, ErrInvalidLocationPurpose)
}

// Values implements the entgo.io/ent/schema/field EnumValues interface.
func (x LocationPurpose) Values() []string {
	return LocationPurposeNames()
}

// LocationPurposeInterfaces returns an interface list of possible values of LocationPurpose.
func LocationPurposeInterfaces() []interface{} {
	var tmp []interface{}
	for _, v := range _LocationPurposeNames {
		tmp = append(tmp, v)
	}
	return tmp
}

// ParseLocationPurposeWithDefault attempts to convert a string to a ContentType.
// It returns the default value if name is empty.
func ParseLocationPurposeWithDefault(name string) (LocationPurpose, error) {
	if name == "" {
		return _LocationPurposeValue[_LocationPurposeNames[0]], nil
	}
	if x, ok := _LocationPurposeValue[name]; ok {
		return x, nil
	}
	var e LocationPurpose
	return e, fmt.Errorf("%s is not a valid LocationPurpose, try [%s]", name, strings.Join(_LocationPurposeNames, ", "))
}

// NormalizeLocationPurpose attempts to parse a and normalize string as content type.
// It returns the input untouched if name fails to be parsed.
// Example:
//
//	"enUM" will be normalized (if possible) to "Enum"
func NormalizeLocationPurpose(name string) string {
	res, err := ParseLocationPurpose(name)
	if err != nil {
		return name
	}
	return res.String()
}
