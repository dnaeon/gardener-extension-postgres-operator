// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package validation_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/gardener/gardener-extension-postgres-operator/pkg/apis/config"
	"github.com/gardener/gardener-extension-postgres-operator/pkg/apis/config/validation"
)

var _ = Describe("Validation Tests", Ordered, func() {
	It("should detect invalid config", func() {
		// TODO(dnaeon): implement me
		err := validation.Validate(config.PostgresConfig{})
		Expect(err).Should(HaveOccurred())
	})

	It("should successfully validate correct config", func() {
		// TODO(dnaeon): implement me
	})
})
