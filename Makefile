# kubeconfig file path and namespace to be used with podtracer
PODTRACER_KUBECONFIG ?= $$HOME/.kube/config
NS ?= podtracer

# This is for developers using GoRemote only
# in order to develop podtracer
DEV_KUBECONFIG_PATH ?= $$HOME/.kube/config
DEV_NS ?= podtracer-dev
BUILDER ?= podman
IMG ?= quay.io/fennec-project/podtracer:0.0.1-2

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
	kubectl create ns ${NS}
	kubectl create secret generic podtracer-kubeconfig --type=string --from-file=${PODTRACER_KUBECONFIG} -n ${NS}
	kubectl apply -f manifests/deploy/ -n ${NS}

delete-podtracer:
	kubectl delete -f manifests/deploy/ -n ${NS}
	kubectl delete secret podtracer-kubeconfig -n ${NS}
	kubectl delete ns ${NS}


# dev-env deploys the development environment using goremote
# A kubebuilder path should be included for it use or the
# default kubebuilder $HOME/.kube/config will be used if available
# Important to note that the target pod must be running on the same node
dev-env:
	kubectl create ns ${DEV_NS}
	kubectl create secret generic podtracer-kubeconfig --type=string --from-file=${DEV_KUBECONFIG_PATH} -n ${DEV_NS}
	kubectl apply -f manifests/dev/ -n ${DEV_NS}


delete-dev-env:
	kubectl delete -f manifests/dev/ -n ${DEV_NS}
	kubectl delete secret podtracer-kubeconfig -n ${DEV_NS}	
	kubectl delete ns ${DEV_NS}

sample-deployment:
	kubectl apply -f manifests/sample_deployment/