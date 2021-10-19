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

	v1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"

	"github.com/go-logr/logr"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"
	wf "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/wfimpl"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
)

func checkMigrateRo(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()
	insId, targetNode := GetMigrateInfo(cluster)
	if insId == "" || targetNode == "" {
		return nil, nil
	}
	for id, insStatus := range cluster.Status.DBInstanceStatus {
		if id == insId && insId != cluster.Status.LeaderInstanceId && insStatus.NodeName != targetNode {
			return statemachine.CreateEvent(statemachine.EventName(define.WorkflowStateMigrateRo), nil), nil
		}
	}
	return nil, nil
}

func GetMigrateInfo(cluster *v1.MPDCluster) (insId string, targetNode string) {
	annotation := cluster.Annotations
	if annotation == nil {
		return "", ""
	}
	migrateInfo, ok := annotation[define.AnnotationMigrate]
	if !ok || migrateInfo == "" {
		return "", ""
	}
	info := strings.Split(migrateInfo, "|")
	if len(info) != 2 {
		return "", ""
	}
	return info[0], info[1]
}

func migrateRoMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "SharedStorageClusterMigrateRo", false, checkMigrateRo)
	return err
}

type GenerateMigrateTempId struct {
	wf.SharedStorageClusterStepBase
}

func (step *GenerateMigrateTempId) DoStep(ctx context.Context, logger logr.Logger) error {
	insId, targetNode := GetMigrateInfo(step.Resource)
	if insId == "" || targetNode == "" {
		return nil
	}
	return step.Service.GenerateTempRoId(ctx, step.Model, insId)
}

type EnsureRoMigrate struct {
	wf.SharedStorageClusterStepBase
}

func (step *EnsureRoMigrate) DoStep(ctx context.Context, logger logr.Logger) error {
	_, targetNode := GetMigrateInfo(step.Resource)
	if targetNode == "" {
		return nil
	}
	for _, tempRoIns := range step.Model.TempRoInses {
		tempRoIns.TargetNode = targetNode
	}
	return step.Service.EnsureNewRoUpToDate(ctx, step.Model)
}
