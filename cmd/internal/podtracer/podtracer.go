package podtracer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	cri "k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// Podtracer struct holds information about current podtracer
// or podtracers task(s) running and all interesting pieces of
// information that may be added to the os/exec write operation
// to enrich data passed along to central data collection endpoints
type Podtracer struct {

	// This client should use only pod's token credentials
	// With the service account being able to list but not
	// change objects.
	client.Client

	// targetPod is the complete pod object for now
	// It may be de-scoped to less information distributed
	// on multiple smaller fields
	TargetPod corev1.Pod

	// ContainersInfo is the complete json blob coming from the
	// info field on CRI-O's containerStatusResponse
	// It may de-scoped to less information distributed on
	// multiple smaller fields
	ContainersInfo []map[string]interface{}

	// The task in execution. May be a cli tool under the run
	// command, executing any go package that uder a different
	// command or eBPF tools for example
	task string
}

func (p *Podtracer) GetClient(kubeconfigPath string) error {

	// TODO: link kubeconfigPath on client.new if empty default to ~/.kube/kubeconfig
	c, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}
	p.Client = c
	return nil
}

func (p *Podtracer) GetPod(targetPodName string, targetNamespace string, kubeconfig string) error {

	targetPod := corev1.Pod{}
	err := p.Get(context.Background(), client.ObjectKey{
		Namespace: targetNamespace,
		Name:      targetPodName,
	}, &targetPod)
	if err != nil {
		return err
	}

	p.TargetPod = targetPod
	return nil
}

func (p *Podtracer) GetContainerIDs(pod corev1.Pod) []string {

	containerIDs := []string{}

	// get container ID list
	for _, containerStatus := range pod.Status.ContainerStatuses {

		containerIDs = append(containerIDs, containerStatus.ContainerID[8:])

	}
	return containerIDs
}

func (p *Podtracer) GetCRIOContainerInfo(containerID string) error {

	var grpcConn *grpc.ClientConn

	// TODO: check how to properly authenticate with grpc on top of crio socket
	// prerrably in read only mode. We don't want to write to the socket just
	// read from it.
	grpcConn, err := grpc.Dial("unix:///var/run/crio/crio.sock", grpc.WithInsecure())
	if err != nil {
		return err
	}

	// TODO: Optimize LOG DEBUG - missing a proper logger
	Log("DEBUG", "Connected with CRI-O at unix:///var/run/crio/crio.sock")

	criClient := cri.NewRuntimeServiceClient(grpcConn)

	request := &cri.ContainerStatusRequest{
		ContainerId: containerID,
		Verbose:     true,
	}
	response, err := cri.RuntimeServiceClient.ContainerStatus(criClient, context.Background(), request)
	if err != nil {
		return err
	}

	grpcConn.Close()

	// parsing container info JSON
	var parsedContainerInfo map[string]interface{}

	containerInfo := response.Info["info"]

	err = json.Unmarshal([]byte(containerInfo), &parsedContainerInfo)
	if err != nil {
		return err
	}

	p.ContainersInfo = append(p.ContainersInfo, parsedContainerInfo)

	return nil
}

// FUTURE:
// For now this method isn't in use. It's for a future use targeting multiple
// pods at the same time with possibly multiple tasks. Yaml configuration
// must be created to handle a more complex set of parameters.
func (p *Podtracer) ListPodsWithMatchingLabels(label string, value string) error {

	podList := &corev1.PodList{}
	// Get the list of pods that have a podNetworkConfig label
	err := p.List(context.Background(), podList, client.MatchingLabels{label: value})
	if err != nil {
		fmt.Printf("failed to list pods matching labels: %v\n", err)
		os.Exit(1)
	}

	// Pods need to be at least created to proceed
	// Checking if the list is empty
	if len(podList.Items) <= 0 {
		return fmt.Errorf("No matching Pods found with label %s: %s", label, value)
	}
	return nil
}
