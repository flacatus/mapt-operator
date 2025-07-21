package openshiftsnc

import (
	"context"
	"time"

	"github.com/go-logr/logr"
	"github.com/konflux-ci/operator-toolkit/controller"
	"github.com/mapt-oss/mapt-operator/api/v1alpha1"
	"github.com/mapt-oss/mapt-operator/pkg/clusters"
	"github.com/mapt-oss/mapt-operator/pkg/controllerutils"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	crcluster "sigs.k8s.io/controller-runtime/pkg/cluster"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

type OpenshiftReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Provisioner clusters.GenericMaptProvisioner
}

func (r *OpenshiftReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("controller", "OpenshiftReconciler", "resource", req.NamespacedName)

	var openshift v1alpha1.Openshift
	if err := r.Get(ctx, req.NamespacedName, &openshift); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Openshift resource not found. It may have been deleted.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, controllerutils.LogError(logger, err, "Failed to fetch Openshift resource")
	}

	prov := r.Provisioner
	if prov == nil {
		var err error
		prov, err = clusters.NewGenericMaptProvisioner(ctx, r.Client)
		if err != nil {
			return ctrl.Result{}, controllerutils.LogError(logger, err, "Failed to initialize provisioner")
		}
	}

	adapter := newAdapter(ctx, r.Client, prov, &openshift, logger)

	result, err := controller.ReconcileHandler([]controller.Operation{
		adapter.EnsureFinalizerIsAdded,
		adapter.EnsureFinalizersAreCalled,
		adapter.EnsureOpenshiftClusterIsProvisioned,
	})

	if err != nil {
		return result, controllerutils.LogError(logger, err, "Reconciliation failed")
	}

	logger.Info("Reconciliation successful. Requeueing after 10 hours")
	result.RequeueAfter = 10 * time.Hour
	return result, nil
}

func (r *OpenshiftReconciler) Register(mgr ctrl.Manager, log *logr.Logger, _ crcluster.Cluster) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Openshift{}).
		Owns(&corev1.Secret{}).
		Named("openshift").
		Complete(r)
}
