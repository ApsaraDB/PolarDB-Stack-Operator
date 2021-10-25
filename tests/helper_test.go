/*
*Copyright (c) 2019-2021, Alibaba Group Holding Limited;
*Licensed under the Apache License, Version 2.0 (the "License");
*you may not use this file except in compliance with the License.
*You may obtain a copy of the License at

*   http://www.apache.org/licenses/LICENSE-2.0

*Unless required by applicable law or agreed to in writing, software
*distributed under the License is distributed on an "AS IS" BASIS,
*WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
*See the License for the specific language governing permissions and
*limitations under the License.
 */

package tests

import (
	"context"
	globalmgr "github.com/ApsaraDB/PolarDB-Stack-Common/manager"
	mpdv1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"
	workflow_shared "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/workflow/shared"
	controllers "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/controllers/mpd"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/component-base/logs"
	"k8s.io/klog/klogr"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"testing"
)

/*
# setup kubebuilder for test
arch=amd64

# download the release
curl -L -O https://storage.googleapis.com/kubebuilder-release/kubebuilder_master_darwin_${arch}.tar.gz

# extract the archive
tar -zxvf kubebuilder_master_darwin_${arch}.tar.gz
mv kubebuilder_master_darwin_${arch} kubebuilder && sudo mv kubebuilder /usr/local/

# update your PATH to include /usr/local/kubebuilder/bin
export PATH=$PATH:/usr/local/kubebuilder/bin
*/

func TestReconcile(t *testing.T) {

	workflow_shared.Register()

	ctx := context.Background()

	env, cc, config, scheme := setupTestEnv(t, t.Name())
	t.Cleanup(func() { teardownTestEnv(t, env) })

	reconciler := setupManager(t, config, scheme)

	clusterNameSpace := "default"
	clusterName := "mpdcluster-open-test"

	cluster := &mpdv1.MPDCluster{}
	cluster.Namespace = clusterNameSpace
	cluster.Name = clusterName
	cluster.Spec = mpdv1.MPDClusterSpec{
		DBClusterType: mpdv1.MPDClusterSharedVol,
	}

	assert.NilError(t, errors.WithStack(reconciler.Client.Create(ctx, cluster)))
	t.Cleanup(func() { assert.Check(t, reconciler.Client.Delete(ctx, cluster)) })

	assert.NilError(t, cc.Get(ctx, client.ObjectKey{
		Namespace: clusterNameSpace,
		Name:      clusterName,
	}, cluster))

	req := ctrl.Request{
		types.NamespacedName{
			clusterNameSpace,
			clusterName,
		},
	}
	reconciler.Reconcile(req)
}

func setupTestEnv(t *testing.T, _ string) (*envtest.Environment, client.Client, *rest.Config, *runtime.Scheme) {
	//specify testEnv configuration
	testEnv := &envtest.Environment{
		CRDDirectoryPaths: []string{filepath.Join("..", "config", "crd", "bases")},
	}
	cfg, err := testEnv.Start()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Test environment started")

	scheme, err := createScheme()
	if err != nil {
		t.Fatal(err)
	}
	client, err := client.New(cfg, client.Options{Scheme: scheme})
	if err != nil {
		t.Fatal(err)
	}
	return testEnv, client, cfg, scheme
}

func teardownTestEnv(t *testing.T, testEnv *envtest.Environment) {
	if err := testEnv.Stop(); err != nil {
		t.Error(err)
	}
	t.Log("Test environment stopped")
}

func createScheme() (*runtime.Scheme, error) {

	// create a new scheme specifically for this manager
	scheme := runtime.NewScheme()

	// add standard resource types to the scheme
	if err := clientgoscheme.AddToScheme(scheme); err != nil {
		return nil, err
	}

	// add custom resource types to the default scheme
	if err := mpdv1.AddToScheme(scheme); err != nil {
		return nil, err
	}

	return scheme, nil
}

func setupManager(t *testing.T, cfg *rest.Config, scheme *runtime.Scheme) *controllers.MPDClusterReconciler {

	mgr, err := manager.New(cfg, manager.Options{
		Scheme: scheme,
	})
	if err != nil {
		t.Fatal(err)
	}
	globalmgr.RegisterManager(mgr)
	reconciler := &controllers.MPDClusterReconciler{
		Client: mgr.GetClient(),
		Log:    ctrl.Log.WithName("controllers").WithName("MpdCluster"),
		Scheme: mgr.GetScheme(),
	}
	if err := reconciler.SetupWithManager(mgr); err != nil {
		t.Fatal(err)
	}

	//contollerSetup(mgr)

	signals := ctrl.SetupSignalHandler()
	go func() {
		if err := mgr.Start(signals); err != nil {
			t.Error(err)
		}
	}()
	t.Log("Manager started")

	return reconciler
}

func prepareEnv(t *testing.T) client.Client {

	env, cc, config, scheme := setupTestEnv(t, t.Name())
	t.Cleanup(func() { teardownTestEnv(t, env) })

	_ = setupManager(t, config, scheme)

	return cc
}

func createResource(t *testing.T, ctx context.Context, cc client.Client, obj runtime.Object, name types.NamespacedName) {

	assert.NilError(t, errors.WithStack(cc.Create(ctx, obj)))
	t.Cleanup(func() { assert.Check(t, cc.Delete(ctx, obj)) })

	assert.NilError(t, cc.Get(ctx, client.ObjectKey{
		Namespace: name.Namespace,
		Name:      name.Name,
	}, obj))
}

func createResourceWithStatusUpdate(t *testing.T, ctx context.Context, cc client.Client, obj runtime.Object, name types.NamespacedName) {

	assert.NilError(t, errors.WithStack(cc.Create(ctx, obj)))
	assert.NilError(t, errors.WithStack(cc.Status().Update(ctx, obj)))
	t.Cleanup(func() { assert.Check(t, cc.Delete(ctx, obj)) })

	assert.NilError(t, cc.Get(ctx, client.ObjectKey{
		Namespace: name.Namespace,
		Name:      name.Name,
	}, obj))
}

func getLogger() logr.Logger {
	ctrl.SetLogger(klogr.New())
	logs.InitLogs()
	logger := ctrl.Log.WithName("test")
	return logger
}
