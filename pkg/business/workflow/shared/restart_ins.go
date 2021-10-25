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

func checkRestartIns(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()

	if cluster.Annotations != nil {
		insId, ok := cluster.Annotations[define.AnnotationRestartIns]
		if ok && insId != "" {
			return statemachine.CreateEvent(statemachine.EventName(define.WorkflowStateRestartIns), map[string]interface{}{
				"insId": insId,
			}), nil
		}
	}
	return nil, nil
}

func restartInsMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "SharedStorageClusterRestartIns", false, checkRestartIns)
	return err
}

type RestartIns struct {
	wf.SharedStorageClusterStepBase
}

func (step *RestartIns) DoStep(ctx context.Context, logger logr.Logger) error {
	insId, ok := step.Resource.Annotations[define.AnnotationRestartIns]
	if !ok || insId == "" {
		return nil
	}
	return step.Service.RestartIns(ctx, step.Model, insId)
}
