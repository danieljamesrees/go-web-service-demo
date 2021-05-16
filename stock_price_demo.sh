#!/usr/bin/env bash

set -o errtrace
set -o errexit
set -o pipefail

#set -o xtrace

EXAMPLE_SYMBOL=IBM
DEFAULT_NUMBER_OF_DAYS_TO_LOOKUP=3
DEFAULT_DOCKER_HUB_USERNAME=djrees

# Obviously should not be called demo if it were a real service.
export SERVICE_NAME=${1:-stockpricedemo}

if [[ -z "${SYMBOL}" ]]; then
  SYMBOL="${EXAMPLE_SYMBOL}"
fi

if [[ -z "${NDAYS}" ]]; then
  NDAYS="${DEFAULT_NUMBER_OF_DAYS_TO_LOOKUP}"
fi

if [[ -z "${DOCKER_HUB_USERNAME}" ]]; then
  DOCKER_HUB_USERNAME="${DEFAULT_DOCKER_HUB_USERNAME}"
fi

if [[ -z "${APIKEY}" ]]; then
  echo "${RED}Must specify the API key secret in the environment variable, APIKEY${NC}"
  exit 1
fi

usage() {
  echo "$0: service-name [ci]"
  echo "$0: The ci argument (taken literally) indicates the script is being run on a CI server with tools already installed - defaults to false"
}

__dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

source ${__dir}/common_functions.sh

if [[ -z "${SERVICE_NAME}" ]]; then
  echo "${RED}Must specify a service name${NC}"
  usage
  exit 1
fi

CI=false

HTTP_PORT=9090
TEST_PORT=30000

if [[ "$2" = "ci" ]]; then
  CI=true
fi

set -o nounset

setup_tool() {
  local _tool=$1
  local _check_command="${2:-$1 --version}"

  if [[ -z "$1" ]]; then
    print_with_date "${RED}Must specify a tool to install${NC}"
    return
  fi

  if ${_check_command} 2> /dev/null; then
    print_with_date "${GREEN}${_tool} already installed${NC}"
  else
    print_with_date "${YELLOW}Installing ${_tool} ${NC}"
    brew install "${_tool}"
  fi
}

setup_local_k8s() {
  print_with_date "${YELLOW}Setting up K8s${NC}"

  setup_tool kind "kind version"

  kind delete cluster || true 
  kind create cluster --config go/src/deployments/k8s/config.yaml --wait 300s

  setup_tool kubectl "kubectl version"

  kubectl wait --for=condition=Ready pods --all --namespace kube-system --timeout=300s

  setup_tool yamllint
  yamllint go/src/deployments/

  print_with_date "${YELLOW}Deploying demo on K8s${NC}"

  kubectl apply --filename=go/src/deployments/k8s/namespace.yaml
  kubectl apply --filename=go/src/deployments/k8s/configmap.yaml
  kubectl apply --filename=go/src/deployments/k8s/secret.yaml
  kubectl apply --filename=go/src/deployments/k8s/deployment.yaml
  kubectl apply --filename=go/src/deployments/k8s/service.yaml
  kubectl wait --for=condition=Ready pods --all --namespace "${SERVICE_NAME}" --timeout=300s

  print_with_date "${YELLOW}Deploying Ingress on K8s${NC}"

  kubectl apply --filename=https://raw.githubusercontent.com/kubernetes/ingress-nginx/master/deploy/static/provider/kind/deploy.yaml

  kubectl wait \
    --namespace ingress-nginx \
    --for=condition=ready pod \
    --selector=app.kubernetes.io/component=controller \
    --timeout=600s

  # Don't ask - required even after the subsequent recommended wait has completed.
  # See https://github.com/kubernetes/ingress-nginx/issues/5583.
  kubectl delete -A ValidatingWebhookConfiguration ingress-nginx-admission

  kubectl apply --filename=go/src/deployments/k8s/ingress.yaml
}

if [[ "${CI}" = false ]]; then
  if [[ "$(uname)" != "Darwin" ]]; then
    print_with_date "${RED}Ensure docker, make, and go are installed on non-Mac OSes${NC}"
  else
    setup_tool docker "docker version"
    setup_tool make
    setup_tool go@1.15 "go version"
  fi
fi

print_with_date "${YELLOW}Analysing code${NC}"

cd go/src

set +o nounset
if [[ -z "${GOPATH}" ]]; then
  GOPATH="${HOME}/go"
fi
set -o nounset

PATH=${GOPATH}/bin:${PATH} make --makefile=../Makefile analyse

cd ../..

print_with_date "${YELLOW}Building via docker${NC}"
# TODO Using buildah or podman would allow mounting of the source volume for commands that automatically update the source code.
# TODO Re-enable build args when in git
DOCKER_BUILDKIT=1 docker build \
  --build-arg arg_build_date=$(date +"%Y%m%d_%H-%M_%S") \
  --build-arg arg_built_by=$(whoami) \
  --build-arg "APIKEY=${APIKEY}" \
  --build-arg "HTTP_PORT=${HTTP_PORT}" \
  --build-arg "NDAYS=${NDAYS}" \
  --build-arg "SERVICE_NAME=${SERVICE_NAME}" \
  --build-arg "SYMBOL=${SYMBOL}" \
  --file "${__dir}/go/src/deployments/Dockerfile" \
  --tag "${SERVICE_NAME}" \
  ${__dir}/go
#  --build-arg arg_build_branch=$(git rev-parse --abbrev-ref HEAD) \
#  --build-arg arg_build_project=$(basename -s .git $(git config --get remote.origin.url)) \

docker tag "${SERVICE_NAME}" "${DOCKER_HUB_USERNAME}/${SERVICE_NAME}"
#docker login --username "${DOCKER_HUB_USERNAME}"
docker push "${DOCKER_HUB_USERNAME}/${SERVICE_NAME}"

print_with_date "${YELLOW}Running docker to build the web service image${NC}"
docker container stop "${SERVICE_NAME}" || true
docker container rm "${SERVICE_NAME}" || true
docker run \
  --env "APIKEY=${APIKEY}" \
  --env "HTTP_PORT=${HTTP_PORT}" \
  --env "NDAYS=${NDAYS}" \
  --env "SERVICE_NAME=${SERVICE_NAME}" \
  --env "SYMBOL=${SYMBOL}" \
  --name "${SERVICE_NAME}" \
  --publish "${TEST_PORT}:${HTTP_PORT}" \
  "${DOCKER_HUB_USERNAME}/${SERVICE_NAME}" &

print_with_date "${YELLOW}Running integration test${NC}"

cd go/src

set +o nounset
if [[ -z "${GOPATH}" ]]; then
  GOPATH="${HOME}/go"
fi
set -o nounset

# TODO Implement a timeout.
while [ "`docker inspect -f {{.State.Running}} ${SERVICE_NAME}`" != "true" ]; do
   echo "Waiting for docker container ${SERVICE_NAME} to start"
   sleep 1;
done

APIKEY="${APIKEY}" HTTP_PORT="${HTTP_PORT}" SERVICE_NAME="${SERVICE_NAME}" TEST_PORT="${TEST_PORT}" PATH=${GOPATH}/bin:${PATH} make --makefile=../Makefile test-integration

print_with_date "${YELLOW}Stopping and removing docker container${NC}"
docker container stop "${SERVICE_NAME}" || true
docker container rm "${SERVICE_NAME}" || true

cd ../..

setup_local_k8s

print_with_date "${YELLOW}Testing demo via Ingress${NC}"
curl --insecure -vvv https://localhost/${SERVICE_NAME}

#kind delete cluster
