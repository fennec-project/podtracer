/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"context"
	"io"
	"strings"

	"os/exec"

	logger "log"
	"os"

	"github.com/containernetworking/plugins/pkg/ns"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs arbitrary Linux command on a targeted kubernetes pod.",
	Long: `podtracer run - runs arbitrary Linux command line tools such as tcpdump, 
	tshark, iperf and others to acquire network data and metrics for observability purposes 
	 without changing the pod.`,

	// ValidArgs: []string{"tcpdump"},
	Args: argFuncs(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {

		// calling main podtracer command
		// p := podtracer.Podtracer{}
		err := Run(args[0], &flags)
		if err != nil {
			fmt.Printf("An error ocurred while running pod tracer run: %v", err.Error())
		}
	},
}

type runCommand struct {
	taskName string

	cliTool string

	// arguments for the tool being run by podtracer
	targetArgs string

	// the of the pod under troubleshooting
	targetPod string

	// namespace of the pod under troubleshooting
	targetNamespace string

	// path for kubeconfig file
	// TODO: still needs investigation if it is really needed
	// The service account running podtracer Pod should be enough to list
	// pods and namespaces. But under dev env with VSCode it's seems to
	// be required.
	kubeconfig string

	// file path to store os/exec cmd.stdout output
	stdoutFile string

	// file path to store os/exec cmd.stderr output
	stderrFile string

	// TODO: Linux namespace set to switch to before running
	// selected tools with podtracer.
	// Needs to be its own type limited to only valid namespaces.
	// linuxNSSet linuxNSSet

	// TODO: enable running non-valid untested args as tools
	// unsafe bool // --unsafe
}

var flags runCommand

func init() {
	rootCmd.AddCommand(runCmd)

	// Flags for run

	runCmd.Flags().StringVarP(&flags.targetArgs, "arguments", "a", "", "arguments to running cli utility.")

	runCmd.Flags().StringVar(&flags.targetPod, "pod", "", "Target pod name.")

	runCmd.Flags().StringVarP(&flags.targetNamespace, "namespace", "n", "", "Kubernetes namespace where the target pod is running")

	runCmd.Flags().StringVarP(&flags.kubeconfig, "kubeconfig", "k", "", "kubeconfig file path to connect to kubernetes cluster - defaults to $HOME/.kube/kubeconfig")

	runCmd.Flags().StringVarP(&flags.stdoutFile, "stdoutFile", "o", "", "file path to save output data from the running tool.")

	runCmd.Flags().StringVarP(&flags.stderrFile, "stderrFile", "e", "", "file path to save output data from the running tool.")

	// Required Flags
	runCmd.MarkFlagRequired("pod")

	runCmd.MarkFlagRequired("namespace")
}

func argFuncs(funcs ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range funcs {
			err := f(cmd, args)
			if err != nil {
				return err
			}
		}
		return nil
	}
}

func GetClient(kubeconfigPath string) (client.Client, error) {

	// TODO: link kubeconfigPath on client.new if empty default to ~/.kube/kubeconfig
	c, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}
	return c, nil
}

func GetPod(targetPod string, targetNamespace string, kubeconfig string) (corev1.Pod, error) {

	c, err := GetClient(kubeconfig)
	if err != nil {
		return corev1.Pod{}, err
	}

	pod := corev1.Pod{}
	err = c.Get(context.Background(), client.ObjectKey{
		Namespace: targetNamespace,
		Name:      targetPod,
	}, &pod)
	if err != nil {
		return corev1.Pod{}, err
	}
	return pod, nil

}

func Run(tool string, runFlags *runCommand) error {

	pod, err := GetPod(runFlags.targetPod, runFlags.targetNamespace, runFlags.kubeconfig)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	// TODO: create a podInspect struct to handle pod and container data
	// and add it as a receiver on the getPid function.

	pid, err := getPid(pod)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	// Get the pod's Linux namespace file descriptor
	targetNS, err := ns.GetNS("/host/proc/" + pid + "/ns/net")
	if err != nil {
		return fmt.Errorf("error getting Pod network namespace: %v", err)
	}

	// Switching Linux Namespaces
	err = targetNS.Do(func(hostNs ns.NetNS) error {

		splitArgs := strings.Split(runFlags.targetArgs, " ")

		logger.Printf("[INFO] Running %s: Pod %s Namespace %s \n\n", tool, runFlags.targetPod, runFlags.targetNamespace)

		writeToBufferAndStdout := io.MultiWriter(stdOutWriters...)
		writeToBufferAndStderr := io.MultiWriter(stdErrWriters...)

		cmd := exec.Command(tool, splitArgs...)
		cmd.Stdout = writeToBufferAndStdout
		cmd.Stderr = writeToBufferAndStderr

		err = cmd.Run()
		if err != nil {
			fmt.Printf("Error: %s\n %v", err.Error(), cmd.Stderr)
			return err
		}

		// Log("DATA", "Stdout: %v \n\n", bufferedStdout.String())
		// Log("DEBUG", "Stderr: %v\n Exit Code: %v", bufferedStderr.String(), err)

		return nil
	})
	if err != nil {
		fmt.Println(err.Error())
		return err
	}

	return nil
}

// func run(...args)
// 	1 - Get a client to talk to kubernetes
// 	2 - Get the target Pod or Pods by the chosen criteria (can we pass context withValues on this get query?)
//  3 - Get the Pod's first container Pid (can we choose the container instead?)
// 	4 - Setup the file descriptor for the Linux namespace switching (can we add other namespaces? Extend this to something like strace or ftrace?)
// 	5 - Switch to target namespace or namespaces
// 		a. setup writers to collect data
//		b. run tool with os.exec
// 			- needs to handle all sorts of SIGnals and terminate properly the child process
