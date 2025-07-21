package kind

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/go-logr/logr"
	maptv1alpha1 "github.com/mapt-oss/mapt-operator/api/v1alpha1"
	"github.com/mapt-oss/mapt-operator/internal/metadata"
	"github.com/mapt-oss/mapt-operator/pkg/clusters"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

var _ = Describe("Kind Adapter (Unit Tests)", func() {
	var (
		kindObj    *maptv1alpha1.Kind
		fakeClient client.Client
		mockProv   *MockProvisioner
		testScheme *runtime.Scheme
		ctx        context.Context
	)

	const (
		KindName      = "test-kind"
		KindNamespace = "default"
	)

	BeforeEach(func() {
		testScheme = scheme.Scheme
		Expect(maptv1alpha1.AddToScheme(testScheme)).To(Succeed())
		ctx = context.Background()
		mockProv = &MockProvisioner{}

		kindObj = &maptv1alpha1.Kind{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "mapt.io/v1alpha1",
				Kind:       "Kind",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      KindName,
				Namespace: KindNamespace,
			},
			Spec: maptv1alpha1.KindSpec{
				OutputKubeconfigSecretName: "custom-secret",
			},
		}
	})

	JustBeforeEach(func() {
		fakeClient = fake.NewClientBuilder().
			WithScheme(testScheme).
			WithObjects(kindObj).
			WithStatusSubresource(kindObj).
			Build()
	})

	Describe("EnsureFinalizerIsAdded", func() {
		It("adds a finalizer", func() {
			adapter, err := newAdapter(ctx, fakeClient, kindObj, mockProv, logr.Discard())
			Expect(err).NotTo(HaveOccurred())

			_, err = adapter.EnsureFinalizerIsAdded()
			Expect(err).NotTo(HaveOccurred())

			var updated maptv1alpha1.Kind
			Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(kindObj), &updated)).To(Succeed())
			Expect(updated.Finalizers).To(ContainElement(metadata.KindFinalizer))
		})
	})

	Describe("EnsureFinalizersAreCalled", func() {
		It("skips finalizer if deletion timestamp is nil", func() {
			adapter, err := newAdapter(ctx, fakeClient, kindObj, mockProv, logr.Discard())
			Expect(err).NotTo(HaveOccurred())
			_, err = adapter.EnsureFinalizersAreCalled()
			Expect(err).NotTo(HaveOccurred())
		})

		It("skips deprovisioning if no provision ID", func() {
			now := metav1.Now()
			kindObj.ObjectMeta.DeletionTimestamp = &now
			kindObj.ObjectMeta.Finalizers = []string{metadata.KindFinalizer}

			fakeClient = fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects(kindObj).
				WithStatusSubresource(kindObj).
				Build()

			adapter, err := newAdapter(ctx, fakeClient, kindObj, mockProv, logr.Discard())
			Expect(err).NotTo(HaveOccurred())

			_, err = adapter.EnsureFinalizersAreCalled()
			Expect(err).NotTo(HaveOccurred())
		})

		It("removes finalizer after successful deprovision", func() {
			provisionID := "mock-provision-id"
			now := metav1.Now()
			kindObj.ObjectMeta.DeletionTimestamp = &now
			kindObj.ObjectMeta.Finalizers = []string{metadata.KindFinalizer}
			kindObj.Status.ProvisionId = &provisionID

			mockProv.MockDeprovision = func(cluster *clusters.MaptCluster) error {
				return nil
			}

			fakeClient = fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects(kindObj).
				WithStatusSubresource(kindObj).
				Build()

			adapter, err := newAdapter(ctx, fakeClient, kindObj, mockProv, logr.Discard())
			Expect(err).NotTo(HaveOccurred())

			_, err = adapter.EnsureFinalizersAreCalled()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("EnsureKindClusterIsProvisioned", func() {
		It("skips provisioning when already running", func() {
			kindObj.Status.Phase = maptv1alpha1.KindPhaseRunning
			fakeClient = fake.NewClientBuilder().
				WithScheme(testScheme).
				WithObjects(kindObj).
				Build()

			mockProv.MockProvision = func(cluster *clusters.MaptCluster) (*clusters.ClusterProvisionerMetadata, error) {
				Fail("Provision should not be called")
				return nil, nil
			}

			adapter, err := newAdapter(ctx, fakeClient, kindObj, mockProv, logr.Discard())
			Expect(err).NotTo(HaveOccurred())
			_, err = adapter.EnsureKindClusterIsProvisioned()
			Expect(err).NotTo(HaveOccurred())
		})

		It("handles provisioning failure", func() {
			mockProv.MockProvision = func(cluster *clusters.MaptCluster) (*clusters.ClusterProvisionerMetadata, error) {
				return &clusters.ClusterProvisionerMetadata{
					Type: clusters.KindClusterType,
					KindMetadata: &clusters.KindMetadata{
						Kubeconfig: "",
					},
				}, errors.New("provision failed")
			}

			adapter, err := newAdapter(ctx, fakeClient, kindObj, mockProv, logr.Discard())
			Expect(err).NotTo(HaveOccurred())
			_, err = adapter.EnsureKindClusterIsProvisioned()
			Expect(err).To(HaveOccurred())

			var updated maptv1alpha1.Kind
			Expect(fakeClient.Get(ctx, client.ObjectKeyFromObject(kindObj), &updated)).To(Succeed())
			Expect(updated.Status.Phase).To(Equal(maptv1alpha1.KindPhaseFailed))
			Expect(updated.Status.Message).To(ContainSubstring("provisioner returned empty kubeconfig"))
		})
	})
})
