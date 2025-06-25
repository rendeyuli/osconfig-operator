/*
Copyright 2025.

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

package controller

import (
	"context"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/log"

	sysconfigv1alpha1 "github.com/rendeyuli/osconfig-operator/api/v1alpha1"
)

// NodeConfigReconciler reconciles a NodeConfig object
type NodeConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=sysconfig.rendeyuli.osconfig,resources=nodeconfigs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=sysconfig.rendeyuli.osconfig,resources=nodeconfigs/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=sysconfig.rendeyuli.osconfig,resources=nodeconfigs/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the NodeConfig object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.4/pkg/reconcile
func (r *NodeConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = logf.FromContext(ctx)

	// TODO(user): your logic here

	// Step 1: 获取 NodeConfig 实例
	var config sysconfigv1alpha1.NodeConfig
	if err := r.Get(ctx, req.NamespacedName, &config); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Step 2: 解析 nodeSelector，筛选节点
	nodes := &corev1.NodeList{}
	selector := labels.Set(config.Spec.NodeSelector).AsSelector()
	if err := r.List(ctx, nodes, &client.ListOptions{LabelSelector: selector}); err != nil {
		return ctrl.Result{}, err
	}

	// Step 3: 生成 ConfigMap，包含配置信息
	cm := generateConfigMap(config)
	if err := r.Create(ctx, &cm); err != nil && !apierrors.IsAlreadyExists(err) {
		return ctrl.Result{}, err
	}

	//Step 4: 更新状态字段
	config.Status.AppliedNodes = extractNodeNames(nodes)
	if err := r.Status().Update(ctx, &config); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *NodeConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&sysconfigv1alpha1.NodeConfig{}).
		Named("nodeconfig").
		Complete(r)
}

// 从NodeConfig对象生成一个ConfigMap，准备下发给目标节点
func generateConfigMap(config sysconfigv1alpha1.NodeConfig) corev1.ConfigMap {
	return corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "node-config-" + config.Name,
			Namespace: config.Namespace,
			Labels: map[string]string{
				"managed-by": "osconfig-operator",
			},
		},
		Data: config.Spec.Data,
	}
}

// 从筛选到的NodeList中提取节点名列表，返回[]string用于填充status.AppliedNodes
func extractNodeNames(nodes *corev1.NodeList) []string {
	var names []string
	for _, node := range nodes.Items {
		names = append(names, node.Name)
	}
	return names
}
