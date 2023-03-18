package controllers

import (
	"fmt"
	pixiuv1alpha1 "github.com/gopixiu-io/mysql-operator/api/v1alpha1"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCompleteSpecConfig(t *testing.T) {
	cr := &pixiuv1alpha1.MysqlCluster{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "pixiu.pixiu.io/v1alpha1",
			Kind:       "MysqlCluster",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "mysqlcluster",
			Namespace: "default",
		},
		Spec: pixiuv1alpha1.MysqlClusterSpec{
			PodSpec:     pixiuv1alpha1.PodSpec{},
			ServiceSpec: pixiuv1alpha1.ServiceSpec{},
		},
	}
	CompleteSpecConfig(cr)
	fmt.Println(cr.Spec.PodSpec)
	fmt.Println(cr.Spec.ServiceSpec)

}
