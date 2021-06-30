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
	"fmt"

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
		fmt.Println("pcap called")
	},
}

func init() {
	rootCmd.AddCommand(pcapCmd)
	pcapCmd.Flags().StringP("tool", "t", "tcpdump", "selects the tool to run the packet capture")
	pcapCmd.Flags().StringP("namespace", "n", "default", "select the k8s namespace where Pod or Container is running")
	pcapCmd.Flags().StringP("args", "a", "", "string with all arguments and options for the selected tool")
	pcapCmd.Flags().String("pod", "", "Pod name for packet capture")
	pcapCmd.Flags().String("netns", "", "Network namespace path")
	pcapCmd.Flags().String("pid", "", "Container Process ID to capture packets from")
}
