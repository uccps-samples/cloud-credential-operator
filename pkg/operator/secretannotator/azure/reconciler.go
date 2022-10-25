package azure

import (
	"context"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	operatorv1 "github.com/uccps-samples/api/operator/v1"

	"github.com/uccps-samples/cloud-credential-operator/pkg/operator/constants"
	"github.com/uccps-samples/cloud-credential-operator/pkg/operator/metrics"
	"github.com/uccps-samples/cloud-credential-operator/pkg/operator/secretannotator/status"
	secretutils "github.com/uccps-samples/cloud-credential-operator/pkg/operator/secretannotator/utils"
	statuscontroller "github.com/uccps-samples/cloud-credential-operator/pkg/operator/status"
	"github.com/uccps-samples/cloud-credential-operator/pkg/operator/utils"

	log "github.com/sirupsen/logrus"
)

var _ reconcile.Reconciler = &ReconcileCloudCredSecret{}

type ReconcileCloudCredSecret struct {
	client.Client
	Logger log.FieldLogger
}

func NewReconciler(mgr manager.Manager) reconcile.Reconciler {
	c := mgr.GetClient()
	r := &ReconcileCloudCredSecret{
		Client: c,
		Logger: log.WithField("controller", constants.SecretAnnotatorControllerName),
	}

	s := status.NewSecretStatusHandler(c)
	statuscontroller.AddHandler(constants.SecretAnnotatorControllerName, s)

	return r
}

func Add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New(constants.SecretAnnotatorControllerName, mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to cluster cloud secret
	p := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			return cloudCredSecretObjectCheck(e.ObjectNew)
		},
		CreateFunc: func(e event.CreateEvent) bool {
			return cloudCredSecretObjectCheck(e.Object)
		},
		DeleteFunc: func(e event.DeleteEvent) bool {
			return cloudCredSecretObjectCheck(e.Object)
		},
	}
	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForObject{}, p)
	if err != nil {
		return err
	}

	secretutils.WatchCCOConfig(c, types.NamespacedName{
		Namespace: constants.CloudCredSecretNamespace,
		Name:      constants.AzureCloudCredSecretName,
	})

	return nil
}

func cloudCredSecretObjectCheck(secret metav1.Object) bool {
	return secret.GetNamespace() == constants.CloudCredSecretNamespace && secret.GetName() == constants.AzureCloudCredSecretName
}

func (r *ReconcileCloudCredSecret) Reconcile(ctx context.Context, request reconcile.Request) (returnResult reconcile.Result, returnErr error) {
	start := time.Now()

	defer func() {
		dur := time.Since(start)
		metrics.MetricControllerReconcileTime.WithLabelValues(constants.SecretAnnotatorControllerName).Observe(dur.Seconds())
	}()

	mode, conflict, err := utils.GetOperatorConfiguration(r.Client, r.Logger)
	if err != nil {
		r.Logger.WithError(err).Error("error checking operator configuration")
		return reconcile.Result{}, err
	}
	if !utils.IsValidMode(mode) {
		r.Logger.Errorf("invalid mode of %s set", mode)
		return reconcile.Result{}, fmt.Errorf("invalide mode of %s set", mode)
	}
	if conflict {
		r.Logger.Errorf("configuratoin conflict between legacy configmap and operator config")
		return reconcile.Result{}, fmt.Errorf("configuratoin conflict")
	}
	if mode == operatorv1.CloudCredentialsModeManual {
		r.Logger.Info("operator in disabled / manual mode")
		return reconcile.Result{}, err
	}

	// From here down we must be in Passthrough mode

	secret := &corev1.Secret{}
	err = r.Get(context.Background(), request.NamespacedName, secret)
	if err != nil {
		r.Logger.Debugf("secret not found: %v", err)
		return reconcile.Result{}, err
	}

	var effectiveMode operatorv1.CloudCredentialsMode

	if mode == operatorv1.CloudCredentialsModeDefault {
		effectiveMode = operatorv1.CloudCredentialsModePassthrough
	} else {
		effectiveMode = mode
	}

	annotation, err := utils.ModeToAnnotation(effectiveMode)
	if err != nil {
		r.Logger.WithError(err).Error("failed to convert operator mode to annotation")
		return reconcile.Result{}, err
	}
	err = r.updateSecretAnnotations(secret, annotation)
	if err != nil {
		r.Logger.WithError(err).Error("errored while annotating secret")
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileCloudCredSecret) updateSecretAnnotations(secret *corev1.Secret, value string) error {
	secretAnnotations := secret.GetAnnotations()
	if secretAnnotations == nil {
		secretAnnotations = map[string]string{}
	}

	secretAnnotations[constants.AnnotationKey] = value
	secret.SetAnnotations(secretAnnotations)

	return r.Update(context.Background(), secret)
}
