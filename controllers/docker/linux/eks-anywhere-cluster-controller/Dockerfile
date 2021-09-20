ARG BASE_IMAGE # https://gallery.ecr.aws/eks-distro-build-tooling/eks-distro-minimal-base
FROM $BASE_IMAGE

COPY _output/bin/eks-anywhere-cluster-controller/linux-amd64/manager /usr/local/bin/manager
COPY _output/LICENSES /LICENSES
COPY ATTRIBUTION.txt /ATTRIBUTION.txt

USER 65534

ENTRYPOINT ["manager"]