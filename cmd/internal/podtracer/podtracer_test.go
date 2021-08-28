package podtracer

import (
	"fmt"
	"testing"
)

func testPodtracer(t *testing.T) {
	testScenarios := []struct {
		name             string
		tool             string
		targetArgs       string
		pod              string
		namespace        string
		kubeconfig       string
		stdoutFile       string
		stderrFile       string
		expectedErrorMsg string
	}{
		// TODO: create the testing scenarios below
		// {
		// 	name:             "kubeconfigNotFound",
		// 	tool:             "tcpdump",
		// 	targetArgs:       "-i eth0 -c 10 -w /pcap-data/test.pcap",
		// 	pod:              "podtest",
		// 	namespace:        "namespacetest",
		// 	kubeconfig:       "/any/any/kubeconfig",
		// 	expectedErrorMsg:
		// },
		// {
		// 	name: "k8sClusterUnreachable",
		// },
		// {
		// 	name: "validNamespaceName",
		// },
		// {
		// 	name: "validPodName",
		// },
		// {
		// 	name: "tcpdumpError",
		// },
	}
	for _, scenario := range testScenarios {
		t.Run(scenario.name, func(t *testing.T) {

			p := Podtracer{}

			// Validating kubeconfig file
			client, err := p.GetClient(scenario.kubeconfig)

			// TODO: check if client is a valid runtime client type
			fmt.Printf("%v", client)

			// Validating errors when kubeconfig is not valid or not found
			if want, got := scenario.expectedErrorMsg, err.Error(); want != got {
				t.Errorf("Expected error %v, but got %v.", want, got)
				return
			} else if err.Error() != "" {
				return // Expected error message ok. Done.
			}

			// Validating Pod and Namespaces names and existence
			pod, err := p.GetPod(scenario.pod, scenario.namespace, p.Kubeconfig)

			// TODO: check if pod is corev1.Pod type
			fmt.Printf("%v", pod)

			if want, got := scenario.name, err.Error(); want != got {
				t.Errorf("Expected error %v, but got %v.", want, got)
				return
			} else if err.Error() != "" {
				return // Expected error message ok. Done.
			}

			err = p.Run(scenario.tool, scenario.targetArgs, scenario.pod, scenario.namespace, scenario.kubeconfig, scenario.stdoutFile, scenario.stderrFile)

			if want, got := scenario.name, err.Error(); want != got {
				t.Errorf("Expected error %v, but got %v.", want, got)
				return
			} else if err.Error() != "" {
				return // Expected error message ok. Done.
			}

		})
	}

}
