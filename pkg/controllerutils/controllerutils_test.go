package controllerutils_test

import (
	"context"
	"errors"
	"testing"
	"time"

	. "github.com/mapt-oss/mapt-operator/pkg/controllerutils"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var _ = Describe("FinalizerManager", func() {
	var (
		ctx        context.Context
		fakeClient client.Client
		fm         *FinalizerManager
		testObj    *corev1.ConfigMap
		testScheme *runtime.Scheme
		finalizer  string
	)

	BeforeEach(func() {
		ctx = context.Background()
		finalizer = "test.finalizer"
		testScheme = runtime.NewScheme()
		Expect(corev1.AddToScheme(testScheme)).To(Succeed())

		testObj = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-cm",
				Namespace: "default",
			},
		}

		fakeClient = fake.NewClientBuilder().
			WithScheme(testScheme).
			WithObjects(testObj).
			Build()

		fm = &FinalizerManager{
			Client:    fakeClient,
			Ctx:       ctx,
			Finalizer: finalizer,
			Logger:    logr.Discard(),
		}
	})

	Describe("EnsureFinalizer", func() {
		It("adds the finalizer if not present", func() {
			_, err := fm.EnsureFinalizer(testObj)
			Expect(err).NotTo(HaveOccurred())
			Expect(controllerutil.ContainsFinalizer(testObj, finalizer)).To(BeTrue())
		})

		It("does nothing if finalizer already present", func() {
			controllerutil.AddFinalizer(testObj, finalizer)
			_, err := fm.EnsureFinalizer(testObj)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("RemoveFinalizer", func() {
		It("removes the finalizer if present", func() {
			controllerutil.AddFinalizer(testObj, finalizer)
			_, err := fm.RemoveFinalizer(testObj)
			Expect(err).NotTo(HaveOccurred())
			Expect(controllerutil.ContainsFinalizer(testObj, finalizer)).To(BeFalse())
		})

		It("does nothing if finalizer not present", func() {
			_, err := fm.RemoveFinalizer(testObj)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("RequeueAfter", func() {
	It("returns the correct duration", func() {
		d := 5 * time.Second
		res := RequeueAfter(d)
		Expect(res.RequeueAfter).To(Equal(d))
	})
})

var _ = Describe("LogError", func() {
	It("logs and returns the error", func() {
		err := errors.New("some error")
		result := LogError(logr.Discard(), err, "error message")
		Expect(result).To(Equal(err))
	})

	It("returns nil if no error", func() {
		result := LogError(logr.Discard(), nil, "should not log")
		Expect(result).To(BeNil())
	})
})

var _ = Describe("SetOrUpdateCondition", func() {
	var conds []metav1.Condition
	condTime := metav1.Now()

	It("adds a new condition if not present", func() {
		cond := metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			Reason:             "Success",
			Message:            "All good",
			LastTransitionTime: condTime,
		}
		SetOrUpdateCondition(&conds, cond)
		Expect(len(conds)).To(Equal(1))
		Expect(conds[0].Type).To(Equal("Ready"))
	})

	It("updates existing condition", func() {
		conds = []metav1.Condition{{
			Type:    "Ready",
			Status:  metav1.ConditionFalse,
			Reason:  "Init",
			Message: "Waiting",
		}}

		updated := metav1.Condition{
			Type:               "Ready",
			Status:             metav1.ConditionTrue,
			Reason:             "Success",
			Message:            "Now good",
			LastTransitionTime: condTime,
		}

		SetOrUpdateCondition(&conds, updated)
		Expect(conds[0].Status).To(Equal(metav1.ConditionTrue))
		Expect(conds[0].Message).To(Equal("Now good"))
		Expect(conds[0].LastTransitionTime).To(Equal(condTime))
	})
})

var _ = Describe("FormatPrice", func() {
	It("formats float to price string", func() {
		result := FormatPrice(0.123456)
		Expect(result).To(Equal("0.1235 USD/hour"))
	})
})

func TestControllerUtils(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ControllerUtils Suite")
}
