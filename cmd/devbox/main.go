package main

import (
	"flag"
	"os"

	"github.com/beclab/devbox/pkg/api/server"
	"github.com/beclab/devbox/pkg/store/db"
	"github.com/beclab/devbox/pkg/webhook"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
)

func main() {

	klog.InitFlags(nil)
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)

	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)

	pflag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	rootCmd := &cobra.Command{
		Use:   "devbox",
		Short: "DevBox",
		Long:  `The DevBox is a Olares App dev tools`,
	}

	serverCmd := &cobra.Command{
		Use:   "server",
		Short: "DevBox server",
		Long:  `Start the DevBox server`,
		Run: func(cmd *cobra.Command, args []string) {
			klog.Info("DevBox starting ... ")

			dbOp := db.NewDbOperator()
			defer func() {
				if err := recover(); err != nil { //catch
					klog.Errorf("Exception: %v", err)
					e := dbOp.Close()
					if e != nil {
						klog.Error("close db error, ", err)
					}
					os.Exit(1)
				}
			}()

			s := server.NewServer(dbOp)

			// err := s.Init()
			// if err != nil {
			// 	panic(err)
			// }

			s.Start()

			klog.Info("db closed ", dbOp.Close())
			klog.Info("DevBox shutdown ")
		},
	}

	cleanCmd := &cobra.Command{
		Use:   "clean",
		Short: "DevBox clean",
		Long:  `clean the DevBox webhooks`,
		Run: func(cmd *cobra.Command, args []string) {
			klog.Info("clean DevBox webhooks ")

			config := ctrl.GetConfigOrDie()
			wh := webhook.Webhook{KubeClient: kubernetes.NewForConfigOrDie(config)}
			runtime.Must(wh.DeleteDevContainerMutatingWebhook())
			runtime.Must(wh.DeleteImageManagerMutatingWebhook())
		},
	}

	rootCmd.AddCommand(serverCmd, cleanCmd)

	if err := rootCmd.Execute(); err != nil {
		klog.Fatalln(err)
	}
}
