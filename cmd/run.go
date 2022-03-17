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
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"regexp"
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
	RunE: func(cmd *cobra.Command, args []string) error {

		r, w := io.Pipe()
		go cmdExec(args[0], w, cmd)

		done := make(chan bool)

		writers, err := buildWriters(cmd)
		if err != nil {
			return err
		}

		go sendData(r, done, writers)

		if cmd.Flag("timer").Value.String() != "" {
			waitForTimer(cmd.Flag("timer").Value.String())
		} else {
			// Wait for a done signal.
			<-done
		}
		return nil
	},
}

func buildWriters(cmd *cobra.Command) ([]io.Writer, error) {

	writers := []io.Writer{}

	if cmd.Flag("stdout").Value.String() == "true" {
		writers = append(writers, os.Stdout)
	}

	if cmd.Flag("file").Value.String() != "" {

		file, err := os.OpenFile(cmd.Flag("file").Value.String(), os.O_RDWR|os.O_CREATE, 0755)

		if err != nil {
			return nil, err
		}

		writers = append(writers, file)
	}

	if net.ParseIP(cmd.Flag("destination").Value.String()) != nil || (IsValidDomain(cmd.Flag("destination").Value.String()) && LookupTest(cmd.Flag("destination").Value.String()) == nil) {

		s := Podtracer.Streamer{}

		s.Init(cmd.Flag("destination").Value.String(),
			cmd.Flag("port").Value.String(),
			cmd.Flag("pod").Value.String())

		writers = append(writers, s)

	}

	return writers, nil
}

func waitForTimer(timer string) {
	// Setting up a timer with the desired duration for the command to run
	d, err := time.ParseDuration(timer)
	if err != nil {
		log.Fatal(err)
	}
	t := time.NewTimer(d)

	// Blocks and wait until timer channel receives a signal.
	endTime := <-t.C
	fmt.Printf("\n\n Command finished at %s ", endTime)
}

func cmdExec(cliTool string, w io.WriteCloser, cmd *cobra.Command) {

	// Initializing podtracer will get all pod and container
	// information from kubeapi-server and container engine.
	containerContext := Podtracer.ContainerContext{}
	err := containerContext.Init(cmd.Flag("pod").Value.String(), cmd.Flag("namespace").Value.String())
	if err != nil {
		log.Fatal(err)
	}

	// Podtracer.Execute switches to the targetPod's Linux
	// network namespace and executes a thread with a different
	// view of the system. This is why it needs its own goroutine.
	// All output from cmd is written through w that is connected
	// to r in the goroutine that will write data out.
	splitArgs := strings.Split(cmd.Flag("arguments").Value.String(), " ")
	targetCmd := exec.Command(cliTool, splitArgs...)
	targetCmd.Stdout = w
	targetCmd.Stderr = os.Stderr
	if err = Podtracer.Execute(targetCmd, &containerContext); err != nil {
		log.Fatal(err)
	}
	w.Close()
}

// This function collects data from the cmdExec and writes
// it to as many io.Writers as desired/available. For the moment
// we have only stdOut, file in container or gRPC streamer in
// package internal/streamer.go
func sendData(r io.Reader, done chan bool, writers []io.Writer) {

	multiWriter := io.MultiWriter(writers...)
	// This keeps running until r gets an EOF from w
	if _, err := io.Copy(multiWriter, r); err != nil {
		log.Fatal(err)
	}

	done <- true
}

func IsValidDomain(domain string) bool {
	var domainRegex = regexp.MustCompile(`(?:[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9][a-z0-9-]{0,61}[a-z0-9]`)
	return domainRegex.MatchString(domain)
}

// Check for a valid domain name or url
func LookupTest(domain string) error {

	var ErrAddressNotFound = errors.New("address not found or unreachable")

	addr, err := net.LookupHost(domain)
	if err != nil {
		return err
	}
	if len(addr) == 0 {
		return ErrAddressNotFound
	}
	return nil
}

func init() {

	// Flags for run
	runCmd.Flags().StringP("arguments", "a", "", "arguments to running cli utility.")
	runCmd.Flags().String("pod", "", "Target pod name.")
	runCmd.Flags().StringP("namespace", "n", "", "Kubernetes namespace where the target pod is running")
	runCmd.Flags().StringP("file", "f", "", "file path to save output data from the running tool.")
	runCmd.Flags().StringP("destination", "d", "", "Destination IP to where send stdout")
	runCmd.Flags().StringP("port", "p", "", "Destination port to where send stdout")
	runCmd.Flags().StringP("timer", "t", "", "It's expressed by a decima number and the time unit. Valid examples are 30s, 1h, 2h30m etc.")
	runCmd.Flags().BoolP("stdout", "s", false, "Use --stdout true for tool output on container's stdout.")

	// Required Flags
	runCmd.MarkFlagRequired("pod")
	runCmd.MarkFlagRequired("namespace")

	// Adding command to root
	rootCmd.AddCommand(runCmd)
}
