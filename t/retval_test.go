package t

import (
	"context"
	"github.com/go-logr/logr"
	gmg "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/rest"
	"k8s.io/klog/v2/ktesting"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"testing"
)

func TestRetValue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	restCfg, tearDownFn := setupEnvtest(t)
	defer tearDownFn(t)

	var log = ctrl.Log.WithName("ret-error-demo")
	kl := ktesting.NewLogger(t, ktesting.NewConfig())
	ctrl.SetLogger(kl)

	manager, err := ctrl.NewManager(restCfg, ctrl.Options{})
	assert.NoError(t, err)

	go func() {
		assert.NoError(t, manager.Start(ctx))
	}()

	keyValid := client.ObjectKey{Name: "a-name", Namespace: "default"}
	keyBadNS := client.ObjectKey{Name: "a-name", Namespace: "in/valid"}

	report(ctx, manager, keyValid, makeUnstructured("ReplicaSet", "apps/v1"), log)
	report(ctx, manager, keyValid, makeUnstructured("Badger", "apps/v1"), log)
	report(ctx, manager, keyBadNS, makeUnstructured("Gherkin", "inexistent.group.com/v1"), log)
}

func makeUnstructured(kind string, apiVersion string) *unstructured.Unstructured {
	rs := &unstructured.Unstructured{}
	rs.SetKind(kind)
	rs.SetAPIVersion(apiVersion)
	return rs
}

func report(ctx context.Context, manager manager.Manager, key client.ObjectKey, obj *unstructured.Unstructured, log logr.Logger) {
	clientCtx, clientCancel := context.WithTimeout(ctx, time.Second)
	defer clientCancel()
	err := manager.GetClient().Get(clientCtx, key, obj)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Kind is present but object was not found", "kind", obj.GetObjectKind().GroupVersionKind().Kind)
	} else if err != nil && meta.IsNoMatchError(err) {
		log.Info("Kind does not exist", "kind", obj.GetObjectKind().GroupVersionKind().Kind)
	} else if err != nil {
		log.Error(err, "Unexpected error")
	}
}

func setupEnvtest(t *testing.T) (*rest.Config, func(t *testing.T)) {
	t.Log("Setup envtest")

	g := gmg.NewWithT(t)
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{"testdata"},
	}

	cfg, err := testEnv.Start()
	g.Expect(err).NotTo(gmg.HaveOccurred())
	g.Expect(cfg).NotTo(gmg.BeNil())

	teardownFunc := func(t *testing.T) {
		t.Log("Stop envtest")
		g.Expect(testEnv.Stop()).To(gmg.Succeed())
	}

	return cfg, teardownFunc
}
