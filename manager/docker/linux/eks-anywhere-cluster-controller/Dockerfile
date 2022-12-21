ARG BASE_IMAGE # https://gallery.ecr.aws/eks-distro-build-tooling/eks-distro-minimal-base-nonroot
FROM $BASE_IMAGE

ARG TARGETARCH
ARG TARGETOS

ARG BASE_PATH=.
ARG MANAGER_BIN_PATH=$BASE_PATH/_output/bin/eks-anywhere-cluster-controller/$TARGETOS-$TARGETARCH/manager
ARG LICENSES_PATH=$BASE_PATH/_output/LICENSES
ARG ATTRIBUTION_PATH=$BASE_PATH/ATTRIBUTION.txt
ARG DEPENDENCY_BINARIES_PATH=$BASE_PATH/_output/dependencies/$TARGETOS-$TARGETARCH/eks-a-tools/binary
ARG DEPENDENCY_LICENSES_PATH=$BASE_PATH/_output/dependencies/$TARGETOS-$TARGETARCH/eks-a-tools/licenses

ARG DST_EKS_A_TOOL_BINARY_DIR=/eks-a-tools/binary
ARG DST_EKS_A_TOOL_LICENSE_DIR=/eks-a-tools/licenses
ARG DST_MANAGER_BINARY_DIR=/usr/local/bin

# This is not needed when using the eks-d minimal nonroot base image
# but it makes builds more consistent when changing the base for Tilt
USER 65532
WORKDIR "/home/nonroot"

ENV PATH="${DST_MANAGER_BINARY_DIR}:${DST_EKS_A_TOOL_BINARY_DIR}:${PATH}"

COPY $DEPENDENCY_BINARIES_PATH $DST_EKS_A_TOOL_BINARY_DIR
COPY $DEPENDENCY_LICENSES_PATH $DST_EKS_A_TOOL_LICENSE_DIR

COPY $LICENSES_PATH /LICENSES
COPY ATTRIBUTION.txt /ATTRIBUTION.txt

COPY $MANAGER_BIN_PATH $DST_MANAGER_BINARY_DIR/manager

ENTRYPOINT ["manager"]
