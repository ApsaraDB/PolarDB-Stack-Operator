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


package workflow_local

import (
	"context"

	"github.com/go-logr/logr"
	wf "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/wfimpl"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
)

func checkInstall(obj statemachine.StateResource) (*statemachine.Event, error) {
	cluster := obj.(*wf.MpdClusterResource).GetMpdCluster()
	if cluster.Status.ClusterStatus == "Init" || cluster.Status.ClusterStatus == "" || string(cluster.Status.ClusterStatus) == string(statemachine.StateCreating) {
		return statemachine.CreateEvent("Create", nil), nil
	}
	return nil, nil
}

func installMainEnter(obj statemachine.StateResource) error {
	resourceWf, err := wf.GetLocalStorageClusterWfManager().CreateResourceWorkflow(obj)
	if err != nil {
		return err
	}
	err = resourceWf.CommonWorkFlowMainEnter(context.TODO(), obj, "CreateLocalStorageDBCluster", false, checkInstall)
	return err
}

type InitStatusInfo struct {
	wf.LocalStorageClusterStepBase
}

func (step *InitStatusInfo) DoStep(ctx context.Context, logger logr.Logger) error {
	logger.Info("InitStatusInfo DoStep")
	return step.Service.InitStatusInfo(step.Model)
}

type CreatePod struct {
	wf.LocalStorageClusterStepBase
}

func (step *CreatePod) DoStep(ctx context.Context, logger logr.Logger) error {
	logger.Info("CreatePod DoStep")
	return step.Service.CreatePod(ctx, step.Model, step.Model.Ins)
}

type InstallDBEngine struct {
	wf.LocalStorageClusterStepBase
}

func (step *InstallDBEngine) DoStep(ctx context.Context, logger logr.Logger) error {
	logger.Info("InstallDBEngine DoStep")
	return step.Service.InstallDBEngine(ctx, step.Model.Ins)
}