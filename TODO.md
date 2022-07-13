reviewers must continiue to bidding at the same time, easy fix.

1. Fix PostGradeDiscussingHandler/PublishAgreedGrade(), so we can have more than 1 reviewer (state 13 is never reached)
2. Test Decision and Claim with multiple submitters and reviewers (not possible until above is fixed)
3. Create unit tests for important parts of protocol (ZKP and Commitment already have working tests I think)
4. Write report (preferably walkthrough of Set Membership ZKP, section 5.3.2.2, since you know all the ins and outs)