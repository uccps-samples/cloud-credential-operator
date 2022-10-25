package api

import (
	kadmissionv1beta1 "k8s.io/api/admission/v1beta1"
	kadmissionregistrationv1beta1 "k8s.io/api/admissionregistration/v1beta1"
	kappsv1 "k8s.io/api/apps/v1"
	kappsv1beta1 "k8s.io/api/apps/v1beta1"
	kappsv1beta2 "k8s.io/api/apps/v1beta2"
	kauthenticationv1 "k8s.io/api/authentication/v1"
	kauthenticationv1beta1 "k8s.io/api/authentication/v1beta1"
	kauthorizationv1 "k8s.io/api/authorization/v1"
	kauthorizationv1beta1 "k8s.io/api/authorization/v1beta1"
	kautoscalingv1 "k8s.io/api/autoscaling/v1"
	kautoscalingv2beta1 "k8s.io/api/autoscaling/v2beta1"
	kbatchv1 "k8s.io/api/batch/v1"
	kbatchv1beta1 "k8s.io/api/batch/v1beta1"
	kcertificatesv1beta1 "k8s.io/api/certificates/v1beta1"
	kcorev1 "k8s.io/api/core/v1"
	keventsv1beta1 "k8s.io/api/events/v1beta1"
	kextensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	kimagepolicyv1alpha1 "k8s.io/api/imagepolicy/v1alpha1"
	knetworkingv1 "k8s.io/api/networking/v1"
	kpolicyv1 "k8s.io/api/policy/v1"
	kpolicyv1beta1 "k8s.io/api/policy/v1beta1"
	krbacv1 "k8s.io/api/rbac/v1"
	krbacv1alpha1 "k8s.io/api/rbac/v1alpha1"
	krbacv1beta1 "k8s.io/api/rbac/v1beta1"
	kschedulingv1alpha1 "k8s.io/api/scheduling/v1alpha1"
	kschedulingv1beta1 "k8s.io/api/scheduling/v1beta1"
	kstoragev1 "k8s.io/api/storage/v1"
	kstoragev1alpha1 "k8s.io/api/storage/v1alpha1"
	kstoragev1beta1 "k8s.io/api/storage/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/uccps-samples/api/apiserver"
	"github.com/uccps-samples/api/apps"
	"github.com/uccps-samples/api/authorization"
	"github.com/uccps-samples/api/build"
	"github.com/uccps-samples/api/cloudnetwork"
	"github.com/uccps-samples/api/config"
	"github.com/uccps-samples/api/helm"
	"github.com/uccps-samples/api/image"
	"github.com/uccps-samples/api/imageregistry"
	"github.com/uccps-samples/api/kubecontrolplane"
	"github.com/uccps-samples/api/machine"
	"github.com/uccps-samples/api/network"
	"github.com/uccps-samples/api/networkoperator"
	"github.com/uccps-samples/api/oauth"
	"github.com/uccps-samples/api/uccpcontrolplane"
	"github.com/uccps-samples/api/operator"
	"github.com/uccps-samples/api/operatorcontrolplane"
	"github.com/uccps-samples/api/osin"
	"github.com/uccps-samples/api/project"
	"github.com/uccps-samples/api/quota"
	"github.com/uccps-samples/api/route"
	"github.com/uccps-samples/api/samples"
	"github.com/uccps-samples/api/security"
	"github.com/uccps-samples/api/servicecertsigner"
	"github.com/uccps-samples/api/sharedresource"
	"github.com/uccps-samples/api/template"
	"github.com/uccps-samples/api/user"

	// just make sure this compiles.  Don't add it to a scheme
	_ "github.com/uccps-samples/api/legacyconfig/v1"
)

var (
	schemeBuilder = runtime.NewSchemeBuilder(
		apiserver.Install,
		apps.Install,
		authorization.Install,
		build.Install,
		config.Install,
		helm.Install,
		image.Install,
		imageregistry.Install,
		kubecontrolplane.Install,
		cloudnetwork.Install,
		network.Install,
		networkoperator.Install,
		oauth.Install,
		uccpcontrolplane.Install,
		operator.Install,
		operatorcontrolplane.Install,
		osin.Install,
		project.Install,
		quota.Install,
		route.Install,
		samples.Install,
		security.Install,
		servicecertsigner.Install,
		sharedresource.Install,
		template.Install,
		user.Install,
		machine.Install,
	)
	// Install is a function which adds every version of every uccp group to a scheme
	Install = schemeBuilder.AddToScheme

	kubeSchemeBuilder = runtime.NewSchemeBuilder(
		kadmissionv1beta1.AddToScheme,
		kadmissionregistrationv1beta1.AddToScheme,
		kappsv1.AddToScheme,
		kappsv1beta1.AddToScheme,
		kappsv1beta2.AddToScheme,
		kauthenticationv1.AddToScheme,
		kauthenticationv1beta1.AddToScheme,
		kauthorizationv1.AddToScheme,
		kauthorizationv1beta1.AddToScheme,
		kautoscalingv1.AddToScheme,
		kautoscalingv2beta1.AddToScheme,
		kbatchv1.AddToScheme,
		kbatchv1beta1.AddToScheme,
		kcertificatesv1beta1.AddToScheme,
		kcorev1.AddToScheme,
		keventsv1beta1.AddToScheme,
		kextensionsv1beta1.AddToScheme,
		kimagepolicyv1alpha1.AddToScheme,
		knetworkingv1.AddToScheme,
		kpolicyv1.AddToScheme,
		kpolicyv1beta1.AddToScheme,
		krbacv1.AddToScheme,
		krbacv1beta1.AddToScheme,
		krbacv1alpha1.AddToScheme,
		kschedulingv1alpha1.AddToScheme,
		kschedulingv1beta1.AddToScheme,
		kstoragev1.AddToScheme,
		kstoragev1beta1.AddToScheme,
		kstoragev1alpha1.AddToScheme,
	)
	// InstallKube is a way to install all the external k8s.io/api types
	InstallKube = kubeSchemeBuilder.AddToScheme
)
