package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	apiv1 "k8s.io/api/core/v1"
	k8serr "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type KubeClient struct {
	ctx       context.Context
	client    *kubernetes.Clientset
	namespace string
}

func New(clusterURL, kubeNamespace string) (*KubeClient, error) {
	var kubeConfPath string
	if home := homedir.HomeDir(); home != "" {
		kubeConfPath = filepath.Join(home, ".kube", "config")
	} else {
		return nil, errors.New("could not find .kube/config file in home directory")
	}

	kubeConf, err := clientcmd.BuildConfigFromFlags(clusterURL, kubeConfPath)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeConf)
	if err != nil {
		return nil, err
	}

	client := &KubeClient{
		ctx:       context.Background(),
		client:    clientset,
		namespace: kubeNamespace,
	}

	err = client.createNamespaceIfNotExists()
	if err != nil {
		return nil, err
	}

	return client, nil
}

func (k *KubeClient) CreateOrReplaceDeployment(deployment *appsv1.Deployment) error {
	deploymentsClient := k.client.AppsV1().Deployments(k.namespace)

	result, err := deploymentsClient.Create(k.ctx, deployment, metav1.CreateOptions{})
	if err == nil {
		fmt.Printf("Created deployment %q.\n", result.GetObjectMeta().GetName())

		return nil
	}

	if !k8serr.IsAlreadyExists(err) {
		return err
	}

	fmt.Println("Deployment existed - replacing")

	_, err = deploymentsClient.Update(k.ctx, deployment, metav1.UpdateOptions{})
	if err != nil {
		return err
	}

	fmt.Println("Deployment replaced successfully!")

	return err
}

func (k *KubeClient) CreateOrReplaceService(service *apiv1.Service) error {
	servicesClient := k.client.CoreV1().Services(k.namespace)

	_, err := servicesClient.Get(k.ctx, service.Name, metav1.GetOptions{})

	if err == nil {
		fmt.Println("Service existed - replacing")

		if _, err = servicesClient.Update(k.ctx, service, metav1.UpdateOptions{}); err != nil {
			return err
		}

		fmt.Println("Service replaced successfully!")

		return nil
	}

	if !k8serr.IsNotFound(err) {
		return err
	}

	_, err = servicesClient.Create(k.ctx, service, metav1.CreateOptions{})

	return err
}

func (k *KubeClient) CreateOrReplacePV(pv *apiv1.PersistentVolume) error {
	pvClient := k.client.CoreV1().PersistentVolumes()

	_, err := pvClient.Create(k.ctx, pv, metav1.CreateOptions{})

	// if err == nil {
	// 	return nil
	// }

	// if !k8serr.IsAlreadyExists(err) {
	// 	return err
	// }

	// _, err = pvClient.Update(k.ctx, pv, metav1.UpdateOptions{})

	return err
}

func (k *KubeClient) CreateOrReplacePVC(pvc *apiv1.PersistentVolumeClaim) error {
	pvcClient := k.client.CoreV1().PersistentVolumeClaims(k.namespace)

	_, err := pvcClient.Create(k.ctx, pvc, metav1.CreateOptions{})

	// if err == nil {
	// 	return nil
	// }

	// if !k8serr.IsAlreadyExists(err) {
	// 	return err
	// }

	// _, err = pvcClient.Update(k.ctx, pvc, metav1.UpdateOptions{})

	return err
}

func (k *KubeClient) CreateOrReplaceAndRunJob(job *batchv1.Job) error {
	jobsClient := k.client.BatchV1().Jobs(k.namespace)

	createdJob, err := jobsClient.Create(k.ctx, job, metav1.CreateOptions{})
	if err == nil {
		fmt.Printf("Created Job: %s\n", createdJob.Name)
		return nil
	}

	if !k8serr.IsAlreadyExists(err) {
		return err
	}

	fmt.Println("Job already exists, recreating...")

	_, err = jobsClient.Update(k.ctx, job, metav1.UpdateOptions{})

	if err != nil {
		return err
	}

	fmt.Println("Job recreated successfully!")

	// for {
	// 	job, err := jobsClient.Get(k.ctx, job.Name, metav1.GetOptions{})
	// 	if err != nil {
	// 		return err
	// 	}

	// 	if job.Status.Succeeded > 0 {
	// 		fmt.Println("Job completed successfully.")
	// 		break
	// 	} else if job.Status.Failed > 0 {
	// 		log.Fatalf("Job failed.")
	// 	}

	// 	time.Sleep(2 * time.Second)
	// }

	return nil
}

func (k *KubeClient) createNamespaceIfNotExists() error {
	_, err := k.client.CoreV1().Namespaces().Get(k.ctx, k.namespace, metav1.GetOptions{})
	if err == nil {
		return nil
	}

	if !k8serr.IsNotFound(err) {
		return fmt.Errorf("failed to get namespace: %v", err)
	}

	namespace := &apiv1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: k.namespace,
		},
	}

	_, err = k.client.CoreV1().Namespaces().Create(k.ctx, namespace, metav1.CreateOptions{})

	if err != nil {
		return fmt.Errorf("failed to create namespace: %v", err)
	}

	fmt.Printf("Namespace '%s' created successfully.\n", k.namespace)

	return nil
}

func CreateGoFlowGRPCService() error {
	return nil
}

func int32Ptr(i int32) *int32 { return &i }
