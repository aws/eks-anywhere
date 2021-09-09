# EKS Anywhere Release Runbook

1. **Create EKS Anywhere Prod Release PR**
    * release/RELEASE (increment)
    * docs/contents/releases/$(cat release/RELEASE)/index.md
    * docs/contents/releases/$(cat release/RELEASE)/CHANGELOG.md
    * docs/contents/releases/$(cat release/RELEASE)/announcement.txt
    * docs/contents/index.md
    * README
1. **Tag Repository**: Hopefully, this step can be automated, but for now, tag the repository:
    * `git tag -a $(cat release/RELEASE) -m $(cat release/RELEASE)`
    * `git push upstream $(cat release/RELEASE)`
