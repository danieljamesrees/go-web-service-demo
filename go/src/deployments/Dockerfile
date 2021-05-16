# https://hub.docker.com/_/golang
FROM golang:1.15-alpine as builder

ARG HTTP_PORT
ARG SERVICE_NAME

ENV APIKEY=dummy

ENV WORKSPACE=/go
ENV GOPATH /tmp/go
#ENV HADOLINT_PATH /tmp

WORKDIR ${WORKSPACE}

RUN \
  echo Updating and installing dependencies && \
  apk update && \
  apk upgrade && \
  apk add --no-cache git make

# TODO Install hadolint to lint the web service Dockerfile! Couldn't get to work on alpine - either from source or a built image.
# 8.10.4
#COPY --from=fpco/alpine-haskell-stack@sha256:1024fe4b3b082a8df64d00e8563b3151220ed90af09604a8f7e1d44040500c30 /usr/local/bin/stack /usr/local/bin/stack
#RUN  \
#  echo Installing hadolint && \
#  apk add curl ghc git musl-dev zlib-dev && \
#  curl --location --show-error --silent https://get.haskellstack.org/ | sh && \
#  git clone https://github.com/hadolint/hadolint && \
#  cd hadolint && \
#  stack --system-ghc install hadolint
#  stack install hadolint

#COPY ./src/deployments/hadolint-config.yaml ${HADOLINT_PATH}/
COPY ./Makefile ${WORKSPACE}/
COPY ./src/go.mod src/
ADD ./swaggerui ${WORKSPACE}/swaggerui

RUN \
  cd src && \
  PATH=${GOPATH}/bin:${PATH} make --makefile=../Makefile prepare

# TODO Should really excludes the deployments directory, as they are not needed for the build.
ADD ./src ${WORKSPACE}/src

RUN \
  echo Making ${SERVICE_NAME} && \
  cd src && \
  PATH=${GOPATH}/bin:${PATH} make --makefile=../Makefile test-unit

# Lint the web service Dockerfile.
#RUN \
#  echo Linting Dockerfile && \
#  hadolint --config ${HADOLINT_PATH}/hadolint-config.yaml

FROM alpine

ARG arg_build_date
ARG arg_build_branch
ARG arg_build_project
ARG arg_built_by

LABEL \
  build-date=${arg_build_date} \
  build-branch=${arg_build_branch} \
  built-project=${arg_build_project} \
  built-by=${arg_built_by}

ENV GOPATH /tmp/go

ENV USERNAME=demo

ARG APIKEY
ARG HTTP_PORT
ARG SERVICE_NAME
ARG SYMBOL

RUN \
  echo Adding user ${USERNAME} && \
  addgroup -S ${USERNAME} && adduser -S ${USERNAME} -G ${USERNAME} -u 1000 && \
  find /home

WORKDIR /home/${USERNAME}

EXPOSE ${HTTP_PORT}

COPY --from=builder ${GOPATH}/${SERVICE_NAME} /home/${USERNAME}/${SERVICE_NAME}

# Not installing bash to resolve the environment variables.
#ENTRYPOINT ["bash", "-c", "/home/${USERNAME}/${SERVICE_NAME}"]
ENTRYPOINT ["/home/demo/stockpricedemo"]

USER ${USERNAME}

# Is disabled on Kubernetes (and appears that it will remain that way), so commented to avoid confusion - see https://github.com/kubernetes/kubernetes/pull/50796. curl may need to be install if this healthcheck is uncommented. Untested.
#HEALTHCHECK --start-period=60s \
#  CMD curl http://localhost:${HTTP_PORT}/${SERVICE_NAME} || exit 1
