package backend_test

import (
	"testing"
	. "swag/backend"

	"github.com/stretchr/testify/assert"
)

func TestClaimAndConfirmPaper(t *testing.T) {
	submitter.Submit(&p)

	submitter.ClaimPaper()

	Pc.ConfirmOwnership(&submitter)

	// Asserting
	_, ClaimPC := GetConfirmMessage(&submitter)
	_, ClaimSubmitter := GetClaimMessage(&submitter)

	assert.Equal(t, ClaimPC, ClaimSubmitter, "failz")
}
