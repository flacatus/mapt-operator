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

package kind

import (
	"context"
	"errors"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	maptv1alpha1 "github.com/mapt-oss/mapt-operator/api/v1alpha1"
	"github.com/mapt-oss/mapt-operator/internal/metadata"
	"github.com/mapt-oss/mapt-operator/pkg/clusters"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("KindReconciler", func() {
	var (
		reconciler *KindReconciler
		mockProv   *MockProvisioner
		fakeClient client.Client
		testScheme *runtime.Scheme
		ctx        context.Context
		req        ctrl.Request
		kindObj    *maptv1alpha1.Kind
	)

	const (
		KindName      = "my-kind-cluster"
		KindNamespace = "default"
		SecretName    = "custom-secret"
		timeout       = time.Second * 10
		interval      = time.Millisecond * 250
	)

	BeforeEach(func() {
		testScheme = scheme.Scheme
		Expect(maptv1alpha1.AddToScheme(testScheme)).To(Succeed())
		ctx = context.Background()

		kindObj = &maptv1alpha1.Kind{
			ObjectMeta: metav1.ObjectMeta{
				Name:      KindName,
				Namespace: KindNamespace,
			},
			Spec: maptv1alpha1.KindSpec{
				OutputKubeconfigSecretName: SecretName,
			},
		}

		req = ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      KindName,
				Namespace: KindNamespace,
			},
		}

		mockProv = &MockProvisioner{}
	})

	JustBeforeEach(func() {
		// Initialize the fake client and the reconciler for each test.
		fakeClient = fake.NewClientBuilder().
			WithScheme(testScheme).
			WithObjects(kindObj).
			WithStatusSubresource(kindObj).
			Build()

		reconciler = &KindReconciler{
			Client:      fakeClient,
			Scheme:      testScheme,
			Provisioner: mockProv,
		}
	})

	Context("when reconciling a new Kind resource", func() {
		It("should successfully provision a cluster", func() {
			// 1. Setup
			tempFile, err := os.CreateTemp("", "kubeconfig-*.yaml")
			Expect(err).NotTo(HaveOccurred())
			DeferCleanup(os.Remove, tempFile.Name())
			Expect(tempFile.Close()).To(Succeed())

			mockProv.MockProvision = func(cluster *clusters.MaptCluster) (*clusters.ClusterProvisionerMetadata, error) {
				return &clusters.ClusterProvisionerMetadata{
					Type: clusters.KindClusterType,
					KindMetadata: &clusters.KindMetadata{
						Username:   "test-user",
						PrivateKey: "mock-private-key",
						Host:       "mock-host",
						Kubeconfig: tempFile.Name(),
						SpotPrice:  0.01,
					},
				}, nil
			}

			// 2. First reconcile: This should add the finalizer
			By("First reconcile: adding the finalizer")
			_, err = reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			var updatedKind maptv1alpha1.Kind
			Expect(fakeClient.Get(ctx, req.NamespacedName, &updatedKind)).To(Succeed())
			Expect(updatedKind.Finalizers).To(ContainElement(metadata.KindFinalizer))

			// 3. Second reconcile: This should trigger provisioning
			By("Second reconcile: provisioning the cluster")
			_, err = reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			// 4. Assert the final state
			By("Verifying the status is Running")
			Expect(fakeClient.Get(ctx, req.NamespacedName, &updatedKind)).To(Succeed())
			Expect(updatedKind.Status.Phase).To(Equal(maptv1alpha1.KindPhaseRunning))
			Expect(updatedKind.Status.ClusterReady).To(BeTrue())
			Expect(*updatedKind.Status.ProvisionId).To(Not(BeEmpty()))
		})

		It("should update the status to Failed if provisioning fails", func() {
			// 1. Setup
			mockProv.MockProvision = func(cluster *clusters.MaptCluster) (*clusters.ClusterProvisionerMetadata, error) {
				return &clusters.ClusterProvisionerMetadata{
					Type: clusters.KindClusterType,
					KindMetadata: &clusters.KindMetadata{
						Username:   "test-user",
						PrivateKey: "mock-private-key",
						Host:       "mock-host",
						Kubeconfig: "kubeconfig",
						SpotPrice:  0.01,
					},
				}, errors.New("pulumi exploded")
			}

			By("Second reconcile: attempting to provision")
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("pulumi exploded"))

			// 3. Assert the final state
			By("Verifying the status is Failed")
			var updatedKind maptv1alpha1.Kind
			Expect(fakeClient.Get(ctx, req.NamespacedName, &updatedKind)).To(Succeed())
			Expect(updatedKind.Status.Phase).To(Equal(maptv1alpha1.KindPhaseFailed))
			Expect(updatedKind.Status.Message).To(ContainSubstring("pulumi exploded"))
		})
	})

	Context("when reconciling a resource being deleted", func() {
		It("should deprovision and remove the finalizer", func() {
			// 1. Setup
			now := metav1.Now()
			kindObj.ObjectMeta.DeletionTimestamp = &now
			kindObj.ObjectMeta.Finalizers = []string{metadata.KindFinalizer}
			provisionId := "prov-id-to-delete"
			kindObj.Status.ProvisionId = &provisionId

			// Re-initialize the client WITH THE STATUS SUBRESOURCE ENABLED. This is the fix.
			fakeClient = fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects(kindObj).
				WithStatusSubresource(kindObj). // This line is critical
				Build()
			reconciler.Client = fakeClient

			deprovisionCalled := false
			mockProv.MockDeprovision = func(cluster *clusters.MaptCluster) error {
				deprovisionCalled = true
				return nil
			}

			// 2. Execute
			By("Reconciling a resource marked for deletion")
			_, err := reconciler.Reconcile(ctx, req)
			Expect(err).NotTo(HaveOccurred())

			// 3. Assert
			By("Verifying deprovision was called and the resource is gone")
			Expect(deprovisionCalled).To(BeTrue())
			err = fakeClient.Get(ctx, req.NamespacedName, &maptv1alpha1.Kind{})
			Expect(apierrors.IsNotFound(err)).To(BeTrue())
		})
	})
})
