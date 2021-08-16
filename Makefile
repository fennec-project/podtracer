BUILDER ?= podman
IMG ?= quay.io/fennec-project/podtracer:0.0.1
KUBECONFIG_PATH ?= $$HOME/.kube/config
NAMESPACE= ?= podtracer-dev

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
deploy-podtracer: kubeconfig-secret --NAMESPACE=podtracer
	kubectl apply -f deploy/ -n podtracer

# Creates the kubeconfig secret to be used with the dev image
# and also with the podtracer image when running it from a container.
kubeconfig-secret:
	kubectl create secret generic podtracer-secret --type=string --from-file=${KUBECONFIG_PATH} -n ${NAMESPACE}

# dev-env deploys the development environment using goremote
# A kubebuilder path should be included for it use or the
# default kubebuilder $HOME/.kube/config will be used if available
# Important to note that the target pod must be running on the same node
dev-env: kubeconfig-secret --NAMESPACE=podtracer-dev
	kubectl apply -f dev/ -n podtracer-dev