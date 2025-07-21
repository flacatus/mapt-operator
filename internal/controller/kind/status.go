package kind

import (
	"github.com/mapt-oss/mapt-operator/api/v1alpha1"
	"github.com/mapt-oss/mapt-operator/pkg/controllerutils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type statusBuilder struct {
	status *v1alpha1.KindStatus
}

func newStatusBuilder(kind *v1alpha1.Kind) *statusBuilder {
	if kind.Status.Conditions == nil {
		kind.Status.Conditions = []metav1.Condition{}
	}
	// DeepCopy to avoid mutating original
	copyStatus := kind.Status.DeepCopy()
	return &statusBuilder{status: copyStatus}
}

func (s *statusBuilder) phase(phase v1alpha1.KindPhase) *statusBuilder {
	s.status.Phase = phase
	return s
}

func (s *statusBuilder) message(msg string) *statusBuilder {
	s.status.Message = msg
	return s
}

func (s *statusBuilder) avgPrice(avgPrice float64) *statusBuilder {
	s.status.AveragePrice = controllerutils.FormatPrice(avgPrice)
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

func (s *statusBuilder) backendID(id string) *statusBuilder {
	if id != "" {
		s.status.ProvisionId = &id
	}
	return s
}

func (a *adapter) updateStatus(update func(*v1alpha1.KindStatus)) error {
	// Create a deep copy of the current object to preserve the original for patching
	original := a.kind.DeepCopy()

	// Apply updates to the current in-memory object
	update(&a.kind.Status)

	// Patch ONLY the status subresource
	return a.client.Status().Patch(a.ctx, a.kind, client.MergeFrom(original))
}
