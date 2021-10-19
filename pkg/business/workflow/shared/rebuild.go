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
	"strings"

	wfdefine "github.com/ApsaraDB/PolarDB-Stack-Workflow/define"

	"github.com/ApsaraDB/PolarDB-Stack-Common/utils"
	v1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"

	mgr "github.com/ApsaraDB/PolarDB-Stack-Common/manager"

	"github.com/go-logr/logr"

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"

	wf "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/wfimpl"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
)

func checkRebuild(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()
	if cluster.Status.ClusterStatus == statemachine.StateRebuild {
		return statemachine.CreateEvent(statemachine.EventName(statemachine.StateRebuild), nil), nil
	}

	if cluster.Status.DBInstanceStatus == nil ||
		len(cluster.Status.DBInstanceStatus) == 0 ||
		cluster.Status.LeaderInstanceId == "" {
		return nil, nil
	}

	if cluster.Annotations != nil {
		r := cluster.Annotations[define.AnnotationForceRebuild]
		if r == "1" || r == "T" || r == "true" {
			return statemachine.CreateEvent(statemachine.EventName(statemachine.StateRebuild), nil), nil
		}
	}

	for _, insStatus := range cluster.Status.DBInstanceStatus {
		if strings.ToLower(insStatus.CurrentState.State) != "failed" {
			return nil, nil
		}
	}

	return statemachine.CreateEvent(statemachine.EventName(statemachine.StateRebuild), nil), nil
}

func rebuildMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "SharedStorageClusterRebuild", true, checkRebuild)
	return err
}

type CleanOldTempMeta struct {
	wf.SharedStorageClusterStepBase
}

func (step *CleanOldTempMeta) DoStep(ctx context.Context, logger logr.Logger) error {
	return cleanTempMeta(step.Resource, []string{define.AnnotationForceRebuild})
}

type SetRebuildTag struct {
	wf.SharedStorageClusterStepBase
}

func (step *SetRebuildTag) DoStep(ctx context.Context, logger logr.Logger) error {
	mpdCluster := step.Resource
	if mpdCluster.Annotations == nil {
		mpdCluster.Annotations = map[string]string{}
	}
	mpdCluster.Annotations[define.AnnotationForceRebuild] = "T"
	return mgr.GetSyncClient().Update(context.TODO(), mpdCluster)
}

type RemoveClusterManager struct {
	wf.SharedStorageClusterStepBase
}

func (step *RemoveClusterManager) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.DeleteCm(ctx, step.Model)
}

type RemoveAllInsPod struct {
	wf.SharedStorageClusterStepBase
}

func (step *RemoveAllInsPod) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.DeleteAllInsPod(step.Model, ctx)
}

type CleanTempRoMeta struct {
	wf.SharedStorageClusterStepBase
}

func (step *CleanTempRoMeta) DoStep(ctx context.Context, logger logr.Logger) error {
	mpdCluster := step.Resource
	var tempRoPods []string
	for _, ins := range mpdCluster.Status.DBInstanceStatus {
		for _, instance := range mpdCluster.Status.DBInstanceStatus {
			if ins.PhysicalInsId == instance.PhysicalInsId && ins.InsId > instance.InsId {
				tempRoPods = append(tempRoPods, ins.InsId)
			}
		}
	}
	for _, tempRoId := range tempRoPods {
		delete(step.Resource.Status.DBInstanceStatus, tempRoId)
	}
	return mgr.GetSyncClient().Status().Update(context.TODO(), mpdCluster)
}

type CleanAllTempMeta struct {
	wf.SharedStorageClusterStepBase
}

func (step *CleanAllTempMeta) DoStep(ctx context.Context, logger logr.Logger) error {
	return cleanTempMeta(step.Resource, []string{})
}

func cleanTempMeta(mpdCluster *v1.MPDCluster, ignoreAnnotations []string) error {
	allAnnotations := []string{
		define.AnnotationForceRebuild,
		define.AnnotationMigrate,
		define.AnnotationFlushParams,
		define.AnnotationTempRoIds,
		define.AnnotationRestartIns,
		define.AnnotationRestartCluster,
		define.AnnotationFlushParamsRestartCluster,
		define.AnnotationSwitchRw,
		define.AnnotationRemoveIns,
		define.AnnotationExtendStorage,
		define.DefaultWfConf[wfdefine.WFInterruptToRecover],
		define.DefaultWfConf[wfdefine.WFInterruptPrevious],
		define.DefaultWfConf[wfdefine.WFInterruptMessage],
		define.DefaultWfConf[wfdefine.WFInterruptReason],
	}
	var cleanAnnotations []string
	for _, annotation := range allAnnotations {
		if !utils.ContainsString(ignoreAnnotations, annotation, nil) {
			cleanAnnotations = append(cleanAnnotations, annotation)
		}
	}
	haveClean := false
	if mpdCluster.Annotations != nil {
		for _, annotation := range cleanAnnotations {
			if _, ok := mpdCluster.Annotations[annotation]; ok {
				delete(mpdCluster.Annotations, annotation)
				if !haveClean {
					haveClean = true
				}
			}
		}
	}
	if mpdCluster.Spec.VersionCfgModifyTo.VersionName != "" {
		haveClean = true
		mpdCluster.Spec.VersionCfgModifyTo = v1.VersionInfo{}
	}
	if mpdCluster.Spec.ClassInfoModifyTo.ClassName != "" {
		haveClean = true
		mpdCluster.Spec.ClassInfoModifyTo = v1.InstanceClassInfo{}
	}

	if haveClean {
		err := mgr.GetSyncClient().Update(context.TODO(), mpdCluster)
		if err != nil {
			return err
		}
	}
	return nil
}
