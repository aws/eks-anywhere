# Issues

## Labels
* Each issue (with the exception of the ones in the `oncall` milestone) should have at least:
	* A label identifying the team responsible for it:
		* `team/cli`
		* `team/providers`
		* `team/packages`
	* A label identifying the area of work. Examples:
		* `area/providers/capc`
		* `area/controller`
		* `area/e2e`
		* `area/release`
		* `area/supportbundle`
	* A label identifying the kind of issue:
		* `kind/bug`
		* `kind/enhancement` (new feature or request)
		* `kind/cleanup` (tech debt, refactors, etc.)

* If the issue is coming from someone outside the internal eks-a teams, the issue should have the `external` label.

## Milestones

Each issue should belong to one of these milestones:
* `next`: work in progress issues and/or planned for the next immediate release
* `next+1`: issues planned for the next release after the one currently in progress
* `next-patch`: work in progress issues and/or planned for the next patch release
* `backlog`: issues that haven't being allocated for a release yet
* `oncall`: external issues currently being triaged. In order to move them out of this milestone, all the necessary labels need to be added first (team, area, kind, etc.).
