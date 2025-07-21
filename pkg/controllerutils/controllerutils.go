// Package controllerutils provides reusable controller helper functions for reconcilers.
package controllerutils

import (
	"context"
	"fmt"
	"time"

	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// FinalizerManager manages finalizer logic for a resource.
type FinalizerManager struct {
	Client    client.Client
	Ctx       context.Context
	Finalizer string
	Logger    logr.Logger
}

// EnsureFinalizer ensures the finalizer is present on the object.
func (f *FinalizerManager) EnsureFinalizer(obj client.Object) (reconcile.Result, error) {
	if controllerutil.ContainsFinalizer(obj, f.Finalizer) {
		return reconcile.Result{}, nil
	}

	orig := obj.DeepCopyObject().(client.Object)
	controllerutil.AddFinalizer(obj, f.Finalizer)
	if err := f.Client.Patch(f.Ctx, obj, client.MergeFrom(orig)); err != nil {
		f.Logger.Error(err, "failed to add finalizer")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// RemoveFinalizer removes the finalizer if present.
func (f *FinalizerManager) RemoveFinalizer(obj client.Object) (reconcile.Result, error) {
	if !controllerutil.ContainsFinalizer(obj, f.Finalizer) {
		return reconcile.Result{}, nil
	}
	orig := obj.DeepCopyObject().(client.Object)
	controllerutil.RemoveFinalizer(obj, f.Finalizer)
	if err := f.Client.Patch(f.Ctx, obj, client.MergeFrom(orig)); err != nil {
		f.Logger.Error(err, "failed to remove finalizer")
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// RequeueAfter returns a reconcile.Result with a requeue duration.
func RequeueAfter(d time.Duration) reconcile.Result {
	return reconcile.Result{RequeueAfter: d}
}

// LogError logs an error with context.
func LogError(log logr.Logger, err error, msg string) error {
	if err != nil {
		log.Error(err, msg)
	}
	return err
}

// SetOrUpdateCondition sets or updates a condition in-place.
func SetOrUpdateCondition(conds *[]metav1.Condition, cond metav1.Condition) {
	existing := findCondition(*conds, cond.Type)
	if existing == nil {
		*conds = append(*conds, cond)
	} else {
		changed := existing.Status != cond.Status || existing.Reason != cond.Reason || existing.Message != cond.Message
		if changed {
			existing.Status = cond.Status
			existing.Reason = cond.Reason
			existing.Message = cond.Message
			existing.LastTransitionTime = cond.LastTransitionTime
		}
	}
}

// findCondition looks for a condition by type.
func findCondition(conds []metav1.Condition, t string) *metav1.Condition {
	for i := range conds {
		if conds[i].Type == t {
			return &conds[i]
		}
	}
	return nil
}

// FormatPrice formats a float price to string.
func FormatPrice(price float64) string {
	return fmt.Sprintf("%.4f USD/hour", price)
}
