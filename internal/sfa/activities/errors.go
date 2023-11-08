package activities

import "errors"

var (
	ErrIlegalFileFormat   = errors.New("Ilegal file format found")
	ErrInvaliSipStructure = errors.New("Invalid SIP structure")
)
