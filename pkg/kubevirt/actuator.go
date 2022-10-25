/*
Copyright 2019 The OpenShift Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package kubevirt

import (
	"context"
	"fmt"
	"reflect"

	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	configv1 "github.com/uccps-samples/api/config/v1"
	operatorv1 "github.com/uccps-samples/api/operator/v1"
	minterv1 "github.com/uccps-samples/cloud-credential-operator/pkg/apis/cloudcredential/v1"
	"github.com/uccps-samples/cloud-credential-operator/pkg/operator/constants"
	actuatoriface "github.com/uccps-samples/cloud-credential-operator/pkg/operator/credentialsrequest/actuator"
	"github.com/uccps-samples/cloud-credential-operator/pkg/operator/utils"
)

type KubevirtActuator struct {
	Client client.Client
}

const (
	KubevirtCredentialsSecretKey = "kubeconfig"
)

// NewActuator creates a new Kubevirt actuator.
func NewActuator(client client.Client) (*KubevirtActuator, error) {
	return &KubevirtActuator{
		Client: client,
	}, nil
}

// Exists checks if the credentials currently exist.
// TODO: in the future validate the expiration of the credentials
func (a *KubevirtActuator) Exists(ctx context.Context, cr *minterv1.CredentialsRequest) (bool, error) {
	logger := a.getLogger(cr)
	logger.Debug("running Exists")

	existingSecret, err := a.getSecret(ctx, cr, logger)
	if err != nil {
		return false, err
	}

	return existingSecret != nil, nil
}

// Create the credentials.
func (a *KubevirtActuator) Create(ctx context.Context, cr *minterv1.CredentialsRequest) error {
	logger := a.getLogger(cr)
	logger.Debug("running Create")
	return a.sync(ctx, cr, logger)
}

// Update the credentials to the provided definition.
func (a *KubevirtActuator) Update(ctx context.Context, cr *minterv1.CredentialsRequest) error {
	logger := a.getLogger(cr)
	logger.Debug("running Update")
	return a.sync(ctx, cr, logger)
}

// Delete credentials
func (a *KubevirtActuator) Delete(ctx context.Context, cr *minterv1.CredentialsRequest) error {
	logger := a.getLogger(cr)
	logger.Debug("running Delete")

	existingSecret, err := a.getSecret(ctx, cr, logger)
	if err != nil {
		return err
	}
	if existingSecret != nil {
		logger.Info("Deleting existing secret")
		if err = a.Client.Delete(ctx, existingSecret); err != nil {
			return err
		}
	}

	return nil
}

// GetCredentialsRootSecretLocation returns the namespace and name where the parent credentials secret is stored.
func (a *KubevirtActuator) GetCredentialsRootSecretLocation() types.NamespacedName {
	return types.NamespacedName{Namespace: constants.CloudCredSecretNamespace, Name: constants.KubevirtCloudCredSecretName}
}

func (a *KubevirtActuator) GetCredentialsRootSecret(ctx context.Context, cr *minterv1.CredentialsRequest) (*corev1.Secret, error) {
	logger := a.getLogger(cr)

	// get the secret of the kubevirt credentials
	kubevirtCredentialsSecret := &corev1.Secret{}
	if err := a.Client.Get(ctx, a.GetCredentialsRootSecretLocation(), kubevirtCredentialsSecret); err != nil {
		msg := "unable to fetch root cloud cred secret"
		logger.WithError(err).Error(msg)
		return nil, &actuatoriface.ActuatorError{
			ErrReason: minterv1.CredentialsProvisionFailure,
			Message:   fmt.Sprintf("%v: %v", msg, err),
		}
	}

	return kubevirtCredentialsSecret, nil
}

func (a *KubevirtActuator) sync(ctx context.Context, cr *minterv1.CredentialsRequest, logger log.FieldLogger) error {
	logger.Debug("running sync")

	credentialsRootSecret, err := a.GetCredentialsRootSecret(context.TODO(), cr)
	if err != nil {
		logger.WithError(err).Error("issue with cloud credentials secret")
		return err
	}

	// get the secret data from the credentials request
	kubevirtCredentialData, err := a.getCredentialsSecretData(credentialsRootSecret, logger)
	if err != nil {
		logger.WithError(err).Error("issue with cloud credentials secret")
		return err
	}

	// get the existing secret in order to check if need to update or create a new
	logger.Debug("provisioning secret")
	existingSecret, err := a.getSecret(ctx, cr, logger)
	if err != nil {
		return err
	}

	// check if need to update or create a new one
	if err = a.syncCredentialSecret(ctx, cr, &kubevirtCredentialData, existingSecret, logger); err != nil {
		msg := "error creating/updating secret"
		logger.WithError(err).Error(msg)
		return &actuatoriface.ActuatorError{
			ErrReason: minterv1.CredentialsProvisionFailure,
			Message:   fmt.Sprintf("%v: %v", msg, err),
		}
	}
	return nil
}

func (a *KubevirtActuator) getCredentialsSecretData(cloudCredSecret *corev1.Secret, logger log.FieldLogger) ([]byte, error) {
	// get the secret data - the infra kubeconfig
	infraClusterKubeconfig, ok := cloudCredSecret.Data[KubevirtCredentialsSecretKey]
	if !ok {
		return nil, fmt.Errorf("kubeVirt credentials secret %s did not contain key %s", cloudCredSecret.Name, KubevirtCredentialsSecretKey)
	}

	logger.Debug("extracted kubevirt credentials")
	return infraClusterKubeconfig, nil
}

func (a *KubevirtActuator) syncCredentialSecret(ctx context.Context, cr *minterv1.CredentialsRequest, kubevirtCredentialData *[]byte, existingSecret *corev1.Secret, logger log.FieldLogger) error {
	if existingSecret == nil {
		if kubevirtCredentialData == nil {
			msg := "new access key secret needed but no key data provided"
			logger.Error(msg)
			return &actuatoriface.ActuatorError{
				ErrReason: minterv1.CredentialsProvisionFailure,
				Message:   msg,
			}
		}

		return a.createNewSecret(logger, cr, kubevirtCredentialData, ctx)
	}

	return a.updateExistingSecret(logger, existingSecret, cr, kubevirtCredentialData)
}

func (a *KubevirtActuator) updateExistingSecret(logger log.FieldLogger, existingSecret *corev1.Secret, cr *minterv1.CredentialsRequest, kubevirtCredentialData *[]byte) error {
	// Update the existing secret:
	logger.Debug("updating secret")
	origSecret := existingSecret.DeepCopy()
	if existingSecret.Annotations == nil {
		existingSecret.Annotations = map[string]string{}
	}
	existingSecret.Annotations[minterv1.AnnotationCredentialsRequest] = fmt.Sprintf("%s/%s", cr.Namespace, cr.Name)
	if kubevirtCredentialData != nil {
		existingSecret.Data = map[string][]byte{
			KubevirtCredentialsSecretKey: *kubevirtCredentialData,
		}
	}

	if !reflect.DeepEqual(existingSecret, origSecret) {
		logger.Info("target secret has changed, updating")
		if err := a.Client.Update(context.TODO(), existingSecret); err != nil {
			msg := "error updating secret"
			logger.WithError(err).Error(msg)
			return &actuatoriface.ActuatorError{
				ErrReason: minterv1.CredentialsProvisionFailure,
				Message:   msg,
			}
		}
	} else {
		logger.Debug("target secret unchanged")
	}

	return nil
}

func (a *KubevirtActuator) createNewSecret(logger log.FieldLogger, cr *minterv1.CredentialsRequest, kubevirtCredentialData *[]byte, ctx context.Context) error {
	logger.Info("creating secret")
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Spec.SecretRef.Name,
			Namespace: cr.Spec.SecretRef.Namespace,
			Annotations: map[string]string{
				minterv1.AnnotationCredentialsRequest: fmt.Sprintf("%s/%s", cr.Namespace, cr.Name),
			},
		},
		Data: map[string][]byte{
			KubevirtCredentialsSecretKey: *kubevirtCredentialData,
		},
	}

	if err := a.Client.Create(ctx, secret); err != nil {
		msg := "error in creating a secret"
		logger.WithError(err).Error(msg)
		return &actuatoriface.ActuatorError{
			ErrReason: minterv1.CredentialsProvisionFailure,
			Message:   msg,
		}
	}

	logger.Info("secret created successfully")
	return nil
}

func (a *KubevirtActuator) getSecret(ctx context.Context, cr *minterv1.CredentialsRequest, logger log.FieldLogger) (*corev1.Secret, error) {
	logger.Debug("running getSecret")

	existingSecret := &corev1.Secret{}
	if err := a.Client.Get(ctx, types.NamespacedName{Namespace: cr.Spec.SecretRef.Namespace, Name: cr.Spec.SecretRef.Name}, existingSecret); err != nil {
		if errors.IsNotFound(err) {
			logger.Debug("target secret does not exist")
			return nil, nil
		}
		return nil, err
	}

	if _, ok := existingSecret.Data[KubevirtCredentialsSecretKey]; !ok {
		logger.Warningf("secret did not have expected key: %s", KubevirtCredentialsSecretKey)
	}

	logger.Debug("target secret exists")
	return existingSecret, nil
}

func (a *KubevirtActuator) getLogger(cr *minterv1.CredentialsRequest) log.FieldLogger {
	return log.WithFields(log.Fields{
		"actuator":     "Kubevirt",
		"targetSecret": fmt.Sprintf("%s/%s", cr.Spec.SecretRef.Namespace, cr.Spec.SecretRef.Name),
		"cr":           fmt.Sprintf("%s/%s", cr.Namespace, cr.Name),
	})
}

func (a *KubevirtActuator) Upgradeable(mode operatorv1.CloudCredentialsMode) *configv1.ClusterOperatorStatusCondition {
	return utils.UpgradeableCheck(a.Client, mode, a.GetCredentialsRootSecretLocation())
}
