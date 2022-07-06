package backend_test

import (
	"testing"
	. "swag/backend"

	"github.com/stretchr/testify/assert"
)

func TestClaimAndConfirmPaper(t *testing.T) {
	submitter.Submit(&p)

	submitter.ClaimPaper(submitter.PaperCommittedValue.Paper.Id)

	Pc.ConfirmOwnership(submitter.PaperCommittedValue.Paper.Id)

	// Asserting
	_, ClaimPC := GetConfirmMessage(submitter.PaperCommittedValue.Paper.Id)
	_, ClaimSubmitter := GetClaimMessage(submitter.PaperCommittedValue.Paper.Id)

	assert.Equal(t, ClaimPC, ClaimSubmitter, "failz")
}
