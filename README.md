# Overview

A demo web service using Go and Kubernetes, lacking proper validation and security features. Unfit for use in a production environment. Only intended to be deployed on a local Kubernetes cluster.

# Usage

To use this demo locally, clone the repo and run the script stock_price_demo.sh on a Linux machine that has docker already installed. It has only been tested on a Mac machine.
**NB. The script will delete any existing kind clusters running on your machine.**

To push the built Docker image to Docker Hub, the user is first required to login, e.g. using `docker login --username BLAH`. It is published at https://hub.docker.com/repository/docker/djrees/stockpricedemo/general.

The stock symbol and number of days to look up and defined using the environment variables SYMBOL and NDAYS respectively.

The API key secret should be defined in an environment variable - in reality it would be encrypted or retrieved from a Secrets Manager. This variable will be passed to Docker for local testing and stored in a K8s Secret for usage in a K8s cluster. See https://kubernetes.io/docs/tasks/administer-cluster/encrypt-data/ for info on encryption secrets, though the protection of the K8s API server and encryption configuration still need to be considered.

# Notes

* No design documentation is provided.
* See https://kind.sigs.k8s.io/docs/user/configuration#extra-port-mappings if the kind cluster doesn't work on non-MacOS Linux.
* For simplicity, no API/web routing framework has been used.
* Some of the Go code is hideously over-engineered for the simple example, as it was based on a use case where the flexibility was more justifiable.
* Ingress NGINX has been chosen as the Ingress Controller as it's the only one of the three controller officially supported by K8s that works with kind (and not AWS or GCP).
* No Docker-based development environment has been provided, due to concerns over the time it would take getting a local K8s cluster with/inside a Docker container.
* Built using Docker in the absence of an approved IDE.
* To minimise deployment complexity and development time, no use of deployment tools, such as Argo or Flux, have been considered. Instead raw K8s manifests are provided. The service name and ports are hard-coded to avoid using kustomization.yamls, a tool such as Ansible to template the variables, or making the ConfigMap more complex.
* To make installation and configuration relatively agnostic of the Linux variant used for development and local deployment, builds and the cluster will be set up using a Makefile. However, the Makefile has only been tested on a single version of MacOS, and it contains a fair few shell commands anyway.
* Little thought has gone into the testing approach - it's largely copied from some tests produced for another Go-based noddy web service.

# Deliberate Omissions

* Cost considerations if deploying the service remotely.
* DR/backups.
* Monitoring and structured logging.
* Performance considerations (e.g. CPU/memory optimisation and scaling) - other than allowing for multiple containers/Pod instances.
* Resilience considerations (e.g. cross-region deployments) - other than within the provided K8s objects.
* Security (e.g. API Gateway, cluster hardening, rate limiting/DOS protection, use of hardened container images).

# TODOs

* Add clean-up and error handling.
* Implement authentication for the service.
* Enable TLS with generated certificates.
* Version pinning/updating.
* Run the build and tests on a CI/CD server.
* shellchecking of scripts.
* Immutable ConfigMap and Secret (for K8s 1.21).
* Optimise the container spec and workflow.
