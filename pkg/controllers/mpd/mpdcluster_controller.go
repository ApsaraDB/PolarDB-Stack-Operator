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

/*


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

package controllers

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"

	mgr "github.com/ApsaraDB/PolarDB-Stack-Common/manager"
	"github.com/ApsaraDB/PolarDB-Stack-Common/utils"

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/configuration"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	v1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	//_ "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/workflow/local"
	_ "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/workflow/shared"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/wfimpl"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/implement"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// MPDClusterReconciler reconciles a MPDCluster object
type MPDClusterReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mpd.polardb.aliyun.com,resources=mpdclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mpd.polardb.aliyun.com,resources=mpdclusters/status,verbs=get;update;patch
func (r *MPDClusterReconciler) Reconcile(req ctrl.Request) (result ctrl.Result, err error) {
	logger := r.Log.WithName(req.Namespace).WithName(req.Name)
	logger.Info("begin mpd reconcile")
	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			err = errors.New(fmt.Sprintf("get panic, err: %v", r))
			logger.Error(err, "panic recovered!", "stack", string(debug.Stack()))
			result = reconcile.Result{}
		}
	}()

	res := &v1.MPDCluster{}
	err = r.Get(context.TODO(), req.NamespacedName, res)
	if err != nil {
		return reconcile.Result{}, err
	}

	if needRemoveFinalizer, err := checkDeleteFinalizer(res, logger); needRemoveFinalizer {
		return reconcile.Result{RequeueAfter: 3 * time.Second}, err
	}

	// 检查是否需要恢复中断流程
	resource := &wfimpl.MpdClusterResource{KubeResource: implement.KubeResource{Resource: res}, Logger: logger}
	if res.Spec.DBClusterType == v1.MPDClusterSharedVol {
		if wfimpl.GetSharedStorageClusterWfManager().CheckRecovery(resource) {
			return reconcile.Result{}, nil
		}
		err = wfimpl.GetSharedStorageClusterStateMachine().RegisterLogger(logger).DoStateMainEnter(resource)
	} else {
		// todo 三节点
		logger.Info("controller localstore mainenter")
		err = wfimpl.GetLocalStorageClusterStateMachine().RegisterLogger(logger).DoStateMainEnter(resource)
	}

	if err != nil {
		logger.Error(err, "reconcile do state main fail!", "status", res.Status.ClusterStatus)
		return reconcile.Result{RequeueAfter: time.Second}, nil
	}
	logger.Info("reconcile succeed!", "status", res.Status.ClusterStatus)
	return result, err

}

func (r *MPDClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1.MPDCluster{}).
		WithOptions(controller.Options{
			MaxConcurrentReconciles: configuration.GetConfig().MaxConcurrentReconciles,
		}).
		WithEventFilter(&DebugPredicate{}).
		Complete(r)
}

func checkDeleteFinalizer(cluster *v1.MPDCluster, logger logr.Logger) (bool, error) {
	if !cluster.ObjectMeta.DeletionTimestamp.IsZero() && utils.ContainsString(cluster.ObjectMeta.Finalizers, define.MpdClusterFinalizer, nil) {
		service := business.NewSharedStorageClusterService(logger)
		model, err := service.GetByName(cluster.Name, cluster.Namespace)
		if err != nil {
			logger.Error(err, "get shared storage cluster model failed", "cluster", cluster.Name)
			return true, err
		}
		// 释放存储
		if err := service.ReleaseStorage(context.TODO(), model); err != nil {
			logger.Error(err, "mpd cluster release storage failed", "cluster", cluster.Name)
			return true, err
		}

		dmRes := &v1.MPDCluster{}
		err = mgr.GetSyncClient().Get(context.TODO(), client.ObjectKey{Namespace: cluster.Namespace, Name: cluster.Name}, dmRes)
		if err != nil {
			logger.Error(err, "get mpd cluster failed, err: %v", "cluster", dmRes.Name)
			return true, err
		}
		dmRes.ObjectMeta.Finalizers = utils.RemoveString(dmRes.ObjectMeta.Finalizers, define.MpdClusterFinalizer, nil)
		if err := mgr.GetSyncClient().Update(context.Background(), dmRes); err != nil {
			logger.Error(err, "remove mpd cluster finalizer failed", "cluster", dmRes.Name)
			return true, err
		}
		return true, nil
	}
	return false, nil
}
