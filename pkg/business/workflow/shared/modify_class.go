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

	"github.com/go-logr/logr"
	commondefine "gitlab.alibaba-inc.com/polar-as/polar-common-domain/define"
	mgr "gitlab.alibaba-inc.com/polar-as/polar-common-domain/manager"
	mpdv1 "gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/apis/mpd/v1"
	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/define"
	wf "gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/wfimpl"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/statemachine"
)

func checkModifyClass(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()
	if cluster.Spec.ClassInfoModifyTo.ClassName != "" && cluster.Spec.ClassInfoModifyTo.ClassName != cluster.Spec.ClassInfo.ClassName {
		return statemachine.CreateEvent(statemachine.EventName(define.WorkflowStateModifyClass), map[string]interface{}{
			"modifyClass": true,
		}), nil
	}
	return nil, nil
}

func modifyClassMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "SharedStorageClusterModifyClass", false, checkModifyClass)
	return err
}

type GenerateTempRoIds struct {
	wf.SharedStorageClusterStepBase
}

func (step *GenerateTempRoIds) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.GenerateTempRoIds(ctx, step.Model)
}

type InitTempRoMeta struct {
	wf.SharedStorageClusterStepBase
}

func (step *InitTempRoMeta) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.InitTempRoMeta(ctx, step.Model)
}

type CreateTempRoForRw struct {
	wf.SharedStorageClusterStepBase
}

func (step *CreateTempRoForRw) DoStep(ctx context.Context, logger logr.Logger) error {
	insId, targetNode := GetMigrateInfo(step.Resource)
	for _, ins := range step.Model.TempRoInses {
		if ins.PhysicalInsId == step.Model.RwIns.PhysicalInsId {
			if targetNode != "" && insId == step.Model.RwIns.InsId {
				ins.TargetNode = targetNode
			}
			return step.Service.CreateTempRoIns(ctx, step.Model, ins)
		}
	}
	return commondefine.CreateInterruptError(define.TempRoForRwMetaLost, nil)
}

type ConvertTempRoToRo struct {
	wf.SharedStorageClusterStepBase
}

func (step *ConvertTempRoToRo) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.ConvertTempRoForRwToRo(ctx, step.Model)
}

type SwitchNewRoToRw struct {
	wf.SharedStorageClusterStepBase
}

func (step *SwitchNewRoToRw) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.SwitchNewRoToRw(ctx, step.Model)
}

type UpdateModifyClassMeta struct {
	wf.SharedStorageClusterStepBase
}

func (step *UpdateModifyClassMeta) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.InitModifyClassMeta(step.Model)
}

type FlushParamsIfNecessary struct {
	wf.SharedStorageClusterStepBase
}

func (step *FlushParamsIfNecessary) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.FlushParamsIfNecessary(ctx, step.Model)
}

type DeleteOldRw struct {
	wf.SharedStorageClusterStepBase
}

func (step *DeleteOldRw) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.DeleteOldRw(ctx, step.Model)
}

type EnsureNewRoUpToDate struct {
	wf.SharedStorageClusterStepBase
}

func (step *EnsureNewRoUpToDate) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.EnsureNewRoUpToDate(ctx, step.Model)
}

type CleanModifyClassTempMeta struct {
	wf.SharedStorageClusterStepBase
}

func (step *CleanModifyClassTempMeta) DoStep(ctx context.Context, logger logr.Logger) error {
	res := step.Resource
	delete(res.Annotations, define.AnnotationTempRoIds)
	if res.Spec.ClassInfoModifyTo.ClassName != "" {
		res.Spec.ClassInfo = res.Spec.ClassInfoModifyTo
		res.Spec.ClassInfoModifyTo = mpdv1.InstanceClassInfo{}
	}
	return mgr.GetSyncClient().Update(ctx, res)
}

type DisableHA struct {
	wf.SharedStorageClusterStepBase
}

func (step *DisableHA) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.DisableHA(ctx, step.Model)
}

type EnableHA struct {
	wf.SharedStorageClusterStepBase
}

func (step *EnableHA) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.EnableHA(ctx, step.Model)
}

type EnsureCmRwAffinity struct {
	wf.SharedStorageClusterStepBase
}

func (step *EnsureCmRwAffinity) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.EnsureCmAffinity(ctx, step.Model)
}

type SaveParamsLastUpdateTime struct {
	wf.SharedStorageClusterStepBase
}

func (step *SaveParamsLastUpdateTime) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.SaveParamsLastUpdateTime(ctx, step.Model)
}
