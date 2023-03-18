/*
Copyright 2023.

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

package controllers

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	pixiuv1alpha1 "github.com/gopixiu-io/mysql-operator/api/v1alpha1"
)

var logcl = log.Log.WithName("MysqlOperator")

// MysqlClusterReconciler reconciles a MysqlCluster object
type MysqlClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=pixiu.pixiu.io,resources=mysqlclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=pixiu.pixiu.io,resources=mysqlclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=pixiu.pixiu.io,resources=mysqlclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the MysqlCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.11.2/pkg/reconcile
func (r *MysqlClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	reqLogger := logcl.WithValues("namespace", req.Namespace, "MysqlOperator", req.Name)
	reqLogger.Info("=== Reconciling MysqlOperator")

	instance := &pixiuv1alpha1.MysqlCluster{}

	// check CRD
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil && errors.IsNotFound(err) {
		return ctrl.Result{}, nil
	} else if err != nil {
		return ctrl.Result{}, err
	}

	// complete CRD's spec configuration
	CompleteSpecConfig(instance)

	// main func
	// according CRD's status control sub-resource
	if instance.Status.Phase == "" {
		instance.Status.Phase = pixiuv1alpha1.PhasePending
	}

	switch instance.Status.Phase {
	// TODO: CR add schedule condition,
	// PENDING status turn to RUNNING status while
	case pixiuv1alpha1.PhasePending:
		reqLogger.Info("Phase is PENDING")
		reqLogger.Info("Scheduling....")

		instance.Status.Phase = pixiuv1alpha1.PhaseRunning
	case pixiuv1alpha1.PhaseRunning:
		reqLogger.Info("Phase is RUNNING")

		// create sub-resource
		// new pod-instance obj
		pod := newPodForCR(instance)
		if err := controllerutil.SetControllerReference(instance, pod, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		// check whether resource already exists
		found := &corev1.Pod{}
		err := r.Get(ctx, types.NamespacedName{
			Namespace: pod.Namespace,
			Name:      pod.Name,
		}, found)
		if err != nil && errors.IsNotFound(err) {
			// need create pod
			err = r.Create(ctx, pod)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else if err != nil && errors.IsAlreadyExists(err) {
			// TODO: add same pod info check
			// seems k8s alraedy can check whether pod name exists
		}

		// new svc-instance obj
		svc := newServiceForCR(instance)
		if err := controllerutil.SetControllerReference(instance, svc, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		foundsvc := &corev1.Service{}
		err = r.Get(ctx, types.NamespacedName{
			Name:      svc.Name,
			Namespace: svc.Namespace,
		}, foundsvc)
		if err != nil && errors.IsNotFound(err) {
			err = r.Create(ctx, svc)
			if err != nil {
				return ctrl.Result{}, err
			}
		} else if err != nil && errors.IsAlreadyExists(err) {
			// TODO: add same svc info check
			// seems k8s alraedy can check whether pod name exists
		}

		// TODO: add subresource watch func
	case pixiuv1alpha1.PhaseDone:
		reqLogger.Info("Phase is DONE")
		return ctrl.Result{}, nil
	}

	err = r.Status().Update(ctx, instance)
	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *MysqlClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&pixiuv1alpha1.MysqlCluster{}).
		Complete(r)
}

// create pod
func newPodForCR(cr *pixiuv1alpha1.MysqlCluster) *corev1.Pod {
	objMeta := metav1.ObjectMeta{
		Name:      cr.Spec.PodSpec.Name,
		Namespace: cr.Namespace,
		Labels: map[string]string{
			"label-pod": "MysqlCluster-pod",
		},
	}

	image := fmt.Sprintf("mysql:%s", cr.Spec.MysqlVersion)

	containerSpec := corev1.Container{
		Name:            cr.Spec.PodSpec.Name + "-container",
		Image:           image,
		ImagePullPolicy: corev1.PullPolicy(cr.Spec.ImagePullPolicy),
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: cr.Spec.PodSpec.Port,
			},
		},
		// TODO: get mysql pass and user from secret
		Env: []corev1.EnvVar{
			{
				Name:  "MYSQL_ROOT_PASSWORD",
				Value: cr.Spec.PodSpec.MysqlPass,
			},
		},
	}

	// create pod's volume from MysqlCluster.yaml
	var volume corev1.Volume
	var volumeMount corev1.VolumeMount

	if cr.Spec.PodSpec.VolumeType == pixiuv1alpha1.EmptyDir {
		volume.Name = cr.Spec.PodSpec.Name + "-emptydir-volume"
		volume.VolumeSource = corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}
	}

	if cr.Spec.PodSpec.VolumeType == pixiuv1alpha1.HostPath {
		volume.Name = cr.Spec.PodSpec.Name + "-hostpath-volume"
		hostPathType := corev1.HostPathDirectoryOrCreate
		volume.VolumeSource = corev1.VolumeSource{HostPath: &corev1.HostPathVolumeSource{
			Path: "/tmp/" + cr.Spec.PodSpec.Name + "/mysql",
			Type: &hostPathType,
		}}
	}

	volumeMount.Name = volume.Name
	volumeMount.MountPath = "/var/lib/mysql"

	containerSpec.VolumeMounts = []corev1.VolumeMount{volumeMount}

	return &corev1.Pod{
		ObjectMeta: objMeta,
		Spec: corev1.PodSpec{
			Containers:    []corev1.Container{containerSpec},
			Volumes:       []corev1.Volume{volume},
			RestartPolicy: corev1.RestartPolicyOnFailure,
		},
	}
}

// create svc
func newServiceForCR(cr *pixiuv1alpha1.MysqlCluster) *corev1.Service {
	objMeta := metav1.ObjectMeta{
		Name:      cr.Spec.ServiceSpec.Name,
		Namespace: cr.Namespace,
	}

	if cr.Spec.ServiceSpec.TargetPort != cr.Spec.PodSpec.Port {
		cr.Spec.ServiceSpec.TargetPort = cr.Spec.PodSpec.Port
	}

	svcSpec := corev1.ServiceSpec{
		Selector: map[string]string{
			"label-pod": "MysqlCluster-pod"},
		Ports: []corev1.ServicePort{
			{
				Protocol:   corev1.ProtocolTCP,
				TargetPort: intstr.IntOrString{IntVal: cr.Spec.ServiceSpec.TargetPort},
				Port:       cr.Spec.ServiceSpec.ServicePort,
			},
		},
	}

	svc := &corev1.Service{
		ObjectMeta: objMeta,
		Spec:       svcSpec,
	}

	if !cr.Spec.NodePortEnable {
		svc.Spec.Type = corev1.ServiceTypeClusterIP
		return svc
	}
	svc.Spec.Type = corev1.ServiceTypeNodePort
	return svc
}

// default para complete func
func CompleteSpecConfig(cr *pixiuv1alpha1.MysqlCluster) {
	CompletePodSpec(cr)
	CompleteSvcSpec(cr)
}

func CompletePodSpec(cr *pixiuv1alpha1.MysqlCluster) {
	podSpec := cr.Spec.PodSpec

	if podSpec.Name == "" {
		podSpec.Name = cr.Name + "-pod"
	}
	if podSpec.ImagePullPolicy == "" {
		podSpec.ImagePullPolicy = string(corev1.PullIfNotPresent)
	}

	podSpec.Port = 3306

	if podSpec.MysqlVersion == "" {
		podSpec.MysqlVersion = "5.7.26"
	}
	if podSpec.MysqlPass == "" {
		// TODO: add aandom password generation func
		podSpec.MysqlPass = "abcd1234"
	}

	switch podSpec.VolumeType {
	case "":
		podSpec.VolumeType = pixiuv1alpha1.HostPath
	case pixiuv1alpha1.EmptyDir:
		// no-op needed
	case pixiuv1alpha1.HostPath:
		// no-op needed
	default:
		podSpec.VolumeType = pixiuv1alpha1.HostPath
	}

	cr.Spec.PodSpec = podSpec
}

func CompleteSvcSpec(cr *pixiuv1alpha1.MysqlCluster) {
	svcSpc := cr.Spec.ServiceSpec

	if svcSpc.Name == "" {
		svcSpc.Name = cr.Name + "-service"
	}
	if svcSpc.TargetPort == 0 {
		svcSpc.TargetPort = 3306
	}
	if svcSpc.ServicePort == 0 {
		svcSpc.ServicePort = 3306
	}
	switch svcSpc.NodePortEnable {
	case true:
	case false:
	default:
		svcSpc.NodePortEnable = false
	}

	cr.Spec.ServiceSpec = svcSpc
}
