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
	"github.com/fennec-project/podtracer/cmd/internal/podtracer"
	"github.com/spf13/cobra"
)

// runCmd represents the run command
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Runs the chosen tool on a target pod.",
	Long: `The run command allows running tools such as tcpdump, tshark, iperf and others
	to acquire network data and metrics for observability purposes without changing the pod.`,
	ValidArgs: []string{"tcpdump"},
	Args:      argFuncs(cobra.MaximumNArgs(1), cobra.OnlyValidArgs),
	Run: func(cmd *cobra.Command, args []string) {

		podtracer.Run(args[0], targetArgs, targetPod, targetNamespace)

	},
}

var targetArgs string
var targetPod string
var targetNamespace string

func init() {
	rootCmd.AddCommand(runCmd)

	runCmd.Flags().StringVarP(&targetArgs, "arguments", "a", "", "arguments to running cli utility.")
	runCmd.Flags().StringVar(&targetPod, "pod", "", "Target pod name.")
	runCmd.Flags().StringVarP(&targetNamespace, "namespace", "n", "", "Kubernetes namespace where the target pod is running")
	runCmd.MarkFlagRequired("pod")
	runCmd.MarkFlagRequired("namespace")
}

func argFuncs(funcs ...cobra.PositionalArgs) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		for _, f := range funcs {
			err := f(cmd, args)
			if err != nil {
				return err
			}
		}
		return nil
	}
}
