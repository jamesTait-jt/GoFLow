package run

import (
	"github.com/jamesTait-jt/goflow/cmd/goflow-cli/internal/config"
	"github.com/jamesTait-jt/goflow/pkg/log"
)

func Destroy(conf *config.Config, logger log.Logger) error {
	// stopLog := logger.Waiting("Connecting to the Kubernetes cluster")

	// kubeClient, err := kubernetes.New(
	// 	conf.Kubernetes.ClusterURL,
	// 	conf.Kubernetes.Namespace,
	// 	kubernetes.WithLogger(logger),
	// )
	// if err != nil {
	// 	stopLog("Failed connecting to kubernetes cluster", false)

	// 	return err
	// }

	// stopLog("Successfully connected to Kubernetes cluster", true)

	// logger.Info(fmt.Sprintf("🔥 Destroying namespace '%s'", conf.Kubernetes.Namespace))

	// // This will delete the namespace and everything contained within
	// err = kubeClient.DestroyNamespace(conf.Kubernetes.Namespace)
	// if err != nil {
	// 	return err
	// }

	// logger.Info("🔥 Destroying non-namespace scoped objects")

	// // Persistent volumes are not associated with a namespace so must be delete individually
	// err = kubeClient.DestroyPV(workerpool.HandlersPV(conf).Name)
	// if err != nil {
	// 	stopLog("Failed destroying objects", false)

	// 	return err
	// }

	// stopLog("Successfully destroyed objects", true)

	// logger.Success("GoFlow successfully destroyed")

	return nil
}
