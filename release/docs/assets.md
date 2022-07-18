# Adding new assets to the bundle

Each component in the EKS-A bundle manifest is composed of a number of assets, which could be container images, archives, tarballs, YAML manifests, etc. For ease of operation, we generate a map of project name to artifacts corresponding to that project, and we call this the "artifacts table". Once assets have been populated into the table, they can be consumed in multiple places while building the bundles for these components, for example, the `kube-vip` project features in several provider bundles, but its artifacts need to be processed only once. Every project has been abstracted as a common struct of type `AssetConfig` with well-defined paramaters.

In order to add a new project to the artifacts table, all you need to do is define the appropriate values in an `AssetConfig` struct and the common processing logic will take care of generating the appropriate artifacts list and packaging it into consumable `Artifact` objects.

The AssetConfig struct has the following schema:

```Go
type AssetConfig struct {
	ProjectName                    string
	ProjectPath                    string
	Archives                       []*Archive
	Images                         []*Image
	ImageRepoPrefix                string
	ImageTagOptions                []string
	Manifests                      []*ManifestComponent
	NoGitTag                       bool
	HasReleaseBranches             bool
	HasSeparateTagPerReleaseBranch bool
	OnlyForDevRelease              bool
	AssignEksATag                  bool
}
```

The above struct and its fields are explained in depth in the table below. This struct can be easily extended to support more parameters.

<details>
<summary>Click to view/hide table</summary>

| Parameter | Description |
| :---: | :---: |
| **AssetConfig** |
| `ProjectName` | Name of the project. This is used as the key for the project's entry in the table |
| `ProjectPath` | Path of the project in the build-tooling repo. Usually takes the form `projects/<org>/<repo>` |
| `Archives` | List of archives corresponding to the project. These could be OS or kernel images, tarballs, etc. Refer below for `Archive` struct field expansion |
| `Images` | List of container images corresponding to the project. Refer below for `Image` struct field expansion |
| `ImageRepoPrefix` | Prefix for the image name in the ECR URI. Usually corresponds to the projects's GitHub org or repository name |
| `ImageTagOptions` | List of elements required to construct the image tag for this image. Can only contain `gitTag`, `projectPath`, `eksDReleaseChannel`, `eksDReleaseNumber` and/or `kubeVersion` |
| `Manifests` | List of manifests corresponding to the project. Refer below for `ManifestComponent` struct field expansion |
| `NoGitTag` | Denotes that this project does not have a git tag |
| `HasReleaseBranches` | Denotes that this project is release-branched |
| `HasSeparateTagPerReleaseBranch` | Denotes that this project has separate git tags for each release branch |
| `OnlyForDevRelease` | Denotes that this project's artifacts should be packaged only for dev release |
| `AssignEksATag` | Denotes that the tags for this project's images should align with EKS-A CLI tag |
| **Archive** |
| `Name` | Name of the archive |
| `Format` | Format of the archive, i.e., OSImage, tarball, kernel archive, etc. |
| `OSName` | Operating system corresponding to OS image archives |
| **Image** |
| `AssetName` | Asset name override for the image component. If not provided, the repository name is used |
| `RepoName` | Name of the image. This is prefixed with `ImageRepoPrefix` to form the fully-qualified repository name |
| `TrimEksAPrefix` | Denotes that the `eks-anywhere` prefix needs to be trimmed from the name of the image |
| `ImageTagConfiguration` |  |
| **ImageTagConfiguration** |
| `NonProdSourceImageTagFormat` | Source image tag format for dev and staging releases |
| `ProdSourceImageTagFormat` | Source image tag format for prod releases` |
| `ReleaseImageTagFormat` | Release image tag format for dev, staging and prod releases |
| **ManifestComponent** |
| `Name` | Name of the manifest component |
| `ReleaseManifestPrefix` | Manifest prefix override for release. If not specified, the name of the project is used |
| `ManifestFiles` | List of manifest filenames corresponding to this manifest component |
| `NoVersionSuffix` | Denotes that the path to the manifest does not contain the project's Git/tracking version as the suffix |

</details>

To add a project's artifacts to the artifacts table,
1. Determine the assets corresponding to the project - images, archives, manifests, Helm charts, etc.

2. Populate the `AssetConfig` struct for the project, using the above table for reference to identify the purpose of each field.

3. Add the struct to the list of configs in `pkg/assets/config/bundle_release.go` file.

4. If you want to add these artifacts in a bundle, add a file corresponding to the project under `pkg/bundles/` with logic to parse the artifacts table for the constituent projects in the bundle.

5. Run `make dev-release` to ensure that the appropriate release URIs have been generated.
