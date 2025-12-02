// SPDX-FileCopyrightText: SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package actuator_test

import (
	"context"
	"path/filepath"
	"testing"

	extensionscontroller "github.com/gardener/gardener/extensions/pkg/controller"
	corev1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	resourcesv1alpha1 "github.com/gardener/gardener/pkg/apis/resources/v1alpha1"
	"github.com/gardener/gardener/pkg/client/kubernetes"
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	acidv1 "github.com/zalando/postgres-operator/pkg/apis/acid.zalan.do/v1"
	zalandov1 "github.com/zalando/postgres-operator/pkg/apis/zalando.org/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	configinstall "github.com/gardener/gardener-extension-postgres-operator/pkg/apis/config/install"
)

var (
	ctx       context.Context
	cancel    context.CancelFunc
	testEnv   *envtest.Environment
	cfg       *rest.Config
	k8sClient client.Client
	logger    logr.Logger
	k8sScheme = runtime.NewScheme()
)

func TestActuators(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Actuator Suite")
}

var _ = BeforeSuite(func() {
	logger = zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true))
	logf.SetLogger(logger)

	ctx, cancel = context.WithCancel(context.TODO())

	Expect(corev1beta1.AddToScheme(k8sScheme)).To(Succeed())
	Expect(extensionscontroller.AddToScheme(k8sScheme)).To(Succeed())
	Expect(resourcesv1alpha1.AddToScheme(k8sScheme)).To(Succeed())
	Expect(kubernetes.AddShootSchemeToScheme(k8sScheme)).To(Succeed())
	Expect(acidv1.AddToScheme(k8sScheme)).To(Succeed())
	Expect(zalandov1.AddToScheme(k8sScheme)).To(Succeed())
	configinstall.Install(k8sScheme)

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		Scheme: k8sScheme,
		CRDDirectoryPaths: []string{
			filepath.Join("..", "..", "test", "manifests", "crd", "extensions.gardener.cloud", "v1alpha1"),
			filepath.Join("..", "..", "test", "manifests", "crd", "resources.gardener.cloud", "v1alpha1"),
			filepath.Join("..", "..", "test", "manifests", "crd", "acid.zalan.do", "v1"),
		},
		ErrorIfCRDPathMissing: true,
	}

	var err error
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	k8sClient, err = client.New(cfg, client.Options{Scheme: k8sScheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())
})

var _ = AfterSuite(func() {
	By("tearing down the test environment")
	cancel()
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
