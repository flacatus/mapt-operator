package openshiftsnc

import (
	"github.com/mapt-oss/mapt-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type statusBuilder struct {
	status *v1alpha1.OpenshiftStatus
}

func newStatusBuilder(obj *v1alpha1.Openshift) *statusBuilder {
	if obj.Status.Conditions == nil {
		obj.Status.Conditions = []metav1.Condition{}
	}
	copyStatus := obj.Status.DeepCopy()
	return &statusBuilder{status: copyStatus}
}

func (s *statusBuilder) phase(p v1alpha1.OpenshiftSncPhase) *statusBuilder {
	s.status.Phase = p
	return s
}

func (s *statusBuilder) message(m string) *statusBuilder {
	s.status.Message = m
	return s
}

func (s *statusBuilder) backendID(id string) *statusBuilder {
	if id != "" {
		s.status.ProvisionId = &id
	}
	return s
}

func (s *statusBuilder) kubeconfigSecret(name string) *statusBuilder {
	if name != "" {
		s.status.KubeconfigSecretName = &name
	}
	return s
}

func (s *statusBuilder) condition(condType string, status metav1.ConditionStatus, reason, msg string) *statusBuilder {
	for _, c := range s.status.Conditions {
		if c.Type == condType && c.Message == msg {
			// Same condition type and message already exists, skip appending
			return s
		}
	}

	cond := metav1.Condition{
		Type:               condType,
		Status:             status,
		Reason:             reason,
		Message:            msg,
		LastTransitionTime: metav1.Now(),
	}

	s.status.Conditions = append(s.status.Conditions, cond)

	return s
}

func (a *adapter) updateStatus(update func(*v1alpha1.OpenshiftStatus)) error {
	// Create a deep copy of the current object to preserve the original for patching
	original := a.openshift.DeepCopy()

	// Apply updates to the current in-memory object
	update(&a.openshift.Status)

	// Patch ONLY the status subresource
	return a.client.Status().Patch(a.ctx, a.openshift, client.MergeFrom(original))
}
