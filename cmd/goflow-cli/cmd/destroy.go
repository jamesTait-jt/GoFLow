package cmd

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/k8s"
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/service"
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

		kubeOperator, err := k8s.NewOperator()
		if err != nil {
			return err
		}

		kubeDeployer, err := k8s.NewDeployer(conf, logger, clientset, kubeOperator)
		if err != nil {
			return err
		}

		deploymentService := service.NewDeploymentService(kubeDeployer)

		return deploymentService.Destroy()
	},
}

func init() {
	rootCmd.AddCommand(destroyCmd)
}
