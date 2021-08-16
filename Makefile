BUILDER ?= podman
IMG ?= quay.io/fennec-project/podtracer:0.0.1
KUBEBUILDER_PATH ?= $$HOME/


# podtracer-build builds podtracer binary
podtracer-build:
	go build -o build/podtracer main.go

# Container Build builds podtracer container image
container-build:
	${BUILDER} build -t ${IMG} build/

# container-push pushes to quay image path indicated by ${IMG}
container-push:
	${BUILDER} push ${IMG}

# Demo Deploy with credentials and permissions
deploy-podtracer:
	kubectl apply -f manifests/

