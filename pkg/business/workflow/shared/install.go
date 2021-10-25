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

package workflow_shared

import (
	"context"

	mgr "github.com/ApsaraDB/PolarDB-Stack-Common/manager"
	"github.com/ApsaraDB/PolarDB-Stack-Common/utils"
	"k8s.io/apimachinery/pkg/types"

	commondefine "github.com/ApsaraDB/PolarDB-Stack-Common/define"

	commondomain "github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"

	wf "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/wfimpl"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/go-logr/logr"
)

func checkInstall(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()
	if cluster.Status.ClusterStatus == "Init" || cluster.Status.ClusterStatus == "" || string(cluster.Status.ClusterStatus) == string(statemachine.StateCreating) {
		return statemachine.CreateEvent(statemachine.EventName(statemachine.StateCreating), nil), nil
	}
	return nil, nil
}

func installMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "CreateSharedStorageCluster", false, checkInstall)
	return err
}

type InitMeta struct {
	wf.SharedStorageClusterStepBase
}

func (step *InitMeta) DoStep(ctx context.Context, logger logr.Logger) error {
	if err := step.Service.InitMeta(step.Model); err != nil {
		return err
	}

	err := mgr.GetSyncClient().Get(context.TODO(), types.NamespacedName{Name: step.Resource.Name, Namespace: step.Resource.Namespace}, step.Resource)
	if err != nil {
		return err
	}
	if utils.ContainsString(step.Resource.Finalizers, define.MpdClusterFinalizer, nil) {
		return nil
	}
	step.Resource.Finalizers = append(step.Resource.Finalizers, define.MpdClusterFinalizer)
	return mgr.GetSyncClient().Update(context.TODO(), step.Resource)
}

type PrepareStorage struct {
	wf.SharedStorageClusterStepBase
}

func (step *PrepareStorage) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Model.UseStorage(ctx, true)
}

type CreateRwPod struct {
	wf.SharedStorageClusterStepBase
}

func (step *CreateRwPod) DoStep(ctx context.Context, logger logr.Logger) error {
	if err := step.Service.CreateRwIns(ctx, step.Model, step.Model.RwIns); err != nil {
		return err
	}
	if err := step.Model.AddInsToClusterManager(ctx, step.Model.RwIns.InsId); err != nil {
		return err
	}
	return nil
}

type CreateRoPods struct {
	wf.SharedStorageClusterStepBase
}

func (step *CreateRoPods) DoStep(ctx context.Context, logger logr.Logger) error {
	for _, ins := range step.Model.RoInses {
		if err := step.Service.CreateRoIns(ctx, step.Model, ins); err != nil {
			return err
		}
		if err := step.Model.AddInsToClusterManager(ctx, "", ins.InsId); err != nil {
			return err
		}
	}
	return nil
}

type CreateNetwork struct {
	wf.SharedStorageClusterStepBase
}

func (step *CreateNetwork) DoStep(ctx context.Context, logger logr.Logger) error {
	return nil
}

type CreateClusterManager struct {
	wf.SharedStorageClusterStepBase
}

func (step *CreateClusterManager) DoStep(ctx context.Context, logger logr.Logger) error {
	pluginConf := map[string]interface{}{
		"type": "mpd",
		"polar_stack": map[string]interface{}{
			"storage_api_info": map[string]string{
				"svc_name":      commondefine.StorageServiceName,
				"svc_namespace": commondefine.StorageServiceNamespace,
				"pvc_name":      step.Model.StorageInfo.DiskID,
				"pvc_namespace": step.Model.Namespace,
				"lock_svc_url":  "/pvcs/lock",
				"get_topo_url":  "/pvcs/topo",
			},
		},
	}
	consensusPort, err := step.Model.PortGenerator.GetNextClusterExternalPort()
	if err != nil {
		logger.Error(err, "Cannot generate consensus port.")
		consensusPort = 5001
	}
	return business.NewCmCreatorService(logger).CreateClusterManager(
		ctx,
		step.Resource,
		commondomain.CmWorkModePure,
		step.Model.LogicInsId,
		step.Model.RwIns.PhysicalInsId,
		step.Model.ImageInfo.Images[define.ClusterManagerImageName],
		step.Model.ClusterManager.Port,
		pluginConf,
		consensusPort,
	)
}

type UpdateRunningStatus struct {
	wf.SharedStorageClusterStepBase
}

func (step *UpdateRunningStatus) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.UpdateRunningStatus(step.Model.Name, step.Model.Namespace)
}
