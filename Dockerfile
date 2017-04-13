# This is the docker file that builds the container that runs the executable compiled by
# the container built by Dockerfile.build.
FROM lacion/docker-alpine:latest

ARG GIT_COMMIT
ARG VERSION
LABEL REPO="https://github.com/jpfairbanks/featex"
LABEL GIT_COMMIT=$GIT_COMMIT
LABEL VERSION=$VERSION

# Because of https://github.com/docker/docker/issues/14914
ENV PATH=$PATH:/opt/featex/bin

WORKDIR /opt/featex/bin

COPY bin/featex /opt/featex/bin/
RUN chmod +x /opt/featex/bin/featex

# This should be mounted, but I don't know how to mount a single file.
# I think an /etc directory would be a good long term solution.
COPY ./featex_config.yaml /opt/featex/featex_config.yaml
# these directories can be either:
#  1. copied to make an image for docker hub
#  2. mounted from the host to be able to edit them.
# COPY ./sql/ /opt/featex/sql
# COPY ./templates/ /opt/featex/templates

CMD cd /opt/featex && /opt/featex/bin/featex
