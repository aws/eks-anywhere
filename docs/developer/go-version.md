# Go version

## How to update the repo's Go Version

1. Make sure the [image builder](https://github.com/aws/eks-anywhere-build-tooling/blob/main/build/lib/install_go_versions.sh#L37) supports the new Go minor version. If it doesn't, you will need to add it there first.
1. Update `GOLANG_VERSION` in the [main `Makefile`](https://github.com/aws/eks-anywhere/blob/main/Makefile#L23)
1. Update `GOLANG_VERSION` in the [`release` `Makefile`](https://github.com/aws/eks-anywhere/blob/main/release/Makefile#L45)
1. Update `go-version` in [`codecov` workflow](https://github.com/aws/eks-anywhere/blob/main/.github/workflows/go-coverage.yml#L17)
1. Update `go-version` in [`golangci-lint` workflow](https://github.com/aws/eks-anywhere/blob/main/.github/workflows/golangci-lint.yml#L17)