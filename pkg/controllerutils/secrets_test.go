package controllerutils

import (
	"context"
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("CreateGeneratedSecret", func() {
	var (
		ctx        context.Context
		scheme     *runtime.Scheme
		fakeClient client.Client
		owner      *corev1.ConfigMap
		secretData map[string][]byte
	)

	BeforeEach(func() {
		ctx = context.Background()
		scheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(scheme)).To(Succeed())

		owner = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-owner",
				Namespace: "default",
				UID:       uuid.NewUUID(),
			},
		}

		fakeClient = fake.NewClientBuilder().
			WithScheme(scheme).
			WithObjects(owner).
			Build()

		secretData = map[string][]byte{
			"key": []byte("value"),
		}
	})

	It("successfully creates a generated secret", func() {
		secretName, err := CreateGeneratedSecret(ctx, fakeClient, scheme, secretData, owner)
		Expect(err).NotTo(HaveOccurred())
		Expect(secretName).To(ContainSubstring("test-owner"))

		// Confirm it's created in cluster
		var created corev1.Secret
		Expect(fakeClient.Get(ctx, types.NamespacedName{Name: secretName, Namespace: "default"}, &created)).To(Succeed())
		Expect(created.Data["key"]).To(Equal([]byte("value")))
		Expect(created.OwnerReferences).To(HaveLen(1))
		Expect(created.OwnerReferences[0].Name).To(Equal("test-owner"))
	})

	It("returns error if creation fails with non-IsAlreadyExists", func() {
		brokenClient := &failingClient{}

		_, err := CreateGeneratedSecret(ctx, brokenClient, scheme, secretData, owner)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("forced failure"))
	})
})

type failingClient struct {
	client.Client
}

func (f *failingClient) Create(_ context.Context, _ client.Object, _ ...client.CreateOption) error {
	return errors.New("forced failure")
}
