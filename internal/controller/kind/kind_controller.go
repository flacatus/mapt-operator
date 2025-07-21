package kind

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

type KindReconciler struct {
	client.Client
	Scheme      *runtime.Scheme
	Provisioner clusters.GenericMaptProvisioner
}

func (r *KindReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx).WithValues("controller", "KindReconciler", "resource", req.NamespacedName)

	var kind v1alpha1.Kind
	if err := r.Get(ctx, req.NamespacedName, &kind); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Kind resource not found. It may have been deleted.")
			return ctrl.Result{}, nil
		}
		return ctrl.Result{}, controllerutils.LogError(logger, err, "Failed to fetch Kind resource")
	}

	kindCopy := kind.DeepCopy()

	prov := r.Provisioner
	if prov == nil {
		var err error
		prov, err = clusters.NewGenericMaptProvisioner(ctx, r.Client)
		if err != nil {
			return ctrl.Result{}, controllerutils.LogError(logger, err, "Failed to initialize provisioner")
		}
	}

	adapter, err := newAdapter(ctx, r.Client, kindCopy, prov, logger)
	if err != nil {
		return ctrl.Result{}, controllerutils.LogError(logger, err, "Failed to create adapter")
	}

	result, err := controller.ReconcileHandler([]controller.Operation{
		adapter.EnsureFinalizersAreCalled,
		adapter.EnsureFinalizerIsAdded,
		adapter.EnsureKindClusterIsProvisioned,
	})
	if err != nil {
		return result, controllerutils.LogError(logger, err, "Reconciliation failed")
	}

	result.RequeueAfter = 15 * time.Minute
	return result, nil
}

func (r *KindReconciler) Register(mgr ctrl.Manager, log *logr.Logger, _ crcluster.Cluster) error {
	r.Client = mgr.GetClient()
	r.Scheme = mgr.GetScheme()

	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Kind{}).
		Owns(&corev1.Secret{}).
		Named("kind").
		Complete(r)
}
