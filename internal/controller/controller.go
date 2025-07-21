package controller

import (
	"github.com/konflux-ci/operator-toolkit/controller"
	"github.com/mapt-oss/mapt-operator/internal/controller/kind"
	openshiftsnc "github.com/mapt-oss/mapt-operator/internal/controller/openshift-snc"
)

// EnabledControllers is a slice containing references to all the controllers that have to be registered
var EnabledControllers = []controller.Controller{
	&kind.KindReconciler{},
	&openshiftsnc.OpenshiftReconciler{},
}
