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
)

func checkMigrateRw(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()
	insId, targetNode := GetMigrateInfo(cluster)
	if insId == "" || targetNode == "" {
		return nil, nil
	}
	for id, insStatus := range cluster.Status.DBInstanceStatus {
		if id == insId && insId == cluster.Status.LeaderInstanceId && insStatus.NodeName != targetNode {
			return statemachine.CreateEvent(statemachine.EventName(define.WorkflowStateMigrateRw), nil), nil
		}
	}
	return nil, nil
}

func migrateRwMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "SharedStorageClusterMigrateRw", false, checkMigrateRw)
	return err
}
