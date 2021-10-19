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
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"
	wf "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/wfimpl"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
)

func checkRestartCluster(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()

	if cluster.Annotations != nil {
		r, ok := cluster.Annotations[define.AnnotationRestartCluster]
		if ok && (r == "1" || r == "T" || r == "true") {
			return statemachine.CreateEvent(statemachine.EventName(define.WorkflowStateRestartCluster), nil), nil
		}
	}
	return nil, nil
}

func restartClusterMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "SharedStorageClusterRestartCluster", false, checkRestartCluster)
	return err
}

type RestartCluster struct {
	wf.SharedStorageClusterStepBase
}

func (step *RestartCluster) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.RestartCluster(ctx, step.Model)
}
