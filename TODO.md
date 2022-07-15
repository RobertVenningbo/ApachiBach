- reviewers must continiue to bidding at the same time, easy fix.
- download in path "../makereview" doesn't work.
- make download popup too (should be some HTTP headers not being set)
- only accepted papers should be checked in the claim stage.
- pc needs to sign the last three messages of the submitters claim message
- make batch script for starting a batch of agents
- pc /checkreview doesnt display papers anymore after

1. Fix PostGradeDiscussingHandler/PublishAgreedGrade(), so we can have more than 1 reviewer (state 13 is never reached) --- WORKS
2. Test Decision and Claim with multiple submitters and reviewers (not possible until above is fixed) --- WORKS
3. Create unit tests for important parts of protocol (ZKP and Commitment already have working tests I think) 
4. Write report (preferably walkthrough of Set Membership ZKP, section 5.3.2.2, since you know all the ins and outs)