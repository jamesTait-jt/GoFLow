package cmd

import (
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/k8s"
	"github.com/jamesTait-jt/goflow/cmd/cli/internal/service"
	"github.com/jamesTait-jt/goflow/pkg/log"
	"github.com/spf13/cobra"
)

var destroyCmd = &cobra.Command{
	Use:   "destroy",
	Short: "Destroy workerpool and redis containers",
	RunE: func(_ *cobra.Command, _ []string) error {
		conf, err := config.Get()
		if err != nil {
			return err
		}

		logger := log.NewConsoleLogger()

		clientset, err := k8s.NewClientset(conf.Kubernetes.ClusterURL, conf.Kubernetes.Namespace)
		if err != nil {
			return err
		}

		kubeDeploymentManager := k8s.NewDeploymentManager(conf, logger, clientset)

		deploymentService := service.NewDeploymentService(kubeDeploymentManager)

		return deploymentService.Destroy()
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
