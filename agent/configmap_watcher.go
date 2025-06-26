package main

import (
	"context"
	"log"

	corev1 "k8s.io/api/core/v1"
	//"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
)

func WatchConfigMap(ctx context.Context) error {
	// 使用集群内配置创建 Kubernetes 客户端
	config, err := rest.InClusterConfig()
	if err != nil {
		return err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	factory := informers.NewSharedInformerFactoryWithOptions(clientset, 0)
	informer := factory.Core().V1().ConfigMaps().Informer()

	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    onChanged,
		UpdateFunc: func(oldObj, newObj interface{}) { onChanged(newObj) },
	})

	factory.Start(ctx.Done())
	if !cache.WaitForCacheSync(ctx.Done(), informer.HasSynced) {
		return nil
	}

	<-ctx.Done()
	return nil
}

func onChanged(obj interface{}) {
	cm, ok := obj.(*corev1.ConfigMap)
	if !ok {
		log.Println("not a configmap")
		return
	}
	log.Printf("ConfigMap updated: %s\n", cm.Name)

	ApplyConfig(cm)
}
