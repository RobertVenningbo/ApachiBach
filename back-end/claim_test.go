package backend

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClaimAndConfirmPaper(t *testing.T) {
	submitter.Submit(&p)

	submitter.ClaimPaper()

	pc.ConfirmOwnership(&submitter)

	// Asserting
	_, ClaimPC := GetConfirmMessage(&submitter)
	_, ClaimSubmitter := GetClaimMessage(&submitter)

	assert.Equal(t, ClaimPC, ClaimSubmitter, "failz")
}
