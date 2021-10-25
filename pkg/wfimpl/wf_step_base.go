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

package wfimpl

import (
	"context"

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"

	v1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"
	wfdefine "github.com/ApsaraDB/PolarDB-Stack-Workflow/define"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/wfengine"
	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/types"

	mgr "github.com/ApsaraDB/PolarDB-Stack-Common/manager"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/domain"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/service"
)

type SharedStorageClusterStepBase struct {
	wfengine.StepAction
	Resource *v1.MPDCluster
	Service  *service.SharedStorageClusterService
	Model    *domain.SharedStorageCluster
}

func (s *SharedStorageClusterStepBase) Init(ctx map[string]interface{}, logger logr.Logger) error {
	name := ctx[define.DefaultWfConf[wfdefine.WorkFlowResourceName]].(string)
	ns := ctx[define.DefaultWfConf[wfdefine.WorkFlowResourceNameSpace]].(string)

	kube := &v1.MPDCluster{}
	err := mgr.GetSyncClient().Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, kube)
	if err != nil {
		return err
	}
	s.Resource = kube
	s.Service = business.NewSharedStorageClusterService(logger)
	useModifyClass := false
	if val, ok := ctx["modifyClass"]; ok {
		useModifyClass = val.(bool)
	}
	useUpgradeVersion := false
	if val, ok := ctx["upgrade"]; ok {
		useUpgradeVersion = val.(bool)
	}
	s.Model = s.Service.GetByData(kube, useModifyClass, useUpgradeVersion)
	return nil
}

func (s *SharedStorageClusterStepBase) DoStep(ctx context.Context, logger logr.Logger) error {
	panic("implement me")
}

func (s *SharedStorageClusterStepBase) Output(logger logr.Logger) map[string]interface{} {
	return map[string]interface{}{}
}

type LocalStorageClusterStepBase struct {
	wfengine.StepAction
	Resource *v1.MPDCluster
	Service  *service.LocalStorageClusterService
	Model    *domain.LocalStorageCluster
}

func (s *LocalStorageClusterStepBase) Init(ctx map[string]interface{}, logger logr.Logger) error {
	logger.Info("wf_step_base Init")
	name := ctx[define.DefaultWfConf[wfdefine.WorkFlowResourceName]].(string)
	ns := ctx[define.DefaultWfConf[wfdefine.WorkFlowResourceNameSpace]].(string)

	kube := &v1.MPDCluster{}
	err := mgr.GetSyncClient().Get(context.TODO(), types.NamespacedName{Name: name, Namespace: ns}, kube)
	if err != nil {
		return err
	}
	s.Resource = kube
	s.Service = business.NewLocalStorageClusterService(logger)
	// fetch data from k8s before each step start
	s.Model = s.Service.GetByData(kube, false, false)
	return nil
}

func (s *LocalStorageClusterStepBase) DoStep(ctx context.Context, logger logr.Logger) error {
	panic("implement me")
}

func (s *LocalStorageClusterStepBase) Output(logger logr.Logger) map[string]interface{} {
	return map[string]interface{}{}
}
