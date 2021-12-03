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
	"time"

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

	// Timer sets how long a command should run
	// It's expressed by a decima number and the time unit
	// Valid examples are 30s, 1h, 2h30m etc.
	timer string

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
	runCmd.Flags().StringVarP(&flags.timer, "timer", "t", "", "It's expressed by a decima number and the time unit. Valid examples are 30s, 1h, 2h30m etc.")

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
		s.Init(flags.destinationIP, flags.destinationPort, flags.targetPodName)

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

	r, w := io.Pipe()
	go cmdExec(cliTool, w)

	done := make(chan bool)
	go sendData(r, done)

	if flags.timer != "" {
		waitForTimer()
	} else {
		// Wait for a done signal.
		<-done
	}
	return nil
}

func waitForTimer() {
	// Setting up a timer with the desired duration for the command to run
	d, err := time.ParseDuration(flags.timer)
	if err != nil {
		log.Fatal(err)
	}
	timer := time.NewTimer(d)

	// Blocks and wait until timer channel receives a signal.
	endTime := <-timer.C
	fmt.Printf("\n\n Command finished at %s ", endTime)
}

func cmdExec(cliTool string, w io.WriteCloser) {

	// Initializing podtracer will get all pod and container
	// information from kubeapi-server and container engine.
	containerContext := Podtracer.ContainerContext{}
	err := containerContext.Init(flags.targetPodName, flags.targetNamespace, flags.kubeconfigPath)
	if err != nil {
		log.Fatal(err)
	}

	// Podtracer.Execute switches to the targetPod's Linux
	// network namespace and executes a thread with a different
	// view of the system. This is why it needs its own goroutine.
	// All output from cmd is written through w that is connected
	// to r in the goroutine that will write data out.
	splitArgs := strings.Split(flags.targetArgs, " ")
	cmd := exec.Command(cliTool, splitArgs...)
	cmd.Stdout = w
	cmd.Stderr = w
	if err = Podtracer.Execute(cmd, &containerContext); err != nil {
		log.Fatal(err)
	}
	w.Close()
}

// This function collects data from the cmdExec and writes
// it to as many io.Writers as desired/available. For the moment
// we have only stdOut, file in container or gRPC streamer in
// package internal/streamer.go
func sendData(r io.Reader, done chan bool) {

	err := initWriters()
	if err != nil {
		log.Fatal(err)
	}

	dstWriters := io.MultiWriter(flags.writers...)
	// This keeps running until r gets an EOF from w
	if _, err := io.Copy(dstWriters, r); err != nil {
		log.Fatal(err)
	}

	done <- true
}
