package podtracer

import (
	"fmt"

	"github.com/containernetworking/plugins/pkg/ns"
)

const HostProcPath string = "/host/proc/"

type Runner interface {
	Run() error
}

type PID interface {
	GetContainerPID() int
}

func Execute(r Runner, pid PID) error {

	// Initializing NetNS for podtracer
	targetNS, err := ns.GetNS(HostProcPath + fmt.Sprintf("%d", pid.GetContainerPID()) + "/ns/net")
	if err != nil {
		return fmt.Errorf("error getting Pod network namespace: %v", err)
	}

	// Switching Linux Namespaces according to the path above
	// Only network namespace are implemented for now
	err = targetNS.Do(func(hostNs ns.NetNS) error {

		err := r.Run()
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
