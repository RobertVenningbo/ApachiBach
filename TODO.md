- download in path "../makereview" doesn't work.
- make download popup too (should be some HTTP headers not being set)
- only accepted papers should be checked in the claim stage.
- more work in claim stage needed (submitter needs different page if they got rejected, they shouldn't be able to claim)
- more testing with multiple reviewers for one paper (suspecious of "PublishAgreedGrade", the place we randomize grades) (I THINK I FIXED IT, PLS CHECK TY :) )

--bash script seems to introduce bugs? certain places, dunno why

3. Create unit tests for important parts of protocol (ZKP and Commitment already have working tests I think) 
4. Write report (preferably walkthrough of Set Membership ZKP, section 5.3.2.2, since you know all the ins and outs)