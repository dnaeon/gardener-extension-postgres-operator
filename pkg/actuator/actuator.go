// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

// Package actuator provides the implementation of a Gardener extension
// actuator.
package actuator

import (
	"context"
	"errors"
	"fmt"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	"github.com/gardener/gardener/extensions/pkg/controller/extension"
	v1beta1helper "github.com/gardener/gardener/pkg/apis/core/v1beta1/helper"
	extensionsv1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	"github.com/go-logr/logr"
	acidv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/util/retry"
	"k8s.io/component-base/featuregate"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/gardener/gardener-extension-postgres-operator/pkg/apis/config"
	"github.com/gardener/gardener-extension-postgres-operator/pkg/apis/config/validation"
	"github.com/gardener/gardener-extension-postgres-operator/pkg/metrics"
)

const (
	// Name is the name of the actuator
	Name = "postgres"
	// ExtensionType is the type of the extension resources, which the
	// actuator reconciles.
	ExtensionType = "postgres"
	// FinalizerSuffix is the finalizer suffix used by the actuator
	FinalizerSuffix = "gardener-extension-postgres-operator"

	// postgresClusterName is the name of the Postgres cluster created by
	// the extension.
	postgresClusterName = "postgres-cluster"
)

// Actuator is an implementation of [extension.Actuator].
type Actuator struct {
	reader  client.Reader
	client  client.Client
	decoder runtime.Decoder
	scheme  *runtime.Scheme

	// The following fields are usually derived from the list of extra Helm
	// values provided by gardenlet during the deployment of the extension.
	//
	// See the link below for more details about how gardenlet provides
	// extra values to Helm during the extension deployment.
	//
	// https://github.com/gardener/gardener/blob/d5071c800378616eb6bb2c7662b4b28f4cfe7406/pkg/gardenlet/controller/controllerinstallation/controllerinstallation/reconciler.go#L236-L263
	gardenerVersion       string
	gardenletFeatureGates map[featuregate.Feature]bool
}

var _ extension.Actuator = &Actuator{}

// Option is a function, which configures the [Actuator].
type Option func(a *Actuator) error

// New creates a new actuator with the given options.
func New(opts ...Option) (*Actuator, error) {
	act := &Actuator{
		gardenletFeatureGates: make(map[featuregate.Feature]bool),
		scheme:                scheme.Scheme,
	}

	for _, opt := range opts {
		if err := opt(act); err != nil {
			return nil, err
		}
	}

	return act, nil
}

// WithClient is an [Option], which configures the [Actuator] with the given
// [client.Client].
func WithClient(c client.Client) Option {
	opt := func(a *Actuator) error {
		a.client = c

		return nil
	}

	return opt
}

// WithReader is an [Option], which configures the [Actuator] with the given
// [client.Reader].
func WithReader(r client.Reader) Option {
	opt := func(a *Actuator) error {
		a.reader = r

		return nil
	}

	return opt
}

// WithDecoder is an [Option], which configures the [Actuator] with the given
// [runtime.Decoder].
func WithDecoder(d runtime.Decoder) Option {
	opt := func(a *Actuator) error {
		a.decoder = d

		return nil
	}

	return opt
}

// WithGardenerVersion is an [Option], which configures the [Actuator] with the
// given version of Gardener. This version of Gardener is usually provided by
// the gardenlet as part of the extra Helm values during deployment of the
// extension.
func WithGardenerVersion(v string) Option {
	opt := func(a *Actuator) error {
		a.gardenerVersion = v

		return nil
	}

	return opt
}

// WithGardenletFeatures is an [Option], which configures the [Actuator] with
// the given gardenlet feature gates. These feature gates are usually provided
// by the gardenlet as part of the extra Helm values during deployment of the
// extension.
func WithGardenletFeatures(feats map[featuregate.Feature]bool) Option {
	opt := func(a *Actuator) error {
		a.gardenletFeatureGates = feats

		return nil
	}

	return opt
}

// Name returns the name of the actuator. This name can be used when registering
// a controller for the actuator.
func (a *Actuator) Name() string {
	return Name
}

// FinalizerSuffix returns the finalizer suffix to use for the actuator. The
// result of this method may be used when registering a controller with the
// actuator.
func (a *Actuator) FinalizerSuffix() string {
	return FinalizerSuffix
}

// ExtensionType returns the type of extension resources the actuator
// reconciles. The result of this method may be used when registering a
// controller with the actuator.
func (a *Actuator) ExtensionType() string {
	return ExtensionType
}

// ExtensionClass returns the [extensionsv1alpha1.ExtensionClass] for the
// actuator. The result of this method may be used when registering a controller
// with the actuator.
func (a *Actuator) ExtensionClass() extensionsv1alpha1.ExtensionClass {
	return extensionsv1alpha1.ExtensionClassShoot
}

