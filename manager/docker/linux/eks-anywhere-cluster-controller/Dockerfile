ARG BASE_IMAGE # https://gallery.ecr.aws/eks-distro-build-tooling/eks-distro-minimal-base
FROM $BASE_IMAGE

ARG TARGETARCH
ARG TARGETOS

COPY _output/bin/eks-anywhere-cluster-controller/$TARGETOS-$TARGETARCH/manager /usr/local/bin/manager
COPY _output/LICENSES /LICENSES
COPY ATTRIBUTION.txt /ATTRIBUTION.txt

ARG EKS_A_TOOL_BINARY_DIR=/eks-a-tools/binary
ARG EKS_A_TOOL_LICENSE_DIR=/eks-a-tools/licenses

COPY _output/dependencies/$TARGETOS-$TARGETARCH/eks-a-tools/binary $EKS_A_TOOL_BINARY_DIR
COPY _output/dependencies/$TARGETOS-$TARGETARCH/eks-a-tools/licenses $EKS_A_TOOL_LICENSE_DIR

ENV PATH="${EKS_A_TOOL_BINARY_DIR}:${PATH}"

USER 65534

ENTRYPOINT ["manager"]
