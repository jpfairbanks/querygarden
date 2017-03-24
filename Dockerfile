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

CMD /opt/featex/bin/featex