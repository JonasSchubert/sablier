package kubernetes

import (
	"context"
	"errors"
	"fmt"
	"github.com/acouvreur/sablier/app/discovery"
	"github.com/acouvreur/sablier/app/providers"
	"strconv"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"

	"github.com/acouvreur/sablier/app/instance"
	providerConfig "github.com/acouvreur/sablier/config"
	log "github.com/sirupsen/logrus"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

// Interface guard
var _ providers.Provider = (*KubernetesProvider)(nil)

type Config struct {
	OriginalName string
	Kind         string // deployment or statefulset
	Namespace    string
	Name         string
	Replicas     int32
}

type Workload interface {
	GetScale(ctx context.Context, workloadName string, options metav1.GetOptions) (*autoscalingv1.Scale, error)
	UpdateScale(ctx context.Context, workloadName string, scale *autoscalingv1.Scale, opts metav1.UpdateOptions) (*autoscalingv1.Scale, error)
}

func (provider *KubernetesProvider) convertName(name string) (*Config, error) {
	s := strings.Split(name, provider.delimiter)
	if len(s) < 4 {
		return nil, errors.New("invalid name should be: kind" + provider.delimiter + "namespace" + provider.delimiter + "name" + provider.delimiter + "replicas")
	}
	replicas, err := strconv.Atoi(s[3])
	if err != nil {
		return nil, err
	}

	return &Config{
		OriginalName: name,
		Kind:         s[0],
		Namespace:    s[1],
		Name:         s[2],
		Replicas:     int32(replicas),
	}, nil
}

type KubernetesProvider struct {
	Client    kubernetes.Interface
	delimiter string
}

func NewKubernetesProvider(providerConfig providerConfig.Kubernetes) (*KubernetesProvider, error) {
	kubeclientConfig, err := rest.InClusterConfig()

	kubeclientConfig.QPS = providerConfig.QPS
	kubeclientConfig.Burst = providerConfig.Burst

	log.Debug(fmt.Sprintf("Provider configuration:  QPS=%v, Burst=%v", kubeclientConfig.QPS, kubeclientConfig.Burst))

	if err != nil {
		return nil, err
	}
	client, err := kubernetes.NewForConfig(kubeclientConfig)
	if err != nil {
		return nil, err
	}

	return &KubernetesProvider{
		Client:    client,
		delimiter: providerConfig.Delimiter,
	}, nil

}

func (provider *KubernetesProvider) Start(ctx context.Context, name string) error {
	config, err := provider.convertName(name)
	if err != nil {
		return err
	}

	return provider.scale(ctx, config, config.Replicas)
}

func (provider *KubernetesProvider) Stop(ctx context.Context, name string) error {
	config, err := provider.convertName(name)
	if err != nil {
		return err
	}

	return provider.scale(ctx, config, 0)

}

func (provider *KubernetesProvider) GetGroups(ctx context.Context) (map[string][]string, error) {
	deployments, err := provider.Client.AppsV1().Deployments(core_v1.NamespaceAll).List(ctx, metav1.ListOptions{
		LabelSelector: discovery.LabelEnable,
	})

	if err != nil {
		return nil, err
	}

	groups := make(map[string][]string)
	for _, deployment := range deployments.Items {
		groupName := deployment.Labels[discovery.LabelGroup]
		if len(groupName) == 0 {
			groupName = discovery.LabelGroupDefaultValue
		}

		group := groups[groupName]
		parsed := DeploymentName(deployment, ParseOptions{Delimiter: provider.delimiter})
		group = append(group, parsed.Original)
		groups[groupName] = group
	}

	statefulSets, err := provider.Client.AppsV1().StatefulSets(core_v1.NamespaceAll).List(ctx, metav1.ListOptions{
		LabelSelector: discovery.LabelEnable,
	})

	if err != nil {
		return nil, err
	}

	for _, statefulSet := range statefulSets.Items {
		groupName := statefulSet.Labels[discovery.LabelGroup]
		if len(groupName) == 0 {
			groupName = discovery.LabelGroupDefaultValue
		}

		group := groups[groupName]
		parsed := StatefulSetName(statefulSet, ParseOptions{Delimiter: provider.delimiter})
		group = append(group, parsed.Original)
		groups[groupName] = group
	}

	return groups, nil
}

func (provider *KubernetesProvider) scale(ctx context.Context, config *Config, replicas int32) error {
	var workload Workload

	switch config.Kind {
	case "deployment":
		workload = provider.Client.AppsV1().Deployments(config.Namespace)
	case "statefulset":
		workload = provider.Client.AppsV1().StatefulSets(config.Namespace)
	default:
		return fmt.Errorf("unsupported kind \"%s\" must be one of \"deployment\", \"statefulset\"", config.Kind)
	}

	s, err := workload.GetScale(ctx, config.Name, metav1.GetOptions{})
	if err != nil {
		return err
	}

	s.Spec.Replicas = replicas
	_, err = workload.UpdateScale(ctx, config.Name, s, metav1.UpdateOptions{})

	return err
}

func (provider *KubernetesProvider) GetState(ctx context.Context, name string) (instance.State, error) {
	config, err := provider.convertName(name)
	if err != nil {
		return instance.UnrecoverableInstanceState(name, err.Error(), int(config.Replicas))
	}

	switch config.Kind {
	case "deployment":
		return provider.getDeploymentState(ctx, config)
	case "statefulset":
		return provider.getStatefulsetState(ctx, config)
	default:
		return instance.UnrecoverableInstanceState(config.OriginalName, fmt.Sprintf("unsupported kind \"%s\" must be one of \"deployment\", \"statefulset\"", config.Kind), int(config.Replicas))
	}
}

func (provider *KubernetesProvider) getDeploymentState(ctx context.Context, config *Config) (instance.State, error) {
	d, err := provider.Client.AppsV1().Deployments(config.Namespace).
		Get(ctx, config.Name, metav1.GetOptions{})

	if err != nil {
		return instance.ErrorInstanceState(config.OriginalName, err, int(config.Replicas))
	}

	if *d.Spec.Replicas == d.Status.ReadyReplicas {
		return instance.ReadyInstanceState(config.OriginalName, int(config.Replicas))
	}

	return instance.NotReadyInstanceState(config.OriginalName, int(d.Status.ReadyReplicas), int(config.Replicas))
}

func (provider *KubernetesProvider) getStatefulsetState(ctx context.Context, config *Config) (instance.State, error) {
	ss, err := provider.Client.AppsV1().StatefulSets(config.Namespace).
		Get(ctx, config.Name, metav1.GetOptions{})

	if err != nil {
		return instance.ErrorInstanceState(config.OriginalName, err, int(config.Replicas))
	}

	if *ss.Spec.Replicas == ss.Status.ReadyReplicas {
		return instance.ReadyInstanceState(config.OriginalName, int(config.Replicas))
	}

	return instance.NotReadyInstanceState(config.OriginalName, int(ss.Status.ReadyReplicas), int(config.Replicas))
}

func (provider *KubernetesProvider) NotifyInstanceStopped(ctx context.Context, instance chan<- string) {

	informer := provider.watchDeployents(instance)
	go informer.Run(ctx.Done())
	informer = provider.watchStatefulSets(instance)
	go informer.Run(ctx.Done())
}

func (provider *KubernetesProvider) watchDeployents(instance chan<- string) cache.SharedIndexInformer {
	handler := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			newDeployment := new.(*appsv1.Deployment)
			oldDeployment := old.(*appsv1.Deployment)

			if newDeployment.ObjectMeta.ResourceVersion == oldDeployment.ObjectMeta.ResourceVersion {
				return
			}

			if *newDeployment.Spec.Replicas == 0 {
				parsed := DeploymentName(*newDeployment, ParseOptions{Delimiter: provider.delimiter})
				instance <- parsed.Original
			}
		},
		DeleteFunc: func(obj interface{}) {
			deletedDeployment := obj.(*appsv1.Deployment)
			parsed := DeploymentName(*deletedDeployment, ParseOptions{Delimiter: provider.delimiter})
			instance <- parsed.Original
		},
	}
	factory := informers.NewSharedInformerFactoryWithOptions(provider.Client, 2*time.Second, informers.WithNamespace(core_v1.NamespaceAll))
	informer := factory.Apps().V1().Deployments().Informer()

	informer.AddEventHandler(handler)
	return informer
}

func (provider *KubernetesProvider) watchStatefulSets(instance chan<- string) cache.SharedIndexInformer {
	handler := cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(old, new interface{}) {
			newStatefulSet := new.(*appsv1.StatefulSet)
			oldStatefulSet := old.(*appsv1.StatefulSet)

			if newStatefulSet.ObjectMeta.ResourceVersion == oldStatefulSet.ObjectMeta.ResourceVersion {
				return
			}

			if *newStatefulSet.Spec.Replicas == 0 {
				parsed := StatefulSetName(*newStatefulSet, ParseOptions{Delimiter: provider.delimiter})
				instance <- parsed.Original
			}
		},
		DeleteFunc: func(obj interface{}) {
			deletedStatefulSet := obj.(*appsv1.StatefulSet)
			parsed := StatefulSetName(*deletedStatefulSet, ParseOptions{Delimiter: provider.delimiter})
			instance <- parsed.Original
		},
	}
	factory := informers.NewSharedInformerFactoryWithOptions(provider.Client, 2*time.Second, informers.WithNamespace(core_v1.NamespaceAll))
	informer := factory.Apps().V1().StatefulSets().Informer()

	informer.AddEventHandler(handler)
	return informer
}