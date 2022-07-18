package framework

import (
	"context"
	"fmt"

	"github.com/onsi/ginkgo"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (f *Framework) GetPvc(namespace string, name string) *corev1.PersistentVolumeClaim {

	pvc, err := f.KubeClientSet.CoreV1().PersistentVolumeClaims(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	assert.Nil(ginkgo.GinkgoT(), err, "getting pvc")
	assert.NotNil(ginkgo.GinkgoT(), pvc, "expected a pvc but none returned")
	return pvc
}

// EnsurePvc creates a pvc object and returns it, throws error if it already exists.
func (f *Framework) EnsurePvc(pvc *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {

	err := createPvcWithRetries(f.KubeClientSet, pvc.Namespace, pvc)
	assert.Nil(ginkgo.GinkgoT(), err, "creating pvc")

	pvcResult := f.GetPvc(pvc.Namespace, pvc.Name)
	return pvcResult
}

func createPvcWithRetries(c kubernetes.Interface, namespace string, obj *corev1.PersistentVolumeClaim) error {
	if obj == nil {
		return fmt.Errorf("object provided to create is empty")
	}
	createFunc := func() (bool, error) {
		_, err := c.CoreV1().PersistentVolumeClaims(namespace).Create(context.TODO(), obj, metav1.CreateOptions{})
		if err == nil {
			return true, nil
		}
		if k8sErrors.IsAlreadyExists(err) {
			return false, err
		}
		if isRetryableAPIError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to create object with non-retriable error: %v", err)
	}

	return retryWithExponentialBackOff(createFunc)
}

// UpdatePvc updates a pvc object and returns the updated object.
func (f *Framework) UpdatePvc(pvc *corev1.PersistentVolumeClaim) *corev1.PersistentVolumeClaim {
	err := updatePvcWithRetries(f.KubeClientSet, pvc.Namespace, pvc)
	assert.Nil(ginkgo.GinkgoT(), err, "updating pvc")
	pvcResult := f.GetPvc(pvc.Namespace, pvc.Name)
	return pvcResult
}

func updatePvcWithRetries(c kubernetes.Interface, namespace string, obj *corev1.PersistentVolumeClaim) error {
	if obj == nil {
		return fmt.Errorf("object provided to update is empty")
	}
	updateFunc := func() (bool, error) {
		_, err := c.CoreV1().PersistentVolumeClaims(namespace).Update(context.TODO(), obj, metav1.UpdateOptions{})
		if err == nil {
			return true, nil
		}
		if isRetryableAPIError(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to update object with non-retriable error: %v", err)
	}

	return retryWithExponentialBackOff(updateFunc)
}
