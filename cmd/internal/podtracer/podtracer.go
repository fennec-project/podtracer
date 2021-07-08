package podtracer

import (
	"fmt"
	"time"

	"os/exec"

	corev1 "k8s.io/api/core/v1"

	"github.com/containernetworking/plugins/pkg/ns"
)

func Run(tool string, targetPod corev1.Pod, targetArgs string) error {

	pid, err := getPid(targetPod)
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

		fmt.Printf("Starting tcpdump on pod %s at %v\n", targetPod.ObjectMeta.Name, time.Now())

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
