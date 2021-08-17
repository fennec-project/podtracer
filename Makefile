# kubeconfig file path and namespace to be used with podtracer
PODTRACER_KUBECONFIG ?= $$HOME/.kube/config
NS ?= podtracer

# This is for developers using GoRemote only
# in order to develop podtracer
DEV_KUBECONFIG_PATH ?= $$HOME/.kube/config
DEV_NS ?= podtracer-dev
BUILDER ?= podman
IMG ?= quay.io/fennec-project/podtracer:0.0.1

# podtracer-build builds podtracer binary
podtracer-build:
	go build -o build/podtracer main.go

# Container Build builds podtracer container image
container-build:
	${BUILDER} build -t ${IMG} build/

# container-push pushes to quay image path indicated by ${IMG}
container-push:
	${BUILDER} push ${IMG}

# Kubeconfig secret for podtracer deployment
podtracer-kubeconfig:
	kubectl create secret generic podtracer-kubeconfig --type=string --from-file=${PODTRACER_KUBECONFIG} -n ${NS}

# deploy-podtracer deploys a container with podtracer image
# It allows engineers to run podtracer from a troubleshooting container 
# on any given cluster if they have the proper permissions.
podtracer-deploy: kubeconfig-secret
	kubectl apply -f deploy/ -n ${NS}

# Creates the kubeconfig secret to be used with the dev image
# and also with the podtracer image when running it from a container.
dev-kubeconfig:
	kubectl create secret generic podtracer-kubeconfig --type=string --from-file=${DEV_KUBECONFIG_PATH} -n ${DEV_NS}

delete-dev-kubeconfig:
	kubectl delete secret podtracer-kubeconfig -n ${DEV_NS}
# dev-env deploys the development environment using goremote
# A kubebuilder path should be included for it use or the
# default kubebuilder $HOME/.kube/config will be used if available
# Important to note that the target pod must be running on the same node
dev-env: dev-kubeconfig
	kubectl apply -f dev/ -n ${DEV_NS}


delete-dev-env: delete-dev-kubeconfig
	kubectl delete -f dev/ -n ${DEV_NS}