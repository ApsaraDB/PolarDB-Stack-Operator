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

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"
	wf "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/wfimpl"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/go-logr/logr"
)

func checkRemoveRo(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()
	annotation := cluster.Annotations
	if annotation == nil {
		return nil, nil
	}
	removeInsId, ok := annotation[define.AnnotationRemoveIns]
	if !ok || removeInsId == "" {
		return nil, nil
	}

	return statemachine.CreateEvent(statemachine.EventName(define.WorkflowStateRemoveRo), nil), nil
}

func removeRoMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "SharedStorageClusterRemoveRo", false, checkRemoveRo)
	return err
}

type RemoveRo struct {
	wf.SharedStorageClusterStepBase
}

func (step *RemoveRo) DoStep(ctx context.Context, logger logr.Logger) error {
	res := step.Resource
	annotation := res.Annotations
	if annotation == nil {
		return nil
	}
	removeInsId, ok := annotation[define.AnnotationRemoveIns]
	if !ok || removeInsId == "" {
		return nil
	}

	var removeInsPhyId string
	for insId, ins := range step.Model.RoInses {
		if insId == removeInsId {
			removeInsPhyId = ins.PhysicalInsId
		}
	}
	if removeInsPhyId == "" {
		return nil
	}

	return step.Service.DeleteOldIns(ctx, step.Model, removeInsPhyId, "", true, true)
}
