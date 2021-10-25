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

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/domain"

	"github.com/pkg/errors"

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"
	wf "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/wfimpl"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/go-logr/logr"
)

func checkRebuildRo(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()

	for _, status := range cluster.Status.DBInstanceStatus {
		if status.InsId == cluster.Status.LeaderInstanceId && strings.ToLower(status.CurrentState.State) == "failed" {
			return nil, nil
		}
	}

	for _, status := range cluster.Status.DBInstanceStatus {
		if status.InsId != cluster.Status.LeaderInstanceId && strings.ToLower(status.CurrentState.State) == "failed" {
			return statemachine.CreateEvent(statemachine.EventName(define.WorkflowStateRebuildRo), nil), nil
		}
	}

	return nil, nil
}

func rebuildRoMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "SharedStorageClusterRebuildRo", false, checkRebuildRo)
	return err
}

type GenerateRebuildRoTempId struct {
	wf.SharedStorageClusterStepBase
}

func (step *GenerateRebuildRoTempId) DoStep(ctx context.Context, logger logr.Logger) error {
	cluster := step.Resource
	for insId, status := range cluster.Status.DBInstanceStatus {
		if status.InsId != cluster.Status.LeaderInstanceId && strings.ToLower(status.CurrentState.State) == "failed" {
			return step.Service.GenerateTempRoId(ctx, step.Model, insId)
		}
	}
	return errors.New("no failed ro was found")
}

type StopOldRo struct {
	wf.SharedStorageClusterStepBase
}

func (step *StopOldRo) DoStep(ctx context.Context, logger logr.Logger) error {
	cluster := step.Model
	phyId, newInsId := getTempRoId(cluster)
	return cluster.DeleteOldIns(ctx, phyId, newInsId, false, false)
}

func getTempRoId(model *domain.SharedStorageCluster) (string, string) {
	for phyId, insId := range model.TempRoIds {
		return phyId, insId
	}
	return "", ""
}
