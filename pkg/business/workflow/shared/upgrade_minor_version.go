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
	mgr "github.com/ApsaraDB/PolarDB-Stack-Common/manager"
	mpdv1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"
	wf "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/wfimpl"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
)

func checkUpgradeMinorVersion(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()
	if cluster.Spec.VersionCfgModifyTo.VersionName != "" && cluster.Spec.VersionCfgModifyTo.VersionName != cluster.Spec.VersionCfg.VersionName {
		return statemachine.CreateEvent(statemachine.EventName(define.WorkflowStateUpgradeMinorVersion), map[string]interface{}{
			"upgrade": true,
		}), nil
	}
	return nil, nil
}

func upgradeMinorVersionMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "SharedStorageClusterUpgradeMinorVersion", false, checkUpgradeMinorVersion)
	return err
}

type InitUpgradeImages struct {
	wf.SharedStorageClusterStepBase
}

func (step *InitUpgradeImages) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.InitImages(step.Model)
}

type CleanUpgradeTempMeta struct {
	wf.SharedStorageClusterStepBase
}

func (step *CleanUpgradeTempMeta) DoStep(ctx context.Context, logger logr.Logger) error {
	res := step.Resource
	delete(res.Annotations, define.AnnotationTempRoIds)
	if res.Spec.VersionCfgModifyTo.VersionName != "" {
		res.Spec.VersionCfg = res.Spec.VersionCfgModifyTo
		res.Spec.VersionCfgModifyTo = mpdv1.VersionInfo{}
	}
	return mgr.GetSyncClient().Update(ctx, res)
}

type UpgradeCmVersion struct {
	wf.SharedStorageClusterStepBase
}

func (step *UpgradeCmVersion) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Model.UpgradeCmVersion(ctx, step.Resource.Spec.VersionCfgModifyTo.ClusterManagerImage)
}
