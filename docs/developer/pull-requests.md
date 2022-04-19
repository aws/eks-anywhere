# Pull Request guidelines
This document outlines best practices to follow when creating and reviewing PRs. Take them as guidelines and not hard rules.

## As an author

### PR size

The recommended maximum number of LoC is 500. Split PRs as necessary and be guided by the [Single Responsibility Principle](https://blog.cleancoder.com/uncle-bob/2014/05/08/SingleReponsibilityPrinciple.html). Put yourself in the reviewers shoes and consider the cognitive load on the reviewer.

For example, a PR with 1000 LoC touching 20 files could be easy to review if the changes are grammar or could be difficult to review if there are functional changes in loosely related components.

### PR descriptions

This section applies to the GitHub PR description and not to the commit messages.

Help the reviewer build context asynchronously with a well structured PR description. A good description can serve as documentation for author and reviewer when tracking down changes in the future and for outside contributors to get a feel for what a change is in regard to.

A good PR description should include:
1. A summary of intent including the problem you're solving. If the summary is fulfilled by a series of PRs, and not entirely by this PR, the summary should indicate that.
2. Dependent on the PR, a slightly more detailed description about what you're specifically addressing in said PR.
3. What's next? Tell the reviewers what your future intentions are (if it applies).
4. Links to GitHub issues, design docs, documentation, etc, as needed.
Here are some recommendations to follow while writing the description:

* Follow the Single Responsibility Principle. If you writing a PR description and you realize that you are describing multiple "responsibilities", that's a good sign you should probably split it.
* Assume the reviewer knows nothing about your work.
* Try to look at your PR from a reviewer perspective: what would you need to be able to review that PR successfully? Common topics:
    * What problem are we solving
    * Why are we solving it
    * How are we doing it
* Be explicit. Even if you think certain information can be inferred from the code, don’t rely on that. Specify everything that is not obvious.
* If they exist, link related GitHub issue/s.
* If they exist (and they are public), link design docs.
* Reflect all the information that is not documented anywhere else, including:
    * Decisions made in other PRs, issues, and private conversations
    * New requirements
    * How does this PR fit the whole project or feature? This is very important if you are splitting something bigger into small PRs. It helps the reviewer fill the gaps and look at your high-level solution without having all the code in front of them.

`TODO: add concrete example for a good PR description`

### Review your own changes

Reviewing your own changes increases code quality and streamlines reviews. Add comments to your own PR to notify reviewers that you've identified the problem and they don't need to comment. Draw attention from reviewers on areas of code you're unsure about and ask for additional input.

### Design docs

Not all PRs need a design doc. But there is an important difference between a design doc and a well written PR description: the design doc is reviewed before the code is written and the PR description after that.

Write design docs when you or the team wants consensus on an approach, solution, or idea before implementing it.

Not all design docs need to be big. Sometimes a paragraph is enough. But even that can help setting a common understanding and context that the author and reviewers can benefit from.

You can find existing EKS Anywhere design docs [here](https://github.com/aws/eks-anywhere/tree/main/designs).
### PR assignment

Most of the time, there are other folks involved in the project/feature you are working on, even if it was only during the design phase. Add all of them. If no one else has been involved or your PR is just a small fix/random improvement you came up with, use the GitHub suggested reviewers. If you are in doubt, ping the team in Slack and ask for help.

### Review comments

In general, all comments are important and should be addressed and/or acknowledged before the PR is merged (emoji reaction, comments, etc.).

If you are planning on not addressing one, talk to the reviewer who wrote it and let them know.

If you are requested to make a change that you would like for some reason to not add to the PR and do it later, talk to the reviewer. But, please don’t use it as an excuse to skip tests. And create issues for it. Checklists in parent issues are super useful for this scenario.

## As a reviewer

### Time to first review SLO

The review target is to respond within one US business day from the time the PR is open to the time the first review is submitted.

If you think you won't able to meet this, talk to the author.

### PR size

If you think the PR is too big and/or could be split in smaller chunks in order to make the review easier, talk to the author. It is acceptable to ask for a PR to be split before moving forward with the review.

### PR assignment

If you don’t think you should be in a particular PR or you don’t think you will have the time to do it, talk to the author so they can find a replacement if needed. Don’t let it sit. If you are assigned, you are still responsible for reviewing it. The GitHub notifications section is very useful to check if you are assigned to a PR.

### Review comments

Making your intent as a reviewer more clear helps the author decide on the next steps, prioritizing their work and avoid future back and forth:

* Whenever it’s something you don’t need to be changed in order to approve the PR, prepend it with “Nit:”. It’s fast, short and everyone understand it. If it’s not necessarily a suggested change and it’s purely educational, make that clear as well. When possible, use the "Add a suggestion" feature. This makes it easier to deal with spelling, one-time-use variable naming, grammar, small logical adjustments, etc.
* If you are OK with resolving one or your comments in a future PR, let the author know. This one is a bit tricky because in general, we don’t want comments to be left for later, if we write them is for a reason. But it’s true that there are times where what we ask for is too big and/or complex and will make more damage to the PR than good to the codebase. If that’s the case and there is no risk implementing it later, let the author know.
    * In general, all the code should be merged with unit tests. Separating the code from its tests in different PRs is not advisable, since there will be un undetermined period of time while a part of the system remains untested.

### Light reviews

It’s totally OK to take a look to a PR and leave comments even if you don’t have the time or the confidence to approve it. More people looking at the code is always better. Even if your comments are only questions. Worse case scenario, the author has an answer and you learn something. Best case scenario, the reviewer doesn’t have an answer so you both learn something and the code ends up being better.

If that’s case, make it clear so the author doesn't keep waiting for your approval.

### Labels used to approve

Our CI blocks the merge of PRs unless both labels are specified (without a `do-not-merge` label):

`lgtm` - Used when a trusted reviewer thinks no other changes need to be made in order to merge the PR. This can be added to a
PR by specifying `/lgtm` as a comment or approving the PR using the Github UI.

`approve` - Used when an approver deems that the PR can be automatically merged. This can be added to a PR by specifying `/approve`
as a comment. Please note the `PRs authored as an approver` section below on specifics on how this label can be used.


### When to approve

These are some questions to ask oneself that can useful in this process:

* Do I understand what the problem is?
* Do I understand why are we solving the problem?
* Do I think we should be solving this problem? This one can be a bit tricky. This should probably have been brought up during release, feature, or whatever planning. If it was, and the team decided to do it, move on. The answer is already yes. If it wasn’t, be careful with how you approach it but question it.
* Do I understand what the proposed solution is?
* Does the proposed solution solve the problem?
* Does the code do what is supposed to do (based on the proposed solution)?

Once the answer to *all those questions is yes*, you can move to the next round:

* Is there a better solution? Was that solution considered? If it was, do I understand why it wasn’t selected?
* Do I agree with the selection of the solution? Can I prove that my option is better with data or referring to solid engineering principles? Or is this purely preference based?

Note: *solution* here is used quite generically and doesn't necessarily refer only to high level design. It includes code structure, code patterns, abstractions, and so on. When there is a design doc, the high-level solution should already have been agreed upon before opening the PR. But we should still ask these questions about the solution's implementation.

Whether or not you push back, when the answer to these questions is no, is up to you. It requires a balance between risk, helping the team move forward and the importance of the suggested changes. In general, we recommend:

* Don’t stop asking questions until you understand the decision being made, even if you disagree. Try to find a balance between creating too much back and forth in the PR comments and solving everything on private discussions. When a decision is made in a private conversation, summarize it in a PR comment.
* Don’t block your team only because you have a different personal preference.

And finally, you can move forward to the code itself:

* Is the code making the codebase better? If the answer is no, don’t approve it. But remember that it doesn’t need to be perfect to be better. When is the code better? We’ll leave that up to you to answer. When you do, write a doc please.
* Is the code properly tested or is there a plan in place to test it properly? If the answer is no, don’t approve it.
* When necessary, encourage folks to create follow-up tickets, so feedback and improvement ideas don't get lost.

If you get here, you are done. Approve it.

### PRs authored as an approver

If you are on the approvers list, you may add an `/approve` comment if you want to merge the PR with just `/lgtm` from
another reviewer. This allows the author who is also an approver to control who they want the PR to be reviewed by 
instead of having another reviewer decide that.
