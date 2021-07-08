package podtracer

import (
	"context"
	"fmt"
	"log"
	"time"

	"os/exec"

	"os"

	"github.com/containernetworking/plugins/pkg/ns"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

func Run(tool string, targetArgs string, targetPod string, targetNamespace string) error {

	// TODO: setup client and get the pod itself here

	c, err := client.New(config.GetConfigOrDie(), client.Options{})
	if err != nil {
		fmt.Println("failed to create client")
		os.Exit(1)
	}

	pod := corev1.Pod{}
	_ = c.Get(context.Background(), client.ObjectKey{
		Namespace: targetNamespace,
		Name:      targetPod,
	}, &pod)

	pid, err := getPid(pod)
	if err != nil {
		return err
	}

	// Get the pods namespace object
	targetNS, err := ns.GetNS("/host/proc/" + pid + "/ns/net")
	if err != nil {
		return fmt.Errorf("error getting Pod network namespace: %v", err)
	}

	err = targetNS.Do(func(hostNs ns.NetNS) error {

		// _, err := netlink.LinkByName(ifName)
		// if err != nil {
		// 	return fmt.Errorf("interface could not be found: %v", err)
		// }

		// Running tcpdump on given Pod and Interface
		cmd := exec.Command(tool, targetArgs)

		err = cmd.Start()
		if err != nil {
			fmt.Println(err)
			return err
		}

		fmt.Printf("Starting tcpdump on pod %s at %v\n", pod.ObjectMeta.Name, time.Now())
		// TODO: get the stderr here - tcpdump is throwing exit status 1 but os.exec is throwing 0 how?
		err = cmd.Wait()
		log.Printf("Command %v finished with exit code: %v", tool+" "+targetArgs, err)

		// time.Sleep(time.Duration(duration) * time.Minute)

		// if err := cmd.Process.Kill(); err != nil {
		// 	fmt.Println(err)
		// 	return err
		// }

		// fmt.Printf("Stopping tcpdump on interface %s at %v\n", ifName, time.Now())

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}
