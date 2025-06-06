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
	// Failed due to a system error.
	SIPStatusError SIPStatus = "error"
	// Failed due to invalid contents.
	SIPStatusFailed SIPStatus = "failed"
	// Awaiting resource allocation.
	SIPStatusQueued SIPStatus = "queued"
	// Undergoing work.
	SIPStatusProcessing SIPStatus = "processing"
	// Awaiting user decision.
	SIPStatusPending SIPStatus = "pending"
	// Successfully ingested.
	SIPStatusIngested SIPStatus = "ingested"
)

var ErrInvalidSIPStatus = fmt.Errorf("not a valid SIPStatus, try [%s]", strings.Join(_SIPStatusNames, ", "))

var _SIPStatusNames = []string{
	string(SIPStatusError),
	string(SIPStatusFailed),
	string(SIPStatusQueued),
	string(SIPStatusProcessing),
	string(SIPStatusPending),
	string(SIPStatusIngested),
}

// SIPStatusNames returns a list of possible string values of SIPStatus.
func SIPStatusNames() []string {
	tmp := make([]string, len(_SIPStatusNames))
	copy(tmp, _SIPStatusNames)
	return tmp
}

// String implements the Stringer interface.
func (x SIPStatus) String() string {
	return string(x)
}

// IsValid provides a quick way to determine if the typed value is
// part of the allowed enumerated values
func (x SIPStatus) IsValid() bool {
	_, err := ParseSIPStatus(string(x))
	return err == nil
}

var _SIPStatusValue = map[string]SIPStatus{
	"error":      SIPStatusError,
	"failed":     SIPStatusFailed,
	"queued":     SIPStatusQueued,
	"processing": SIPStatusProcessing,
	"pending":    SIPStatusPending,
	"ingested":   SIPStatusIngested,
}

// ParseSIPStatus attempts to convert a string to a SIPStatus.
func ParseSIPStatus(name string) (SIPStatus, error) {
	if x, ok := _SIPStatusValue[name]; ok {
		return x, nil
	}
	return SIPStatus(""), fmt.Errorf("%s is %w", name, ErrInvalidSIPStatus)
}

// Values implements the entgo.io/ent/schema/field EnumValues interface.
func (x SIPStatus) Values() []string {
	return SIPStatusNames()
}

// SIPStatusInterfaces returns an interface list of possible values of SIPStatus.
func SIPStatusInterfaces() []interface{} {
	var tmp []interface{}
	for _, v := range _SIPStatusNames {
		tmp = append(tmp, v)
	}
	return tmp
}

// ParseSIPStatusWithDefault attempts to convert a string to a ContentType.
// It returns the default value if name is empty.
func ParseSIPStatusWithDefault(name string) (SIPStatus, error) {
	if name == "" {
		return _SIPStatusValue[_SIPStatusNames[0]], nil
	}
	if x, ok := _SIPStatusValue[name]; ok {
		return x, nil
	}
	var e SIPStatus
	return e, fmt.Errorf("%s is not a valid SIPStatus, try [%s]", name, strings.Join(_SIPStatusNames, ", "))
}

// NormalizeSIPStatus attempts to parse a and normalize string as content type.
// It returns the input untouched if name fails to be parsed.
// Example:
//
//	"enUM" will be normalized (if possible) to "Enum"
func NormalizeSIPStatus(name string) string {
	res, err := ParseSIPStatus(name)
	if err != nil {
		return name
	}
	return res.String()
}
