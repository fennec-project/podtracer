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
	"github.com/spf13/cobra"
)

// pcapCmd represents the pcap command
var pcapCmd = &cobra.Command{
	Use:   "pcap",
	Short: "captures packets on the desired pod interface",
	Long: `
pcap captures packets by wrapping around packet capture tools and
running them on selected Pods, Containers, Linux namespaces or process IDs.
Once the it switches to the desired Linux namespace using whatever selector
argument was provided it runs the command according to its parameter string.
Some examples would be:
	
	podtracer pcap -t tcpdump -a "-i eth0 -w /pcap-data/mypcapfile.pcap" --pod mypod --namespace mypodk8snamespace
	podtracer pcap -t tcpdump -a "-i eth0 -w /pcap-data/mypcapfile.pcap" --netns <netns FD object>
	podtracer pcap -t tcpdump -a "-i eth0 -w /pcap-data/mypcapfile.pcap" --pid "1111"
	podtracer pcap -t tcpdump -a "-i eth0 -w /pcap-data/mypcapfile.pcap"
`,
	Run: func(cmd *cobra.Command, args []string) {

	},
}

var tool string
var args string
var namespace string
var pod string
var netns string
var pid string

func init() {
	rootCmd.AddCommand(pcapCmd)

	// Tool to be run by the podtracer application
	pcapCmd.Flags().StringVarP(&tool, "tool", "t", "tcpdump", "selects the tool to run the packet capture")

	// For now we're only wrapping the tools. So it's required to inform which tool to use.
	pcapCmd.MarkFlagRequired("tool")
	pcapCmd.Flags().StringVarP(&args, "args", "a", "", "string with all arguments and options for the selected tool")
	pcapCmd.MarkFlagRequired("args")

	pcapCmd.Flags().StringVarP(&namespace, "namespace", "n", "default", "select the k8s namespace where Pod or Container is running")
	pcapCmd.Flags().StringVar(&pod, "pod", "", "Pod name for packet capture")

	// Advanced options using the netns path or pid to get the network namespace file descriptor directly
	pcapCmd.Flags().StringVar(&netns, "netns", "", "Network namespace path")
	pcapCmd.Flags().StringVar(&pid, "pid", "", "Container Process ID to capture packets from")
}
