# podtracer

### Community Meetings

Public calendar with regular meetings for 2022 will be available soon. We'll be back in January 14! Happy Holidays!

---

podtracer is a cli tool inspired in the Linux command line utility called nsenter. nsenter is capable of running programs in selected Linux namespaces taking as input the file paths for the namespaces or process ids from where it derives the namespaces.

podtracer does the same but taking as input pod names and kubernetes namespaces in order to run Linux tools against the pods as targets. That enables tools such as tcpdump, iperf, tc, ip and others to be used against pods and containers directly without the itermediary process of finding their respective process IDs and subsequent namespace file paths.

It's designed to be run by automated processes such as [snoopy-operator](https://github.com/fennec-project/snoopy-operator) that brings scalability to the next level. It's done by running podtracer as a container entry point for kubernetes jobs. Multiple jobs running tcpdump, for example, can capture packets on multiple pods at the same time and send the extracted data to a central data processing server.

The tool is also intended to be used as troubleshooting tool in a stand alone mode to inspect specific containers inside a Kubernetes platform. Check the usage section.

# Requirements

Since podtracer needs to run inside a kubernetes cluster node it needs to be copied over a troubleshooting container image. This tool requires privileges on the container it's being run. It needs to run as root for the moment although not all Linux capabilities from root are used. In the future this code should be capability aware in order to avoid running with root user.

For OpenShift additional RBAC yaml files are necessary to allow the privileged container to run as root. Among them a role and role binding using the privileged SCC. Don't forget to create the container spec with a SecurityContext saying `privileged: true`.

Remember that at this point podtracer only supports CRIO. Other container engines are on our road map.

Important to note as well that when troubleshooting with podtracer it needs to run on the same worker node as the pod being analyzed.

# Build and Install

There is no tagged release for podtracer at this point in the development but it can be built with go 1.15 or later like below. It's important to notice that at this point it doesn't support other operating systems other than Linux. It's under our plan to have a stand alone release for MacOS users reaching out to remote k8s clusters.

- Clone the project:
```
git clone https://github.com/fennec-project/podtracer
```

- Build the binary:
```
go build -o build/podtracer
```

- The binary can be moved to any place in your $PATH
```
mv build/podtracer /usr/bin/
```

# Usage

The use of podtracer for now is pretty much similar to nsenter but limited to the network namespace. That means that tools that require other Linux namespaces from the target pods won't work as expected. Networking tools should mostly work well. Other categories of tools will be supported in the future. Check our [road map](docs/roadmap.md).

At this point podtracer needs to be run from within the a kubernetes cluster for those trying to troubleshoot some workload. Here is how you can quickly spin up a troubleshooting pod with podtracer installed and get started:

```
make podtracer-deploy
```
That applies an one replica deployment with a Pod just waiting for the user to open a shell on that Pod and run podtracer from it.

In order to run podtracer itself:

```
podtracer run < desired tool > -a < desired arguments > --pod < pod name > -n < k8s namespace name >
```
Example:
```
potracer run tcpdump -a "-i eth0 -c 100 -w /pcap/test.pcap" --pod mypodname -n mynamespace
```

