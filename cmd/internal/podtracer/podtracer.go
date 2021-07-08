package podtracer

import (
	"context"
	"fmt"
	"time"

	"os/exec"

	"github.com/containernetworking/plugins/pkg/ns"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var c client.Client

func Run(tool string, targetArgs string, targetPod string, targetNamespace string) error {

	// TODO: setup client and get the pod itself here

	pod := &corev1.Pod{}
	_ = c.Get(context.Background(), client.ObjectKey{
		Namespace: targetNamespace,
		Name:      targetPod,
	}, pod)

	pid, err := getPid(*pod)
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
