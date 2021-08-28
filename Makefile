# kubeconfig file path and namespace to be used with podtracer
NAMESPACE ?= podtracer

# This is for developers using GoRemote only
# in order to develop podtracer
DEV_NAMESPACE ?= podtracer-dev
BUILDER ?= podman
IMG ?= quay.io/fennec-project/podtracer:0.0.1-5

# podtracer-build builds podtracer binary
podtracer-build:
	go build -o build/podtracer main.go

# Container Build builds podtracer container image
container-build:
	${BUILDER} build -t ${IMG} build/

# container-push pushes to quay image path indicated by ${IMG}
container-push:
	${BUILDER} push ${IMG}

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