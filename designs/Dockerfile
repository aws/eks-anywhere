# - Example:
# docker run \
# 	--rm \
# 	--volume ${PWD}:/figures \
# 	--user $(shell id --user):$(shell id --group) \
# 	${IMAGE_TAG} \
# 	-v /figures/*.plantuml

FROM maven:3-jdk-8

ARG PLANTUML_VERSION=1.2022.6

RUN apt-get update && apt-get install -y --no-install-recommends graphviz fonts-symbola fonts-wqy-zenhei && rm -rf /var/lib/apt/lists/*
RUN wget -O /plantuml.jar "https://downloads.sourceforge.net/project/plantuml/${PLANTUML_VERSION}/plantuml.${PLANTUML_VERSION}.jar"

# By default, java writes a 'hsperfdata_<username>' directory in the work dir.
# This directory is not needed; to ensure it is not written, we set `-XX:-UsePerfData`
ENTRYPOINT [ "java", "-XX:-UsePerfData", "-jar", "/plantuml.jar" ]