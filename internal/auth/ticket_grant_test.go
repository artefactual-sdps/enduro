package auth_test

import (
	"testing"

	"gotest.tools/v3/assert"

	"github.com/artefactual-sdps/enduro/internal/auth"
)

func TestTicketGrant(t *testing.T) {
	t.Parallel()

	grant := auth.NewTicketGrant(auth.TicketPurposeIngestSIPDownload, "resource")

	data, err := grant.MarshalBinary()
	assert.NilError(t, err)

	var decoded auth.TicketGrant
	assert.NilError(t, decoded.UnmarshalBinary(data))
	assert.DeepEqual(t, decoded, grant)

	assert.NilError(t, decoded.Validate(auth.TicketPurposeIngestSIPDownload, "resource"))
	assert.ErrorContains(
		t,
		decoded.Validate(auth.TicketPurposeStorageAIPDownload, "resource"),
		"ticket purpose mismatch",
	)
	assert.ErrorContains(
		t,
		decoded.Validate(auth.TicketPurposeIngestSIPDownload, "other-resource"),
		"ticket resource mismatch",
	)
}
