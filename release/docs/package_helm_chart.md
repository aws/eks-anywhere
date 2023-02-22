### Package Helm Chart Overview in Release
This is an overview of the changes that happen in the EKSA bundle and the package helm chart during each stage.

#### Dev
Dev where the branch is `main` undergoes the most changes. There is a function that checks for the latest dev tags of Helm Charts, and Images in Private ECR of the source account. The Dev tagged artifacts are `0.0.0` and `v0.0.0` for Helm, and for images. 

The next thing that occurs is a check for the existance of these artifacts in Public ECR or the Destination registry, and if they don't exist they are copied to that Destination registry.

Next a function will kick off to replace the image tags in the latest dev Helm Chart `values.yaml` for the controller, and token refresher with the images which will be published to the EKSA Bundle. This new modified helm chart will be pushed to Public ECR now, alongside the matching Images from the previous steps.

Then the Artifact Table image URI for the package controller helm chart, and images is overridden with the tags from the published dev artifacts in the previous steps.

#### Dev Release

Dev where the branch is NOT `main` no changes occur.

#### Staging and Production

Staging and Production both undergo a fix for the `Chart.yaml`. This happens because when the helm chart is built, the `Chart.yaml` is published with the original build tags which is a GIT_TAG and GIT_SHA. When the helm chart is moved to staging, and production it gets new a tag from the release process.

There is a fix which will override the version in the `Chart.yaml` to the destination artifact URI tag, so that the version in the `Chart.yaml` matches with the public registry tag of the helm chart.
