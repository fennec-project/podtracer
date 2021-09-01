package podtracer

import (
	"fmt"

	"io"
	"strings"

	"os/exec"

	"github.com/containernetworking/plugins/pkg/ns"
)

const HostProcPath string = "/host/proc/"

type Runner struct {

	// containerID for execution
	// podtracer executes commands and functions against containers
	// not pods
	ContainerID string

	// container process ID to identify its Linux namespaces
	ContainerPID string

	// Linux network namespace for podtracer to switch to
	NetNS ns.NetNS
}

func (r *Runner) Init(containerContext ContainerContext) error {

	// Storing container ID and PID for future auditing feature
	r.ContainerID = containerContext.GetContainerID()
	r.ContainerPID = containerContext.GetContainerPID()

	// Initializing NetNS for podtracer
	targetNS, err := ns.GetNS(HostProcPath + r.ContainerPID + "/ns/net")
	if err != nil {
		return fmt.Errorf("error getting Pod network namespace: %v", err)
	}
	r.NetNS = targetNS
	return nil
}

func (r *Runner) RunOSExec(f func() error) error {

	// Switching Linux Namespaces according to the path above
	// Only network namespace are implemented for now
	err := r.NetNS.Do(func(hostNs ns.NetNS) error {

		err := f()
		if err != nil {
			return err
		}

		splitArgs := strings.Split(targetArgs, " ")

		cmd := exec.Command(tool, splitArgs...)
		cmd.Stdout = io.MultiWriter(stdOutWriters...)
		cmd.Stderr = io.MultiWriter(stdErrWriters...)

		err := cmd.Run()
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

func (r *Runner) OsExec(stdOutWriters []io.Writer, stdErrWriters []io.Writer, tool string, targetArgs string) error {

	splitArgs := strings.Split(targetArgs, " ")

	cmd := exec.Command(tool, splitArgs...)
	cmd.Stdout = io.MultiWriter(stdOutWriters...)
	cmd.Stderr = io.MultiWriter(stdErrWriters...)

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error: %s\n %v", err.Error(), cmd.Stderr)
		return err
	}

	return nil
}
