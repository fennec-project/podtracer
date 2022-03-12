# kubeconfig file path and namespace to be used with podtracer
NAMESPACE ?= podtracer

# This is for developers using GoRemote only
# in order to develop podtracer
DEV_NAMESPACE ?= podtracer-dev
BUILDER ?= podman
IMG ?= quay.io/fennec-project/podtracer:0.1.0

GOOS=linux
GOARCH=amd64

PODTRACER_VERSION := "v0.1.0-alpha"
GO_VERSION := $(shell go version | cut -f 3 -d " ")
BUILD_TIME := $(shell date)
GIT_USER := $(shell git log | grep -A2 $$(git rev-list -1 HEAD) | grep Author)
GIT_COMMIT := $(shell git rev-list -1 HEAD)

# build builds podtracer binary
.PHONY: build
build:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -ldflags "-X 'github.com/fennec-project/podtracer/cmd.Version=$(PODTRACER_VERSION)' \
													-X 'github.com/fennec-project/podtracer/cmd.GoVersion=$(GO_VERSION)' \
													-X 'github.com/fennec-project/podtracer/cmd.BuildTime=$(BUILD_TIME)' \
													-X 'github.com/fennec-project/podtracer/cmd.GitUser=$(GIT_USER)' \
													-X 'github.com/fennec-project/podtracer/cmd.GitCommit=$(GIT_COMMIT)'" \
													-o build/bin/podtracer main.go

# Container Build builds podtracer container image
.PHONY: container-build
container-build:
	${BUILDER} build -t ${IMG} build/

# container-push pushes to quay image path indicated by ${IMG}
.PHONY: container-push
container-push:
	${BUILDER} push ${IMG}

container-latest: build
	${BUILDER} build -t quay.io/fennec-project/podtracer:latest build/
	${BUILDER} push quay.io/fennec-project/podtracer:latest

# deploy-podtracer deploys a container with podtracer image
# It allows engineers to run podtracer from a troubleshooting container 
# on any given cluster if they have the proper permissions.
podtracer-deploy:
	kubectl create ns ${NAMESPACE}
	kubectl apply -f manifests/deploy/ -n ${NAMESPACE}

podtracer-delete:
	kubectl delete -f manifests/deploy/ -n ${NAMESPACE}
	kubectl delete ns ${NAMESPACE}

# dev-env deploys the development environment using goremote
# A kubebuilder path should be included for it use or the
# default kubebuilder $HOME/.kube/config will be used if available
# Important to note that the target pod must be running on the same node
podtracer-dev:
	kubectl create ns ${DEV_NAMESPACE}
	kubectl apply -f manifests/dev/ -n ${DEV_NAMESPACE}

delete-podtracer-dev:
	kubectl delete -f manifests/dev/ -n ${DEV_NAMESPACE}
	kubectl delete ns ${DEV_NAMESPACE}

sample-deployment:
	kubectl apply -f manifests/sample_deployment/

delete-samples:
	kubectl delete -f manifests/sample_deployment/