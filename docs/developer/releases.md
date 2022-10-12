# EKS-A releases

## Patch releases

### What should go in a patch release

* Fixes for bugs that break the application flow. Fixes should be **low risk**.

### What should not go in a patch release

* New features
* API changes (unless required to fix a critical bug)
* High risk bug fixes (talk to the team if an exception is required)
* Backwards incompatible changes
	* Example: pre v1, we can potentially remove API fields as long they have been deprecated for several minor releases and we offer some kind of transition, potentially automated. Removing and/or deprecating fields should always happen in minor releases.

### Adding changes to patch release

1. Create a PR with the fix against `main` and get it reviewed and approved
2. Use the `/cherry-pick` command in GitHub to create a backport PR against the release branch. Assuming you want to backport to the next patch for v0.12, comment this in the original PR:
```
/cherry-pick release-0.12
```
3. Merge the backport PR

When a bugfix spans secondary changes (cleanup, refactor, etc.), try to split them from the actual fix. Try to minimize the changes in the first PR that you will backport and just follow up with another one against `main` on top of the fix.
