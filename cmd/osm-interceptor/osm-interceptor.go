/*
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

// Package main implements osm intercepter.
package main

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/openservicemesh/osm/pkg/cni/config"
	"github.com/openservicemesh/osm/pkg/cni/controller"
	"github.com/openservicemesh/osm/pkg/cni/controller/helpers"
	cniserver "github.com/openservicemesh/osm/pkg/cni/controller/server"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "osm-interceptor",
	Short: "Use eBPF to speed up your Service Mesh like crossing an Einstein-Rosen Bridge.",
	Long:  `Use eBPF to speed up your Service Mesh like crossing an Einstein-Rosen Bridge.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := helpers.LoadProgs(config.Debug, config.Skip); err != nil {
			return fmt.Errorf("failed to load ebpf programs: %v", err)
		}

		stop := make(chan struct{}, 1)
		cniReady := make(chan struct{}, 1)
		if config.EnableCNI {
			s := cniserver.NewServer(path.Join(config.HostVarRun, "osm-cni.sock"),
				"/sys/fs/bpf", cniReady, stop)
			if err := s.Start(); err != nil {
				log.Fatal(err)
				return err
			}
		}
		// todo: wait for stop
		if err := controller.Run(config.DisableWatcher, config.Skip, cniReady, stop); err != nil {
			log.Fatal(err)
			return err
		}
		return nil
	},
}

func execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func main() {
	execute()
}

func init() {
	// Setup log format
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp:       false,
		FullTimestamp:          true,
		DisableLevelTruncation: true,
		DisableColors:          true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			fs := strings.Split(f.File, "/")
			filename := fs[len(fs)-1]
			ff := strings.Split(f.Function, "/")
			_f := ff[len(ff)-1]
			return fmt.Sprintf("%s()", _f), fmt.Sprintf("%s:%d", filename, f.Line)
		},
	})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
	log.SetReportCaller(true)

	// Get some flags from commands
	rootCmd.PersistentFlags().BoolVarP(&config.Debug, "debug", "d", false, "Debug mode")
	rootCmd.PersistentFlags().BoolVarP(&config.Skip, "skip", "s", false, "Skip init bpf")
	rootCmd.PersistentFlags().BoolVarP(&config.DisableWatcher, "disableWatcher", "w", false, "disable Pod watcher")
	rootCmd.PersistentFlags().BoolVarP(&config.IsKind, "kind", "k", false, "Enable when Kubernetes is running in Kind")
	_ = rootCmd.PersistentFlags().MarkDeprecated("ips-file", "no need to collect node IPs")
	rootCmd.PersistentFlags().BoolVar(&config.EnableCNI, "cni-mode", false, "Enable CNI plugin")
	rootCmd.PersistentFlags().StringVar(&config.HostProc, "host-proc", "/host/proc", "/proc mount path")
	rootCmd.PersistentFlags().StringVar(&config.CNIBinDir, "cni-bin-dir", "/host/opt/cni/bin", "/opt/cni/bin mount path")
	rootCmd.PersistentFlags().StringVar(&config.CNIConfigDir, "cni-config-dir", "/host/etc/cni/net.d", "/etc/cni/net.d mount path")
	rootCmd.PersistentFlags().StringVar(&config.HostVarRun, "host-var-run", "/host/var/run", "/var/run mount path")
	rootCmd.PersistentFlags().StringVar(&config.KubeConfig, "kubeconfig", "", "Kubernetes configuration file")
	rootCmd.PersistentFlags().StringVar(&config.Context, "kubecontext", "", "The name of the kube config context to use")
	rootCmd.PersistentFlags().BoolVar(&config.EnableHotRestart, "enable-hot-restart", false, "enable hot restart")
}
