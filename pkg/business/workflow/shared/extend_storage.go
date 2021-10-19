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

	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"
	wf "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/wfimpl"
)

func checkExtendStorage(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()

	if cluster.Annotations != nil {
		r, ok := cluster.Annotations[define.AnnotationExtendStorage]
		if ok && (r == "1" || r == "T" || r == "true") {
			return statemachine.CreateEvent("ExtendStorage", nil), nil
		}
	}

	return nil, nil
}

func extendStorageMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetSharedStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "ExtendStorageSharedStorageCluster", true, checkExtendStorage)
	return err
}

type ExtendStorageFs struct {
	wf.SharedStorageClusterStepBase
}

func (s *ExtendStorageFs) DoStep(ctx context.Context, logger logr.Logger) error {
	return s.Service.GrowStorage(ctx, s.Model)
}
