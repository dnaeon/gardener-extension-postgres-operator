package validation

import (
	"k8s.io/apimachinery/pkg/util/validation/field"

	"github.com/gardener/gardener-extension-postgres-operator/pkg/apis/config"
)

// Validate validates the given [config.PostgresConfig]
func Validate(cfg config.PostgresConfig) error {
	allErrs := make(field.ErrorList, 0)

	if cfg.Spec.VolumeSize.IsZero() {
		allErrs = append(allErrs, field.Required(field.NewPath("spec.volumeSize"), "no volume size specified"))
	}

	if cfg.Spec.Replicas <= 0 {
		allErrs = append(allErrs, field.Invalid(field.NewPath("spec.replicas"), cfg.Spec.Replicas, "invalid number of replicas"))
	}

	if len(cfg.Spec.Users) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("spec.users"), "no users specified"))
	}

	if len(cfg.Spec.Databases) == 0 {
		allErrs = append(allErrs, field.Required(field.NewPath("spec.databases"), "no databases specified"))
	}

	if cfg.Spec.PostgresVersion == "" {
		allErrs = append(allErrs, field.Required(field.NewPath("spec.postgresVersion"), "no postgres version specified"))
	}

	return allErrs.ToAggregate()
}
