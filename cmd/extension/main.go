// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"os"

	"github.com/urfave/cli/v3"
	ctrl "sigs.k8s.io/controller-runtime"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	managercmd "github.com/gardener/gardener-extension-postgres-operator/cmd/extension/internal/manager"
	"github.com/gardener/gardener-extension-postgres-operator/pkg/version"
)

func main() {
	app := &cli.Command{
		Name:                  "gardener-extension-postgres-operator",
		Version:               version.Version,
		EnableShellCompletion: true,
		Usage:                 "Gardener Extension for PostgreSQL",
		Commands: []*cli.Command{
			managercmd.New(),
		},
	}

	ctx := ctrl.SetupSignalHandler()
	if err := app.Run(ctx, os.Args); err != nil {
		ctrllog.Log.Error(err, "failed to start extension")
		os.Exit(1)
	}
}
