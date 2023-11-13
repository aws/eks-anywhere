# EKS Anywhere Release Artifact Table Generation

## Introduction

**Problem:** During the EKS-A bundle release process, we construct a map of project name to artifacts corresponding to that project. This includes the images, manifests, tarballs we build and vend for that project. Each artifact contains a few common elements such as its name, source/release ECR URIs (for images), and source/release S3 paths (for manifests and tarballs). Constructing this "table" makes it easier to record and track the source and destination locations for all artifacts. Currently, each project has its own file, i.e., `assets_<project_name>.go`, for defining and packaging the artifacts for that project. However, since several projects are bundled in the same way, there is a significant amount of code duplication, and common refactors that need to be applied to one process affect all the `assets_X` files, making the review process harder.

### Tenets

* **Simple:** Adding another component to the bundle should be simple and with minimal changes and little to no code duplication
* **Declarative:** The list of artifacts for a project should be stored declaratively, for abstraction and easier tracking

## Goals

As an EKS Anywhere developer, I want to

* Make it easy to add another component to the bundle
* Make the review process for bundle changes and refactors easier
* Reduce code duplication and techdebt in the release tooling

## Implementation details

Currently, the artifacts declaration for each project is stored in its separate Go file, along with the logic for packaging assets (configuring source/release destinations, versioning, etc), that is mostly common across all these files.

Instead of this, it would be ideal to take the declarative approach and define the artifacts for a project as a set of Go custom structs in a single file. This file would contain the list of artifacts (images, manifests, archives) for each component and some other parameters that are required by the common methods, which will all be extracted into a single file. During the release process, we can then iterate over these structs and feed them into the "oracle" that contains the business logic to get the final artifacts table.

**Current folder structure:**

```
   release
      |_____ pkg
              |_______ assets_capi.go
              |_______ assets_capa.go
              |_______ assets_capd.go
              |_______ assets_capv.go
              .
              .
              .
```

**Proposed folder structure:**

```
   release
      |_____ pkg
              |_______ assets
                         |_______ assets.go # contains processing logic for artifacts
                         |_______ config.go # contains declarative definition for project artifacts
```

The following schema will be used to define project artifacts as a Go struct:

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

The fields in the struct will be used to construct the artifacts table appropriately. If any change needs to be made to the business logic of asset generation, developers will only need to touch the `assets.go` file.
