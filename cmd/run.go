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
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"

	Podtracer "github.com/fennec-project/podtracer/cmd/internal/podtracer"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs arbitrary Linux command on a targeted kubernetes pod.",
	Long: `podtracer run - runs arbitrary Linux command line tools such as tcpdump, 
		tshark, iperf and others to acquire network data and metrics for observability purposes 
		 without changing the pod.`,

	// ValidArgs: []string{"tcpdump"},
	Args: argFuncs(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {

		err := Run(args[0])
		if err != nil {
			fmt.Printf("An error ocurred while running pod tracer run: %v", err.Error())

		}
	},
}

// runCmd represents the run command

type runCommand struct {

	// arguments for the tool being run by podtracer
	targetArgs string

	// the of the pod under troubleshooting
	targetPodName string

	// namespace of the pod under troubleshooting
	targetNamespace string

	// path for kubeconfig file
	// TODO: still needs investigation if it is really needed
	// The service account running podtracer Pod should be enough to list
	// pods and namespaces. But under dev env with VSCode it's seems to
	// be required.
	kubeconfigPath string

	// file path to store os/exec cmd.stdout output
	stdoutFile string

	// Destination IP to send captured packets to
	destinationIP string

	// Destination port to send captured packets to
	destinationPort string

	// Writers send data to a desired destination
	writers []io.Writer

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
	runCmd.Flags().StringVar(&flags.targetPodName, "pod", "", "Target pod name.")
	runCmd.Flags().StringVarP(&flags.targetNamespace, "namespace", "n", "", "Kubernetes namespace where the target pod is running")
	runCmd.Flags().StringVarP(&flags.kubeconfigPath, "kubeconfig", "k", "", "kubeconfig file path to connect to kubernetes cluster - defaults to $HOME/.kube/kubeconfig")
	runCmd.Flags().StringVarP(&flags.stdoutFile, "stdoutFile", "o", "", "file path to save output data from the running tool.")
	runCmd.Flags().StringVarP(&flags.destinationIP, "destination", "d", "", "Destination IP to where send stdout")
	runCmd.Flags().StringVarP(&flags.destinationPort, "port", "p", "", "Destination port to where send stdout")

	// Required Flags
	runCmd.MarkFlagRequired("pod")
	runCmd.MarkFlagRequired("namespace")

}

func initWriters() error {

	flags.writers = append(flags.writers, os.Stdout)
	if flags.stdoutFile != "" {

		stdoutFile, err := os.OpenFile(flags.stdoutFile, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			return err
		}
		flags.writers = append(flags.writers, stdoutFile)
	}

	if net.ParseIP(flags.destinationIP) != nil {

		s := Podtracer.Streamer{}
		s.Init(flags.destinationIP, flags.destinationPort)

		flags.writers = append(flags.writers, s)
	}

	return nil
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

func Run(cliTool string) error {

	// Initializing podtracer will get all pod and container
	// information from kubeapi-server and container engine.
	containerContext := Podtracer.ContainerContext{}
	err := containerContext.Init(flags.targetPodName, flags.targetNamespace, flags.kubeconfigPath)
	if err != nil {
		return err
	}

	r, w := io.Pipe()

	go func() {
		splitArgs := strings.Split(flags.targetArgs, " ")
		cmd := exec.Command(cliTool, splitArgs...)
		cmd.Stdout = w
		cmd.Stderr = w
		Podtracer.Execute(cmd, &containerContext)
		w.Close()
	}()

	err = initWriters()
	if err != nil {
		return err
	}

	dstWriters := io.MultiWriter(flags.writers...)

	if _, err := io.Copy(dstWriters, r); err != nil {
		log.Fatal(err)
	}

	return nil
}
