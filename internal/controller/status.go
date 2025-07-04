package controller

import (
	"github.com/flacatus/mapt-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type statusBuilder struct {
	status *v1alpha1.KindStatus
}

func newStatusBuilder(kind *v1alpha1.Kind) *statusBuilder {
	return &statusBuilder{status: &kind.Status}
}

func (s *statusBuilder) phase(phase v1alpha1.KindPhase) *statusBuilder {
	s.status.Phase = phase
	return s
}

func (s *statusBuilder) message(msg string) *statusBuilder {
	s.status.Message = msg
	return s
}

func (s *statusBuilder) condition(condType string, status metav1.ConditionStatus, reason, msg string) *statusBuilder {
	meta.SetStatusCondition(&s.status.Conditions, metav1.Condition{
		Type:    condType,
		Status:  status,
		Reason:  reason,
		Message: msg,
	})
	return s
}

func (s *statusBuilder) backendID(id string) *statusBuilder {
	if id != "" {
		s.status.ProvisionId = &id
	}
	return s
}

func (a *adapter) patchStatus(mutator func(*v1alpha1.KindStatus)) error {
	patch := client.MergeFrom(a.kind.DeepCopy())
	if a.kind.Status.Conditions == nil {
		a.kind.Status.Conditions = []metav1.Condition{}
	}
	mutator(&a.kind.Status)
	return a.client.Status().Patch(a.ctx, a.kind, patch)
}
