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
	SIPFailedAsSIP SIPFailedAs = "SIP"
	SIPFailedAsPIP SIPFailedAs = "PIP"
)

var ErrInvalidSIPFailedAs = fmt.Errorf("not a valid SIPFailedAs, try [%s]", strings.Join(_SIPFailedAsNames, ", "))

var _SIPFailedAsNames = []string{
	string(SIPFailedAsSIP),
	string(SIPFailedAsPIP),
}

// SIPFailedAsNames returns a list of possible string values of SIPFailedAs.
func SIPFailedAsNames() []string {
	tmp := make([]string, len(_SIPFailedAsNames))
	copy(tmp, _SIPFailedAsNames)
	return tmp
}

// String implements the Stringer interface.
func (x SIPFailedAs) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x SIPFailedAs) IsValid() bool {
	_, err := ParseSIPFailedAs(string(x))
	return err == nil
}

var _SIPFailedAsValue = map[string]SIPFailedAs{
	"SIP": SIPFailedAsSIP,
	"PIP": SIPFailedAsPIP,
}

// ParseSIPFailedAs attempts to convert a string to a SIPFailedAs.
func ParseSIPFailedAs(name string) (SIPFailedAs, error) {
	if x, ok := _SIPFailedAsValue[name]; ok {
		return x, nil
	}
	return SIPFailedAs(""), fmt.Errorf("%s is %w", name, ErrInvalidSIPFailedAs)
}

// Values implements the entgo.io/ent/schema/field EnumValues interface.
func (x SIPFailedAs) Values() []string {
	return SIPFailedAsNames()
}

// SIPFailedAsInterfaces returns an interface list of possible values of SIPFailedAs.
func SIPFailedAsInterfaces() []interface{} {
	var tmp []interface{}
	for _, v := range _SIPFailedAsNames {
		tmp = append(tmp, v)
	}
	return tmp
}

// ParseSIPFailedAsWithDefault attempts to convert a string to a ContentType.
// It returns the default value if name is empty.
func ParseSIPFailedAsWithDefault(name string) (SIPFailedAs, error) {
	if name == "" {
		return _SIPFailedAsValue[_SIPFailedAsNames[0]], nil
	}
	if x, ok := _SIPFailedAsValue[name]; ok {
		return x, nil
	}
	var e SIPFailedAs
	return e, fmt.Errorf("%s is not a valid SIPFailedAs, try [%s]", name, strings.Join(_SIPFailedAsNames, ", "))
}

// NormalizeSIPFailedAs attempts to parse a and normalize string as content type.
// It returns the input untouched if name fails to be parsed.
// Example:
//
//	"enUM" will be normalized (if possible) to "Enum"
func NormalizeSIPFailedAs(name string) string {
	res, err := ParseSIPFailedAs(name)
	if err != nil {
		return name
	}
	return res.String()
}