// Reconcile reconciles the [extensionsv1alpha1.Extension] resource by taking
// care of any resources managed by the [Actuator]. This method implements the
// [extension.Actuator] interface.
func (a *Actuator) Reconcile(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// The cluster name is the same as the name of the namespace for our
	// [extensionsv1alpha1.Extension] resource.
	clusterName := ex.Namespace

	// Increment our example metrics counter
	defer func() {
		metrics.ActuatorOperationTotal.WithLabelValues(clusterName, "reconcile").Inc()
	}()

	logger.Info("reconciling extension", "name", ex.Name, "cluster", clusterName)

	cluster, err := extensionscontroller.GetCluster(ctx, a.client, clusterName)
	if err != nil {
		return fmt.Errorf("failed to get cluster: %w", err)
	}

	// Nothing to do here, if the shoot cluster is hibernated at the moment.
	if v1beta1helper.HibernationIsEnabled(cluster.Shoot) {
		return nil
	}

	// Validate the extension configuration
	if ex.Spec.ProviderConfig == nil {
		return errors.New("no provider config specified")
	}

	// Decode provider spec configuration into our known config type.
	var cfg config.PostgresConfig
	if err := runtime.DecodeInto(a.decoder, ex.Spec.ProviderConfig.Raw, &cfg); err != nil {
		return errors.New("invalid provider spec configuration")
	}

	if err := validation.Validate(cfg); err != nil {
		return err
	}

	users := make(map[string]acidv1.UserFlags)
	for user, roles := range cfg.Spec.Users {
		users[user] = acidv1.UserFlags(roles)
	}

	postgres := &acidv1.Postgresql{
		ObjectMeta: v1.ObjectMeta{
			Namespace: ex.Namespace,
			Name:      postgresClusterName,
		},
		Spec: acidv1.PostgresSpec{
			Volume: acidv1.Volume{
				Size: cfg.Spec.VolumeSize.String(),
			},
			NumberOfInstances: cfg.Spec.Replicas,
			Users:             users,
			Databases:         cfg.Spec.Databases,
			PostgresqlParam: acidv1.PostgresqlParam{
				PgVersion: cfg.Spec.PostgresVersion,
			},
		},
	}

	// Create the Postgres cluster, unless it exists already.
	pc := acidv1.Postgresql{}
	if err := a.client.Get(ctx, client.ObjectKeyFromObject(postgres), &pc); apierrors.IsNotFound(err) {
		logger.Info("creating postgres cluster resource")
		return a.client.Create(ctx, postgres)
	}

	// Update the Postgres cluster spec
	return retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return a.client.Update(ctx, postgres)
	})
}

// Delete deletes any resources managed by the [Actuator]. This method
// implements the [extension.Actuator] interface.
func (a *Actuator) Delete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// Increment our example metrics counter
	defer func() {
		metrics.ActuatorOperationTotal.WithLabelValues(ex.Namespace, "delete").Inc()
	}()

	logger.Info("deleting postgres cluster resource")

	postgres := &acidv1.Postgresql{
		ObjectMeta: v1.ObjectMeta{
			Namespace: ex.Namespace,
			Name:      postgresClusterName,
		},
	}

	return client.IgnoreNotFound(a.client.Delete(ctx, postgres))
}

// ForceDelete signals the [Actuator] to delete any resources managed by it,
// because of a force-delete event of the shoot cluster. This method implements
// the [extension.Actuator] interface.
func (a *Actuator) ForceDelete(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// Increment our example metrics counter
	defer func() {
		metrics.ActuatorOperationTotal.WithLabelValues(ex.Namespace, "force_delete").Inc()
	}()

	logger.Info("shoot has been force-deleted, deleting resources managed by extension")

	return a.Delete(ctx, logger, ex)
}

// Restore restores the resources managed by the extension [Actuator]. This
// method implements the [extension.Actuator] interface.
func (a *Actuator) Restore(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// Increment our example metrics counter
	defer func() {
		metrics.ActuatorOperationTotal.WithLabelValues(ex.Namespace, "restore").Inc()
	}()

	return a.Reconcile(ctx, logger, ex)
}

// Migrate signals the [Actuator] to reconcile the resources managed by it,
// because of a shoot control-plane migration event. This method implements the
// [extension.Actuator] interface.
func (a *Actuator) Migrate(ctx context.Context, logger logr.Logger, ex *extensionsv1alpha1.Extension) error {
	// Increment our example metrics counter
	defer func() {
		metrics.ActuatorOperationTotal.WithLabelValues(ex.Namespace, "migrate").Inc()
	}()

	return a.Reconcile(ctx, logger, ex)
}
