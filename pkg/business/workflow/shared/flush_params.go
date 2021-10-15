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
	mgr "gitlab.alibaba-inc.com/polar-as/polar-common-domain/manager"
	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/define"
	wf "gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/wfimpl"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/statemachine"
)

func checkFlushParams(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()

	if cluster.Annotations != nil {
		r, ok := cluster.Annotations[define.AnnotationFlushParams]
		if ok && (r == "1" || r == "T" || r == "true") {
			return statemachine.CreateEvent(statemachine.EventName(define.WorkflowStateFlushParams), nil), nil
		}
	}
	return nil, nil
}

func flushParamsMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "SharedStorageClusterFlushParams", false, checkFlushParams)
	return err
}

type FlushParams struct {
	wf.SharedStorageClusterStepBase
}

func (step *FlushParams) DoStep(ctx context.Context, logger logr.Logger) error {
	if needRestart, err := step.Service.FlushClusterParams(ctx, step.Model); err != nil {
		return err
	} else if needRestart {
		step.Resource.Annotations[define.AnnotationFlushParamsRestartCluster] = "true"
		return mgr.GetSyncClient().Update(ctx, step.Resource)
	}
	return nil
}

type RestartClusterIfNeed struct {
	wf.SharedStorageClusterStepBase
}

func (step *RestartClusterIfNeed) DoStep(ctx context.Context, logger logr.Logger) error {
	if step.Resource.Annotations[define.AnnotationFlushParamsRestartCluster] == "true" {
		return step.Service.RestartCluster(ctx, step.Model)
	}
	return nil
}
