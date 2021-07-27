# podtracer

podtracer is a cli tool inspired in the Linux command line utility called nsenter. nsenter is capable of running programs in selected Linux namespaces taking as input the file paths for the namespaces or process ids from where it derives the namespaces.

podtracer does the same but taking as input pod names and kubernetes namespaces in order to run Linux tools against the pods as targets. That enables tools such as tcpdump, iperf, tc, ip and others to be used against pods and containers directly without the itermediary process of finding their respective process IDs and subsequent namespace file paths.

It's designed to be run by an automated processes such as [snoopy-operator](https://github.com/fennec-project/snoopy-operator) that brings scalability to the next level. It's done by running podtracer as a container entry point for kubernetes jobs. Multiple jobs running tcpdump, for example, can capture packets on multiple pods at the same time and send the extracted data to a central data processing server.

The tool is also intended to be used as troubleshooting tool in a stand alone mode to inspect speficic containers inside a Kubernetes platform.

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
mv build/podtracer usr/bin/
```

# Usage

The use of podtracer for now is pretty much similar to nsenter but limited to the network namespace. This is why the only tool alowed for it to run is tcpdump. Other tools will be supported in the future. Check our [road map](docs/roadmap.md).

At this point it needs to be run from within the a kubernetes cluster for those trying to troubleshoot some workload. Check the docker file under the build/ folder. It can be used to create the troubleshoot/debug pod/container to use podtracer or podtracer can be added to any debug container image that may be already in use.

```
podtracer run < desired tool > -a < desired arguments > --pod < pod name > -n < k8s namespace name > --kubeconfig < path to kubeconfig or defaults to ~/.kube/config >
```

# Contribution

Public meetings, slack channel and YouTube channel you be published soon alongside with contribution guidelines.