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
	mpdv1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"
	wf "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/wfimpl"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
)

func checkAddRo(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()
	currentRoCount := 0
	for _, ins := range cluster.Status.DBInstanceStatus {
		if ins.InsType == mpdv1.MPDClusterInstanceTypeRO {
			currentRoCount++
		}
	}
	if cluster.Spec.FollowerNum > currentRoCount {
		return statemachine.CreateEvent(statemachine.EventName(define.WorkflowStateAddRo), nil), nil
	}

	return nil, nil
}

func addRoMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "SharedStorageClusterAddRo", false, checkAddRo)
	return err
}

type GenerateAddRoTempId struct {
	wf.SharedStorageClusterStepBase
}

func (step *GenerateAddRoTempId) DoStep(ctx context.Context, logger logr.Logger) error {
	return step.Service.GenerateNewRoTempId(ctx, step.Model)
}
