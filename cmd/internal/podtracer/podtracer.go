package podtracer

import (
	"context"
	"fmt"
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

type Podtracer struct {
	Tool       string
	TargetArgs string
	Pod        string
	Namespace  string
	Kubeconfig string
}

func Log(msgLogLevel string, msg string, args ...interface{}) {
	systemLogLevel := os.Getenv("PODTRACER_LOGLEVEL")
	if systemLogLevel == "DEBUG" {
		logger.Printf("["+msgLogLevel+"] "+msg, args)
		return
	} else if msgLogLevel != "DEBUG" {
		logger.Printf("["+msgLogLevel+"] "+msg, args)
		return
	} else {
		return
	}
}

func (podtracer Podtracer) GetClient(kubeconfigPath string) (client.Client, error) {

	// TODO: link kubeconfigPath on client.new if empty default to ~/.kube/kubeconfig
	c, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}
	return c, nil
}

func (podtracer Podtracer) GetPod(targetPod string, targetNamespace string, kubeconfig string) (corev1.Pod, error) {

	c, err := podtracer.GetClient(kubeconfig)
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

func (podtracer Podtracer) Run(tool string, targetArgs string, targetPod string, targetNamespace string, kubeconfig string, stdoutFile string, stderrFile string) error {

	pod, err := podtracer.GetPod(targetPod, targetNamespace, kubeconfig)
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

		splitArgs := strings.Split(targetArgs, " ")

		logger.Printf("[INFO] Running %s: Pod %s Namespace %s \n\n", tool, targetPod, targetNamespace)

		// Creating list of writers for sending retrived data
		// TODO: needs to become its own function or method
		// planning to add kafka writer and others here

		stdOutWriters := []io.Writer{}
		stdOutWriters = append(stdOutWriters, os.Stdout)

		if stdoutFile != "" {
			stdoutFile, err := os.OpenFile(stdoutFile, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				return err
			}
			defer func() {
				if err := stdoutFile.Close(); err != nil {
					logger.Printf("Couldn't close file stdout.txt")
					return
				}
			}()
			stdOutWriters = append(stdOutWriters, stdoutFile)
		}

		stdErrWriters := []io.Writer{}
		stdErrWriters = append(stdErrWriters, os.Stderr)

		if stderrFile != "" {
			stderrFile, err := os.OpenFile(stderrFile, os.O_RDWR|os.O_CREATE, 0755)
			if err != nil {
				return err
			}
			defer func() {
				if err := stderrFile.Close(); err != nil {
					logger.Printf("Couldn't close file stdout.txt")
					return
				}
			}()
			stdErrWriters = append(stdErrWriters, stderrFile)
		}

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
