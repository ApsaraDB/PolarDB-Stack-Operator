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


package controllers

import (
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	"gitlab.alibaba-inc.com/polar-as/polar-common-domain/define"
)

type PolarOptLead struct {
	RestConfig      *rest.Config
	KubeClient      client.Client
	InjectFunctions inject.Func
	RsMapper        meta.RESTMapper
	OwnedMgr        manager.Manager
}

func (lead *PolarOptLead) NeedLeaderElection() bool {
	return true
}

func (lead *PolarOptLead) Start(stop <-chan struct{}) error {
	logger := ctrl.Log.WithName("leader election runnables")
	logger.Info("----------***---start self define lead start!!!!-------------")
	logger.Info("----------***---set define.IsLeader = true -------------")
	define.IsLeader = true
	<-stop
	logger.Info("----------***---leader is stop!")
	logger.Info("----------***---set define.IsLeader = false -------------")
	define.IsLeader = false
	return nil
}

func (lead *PolarOptLead) InjectConfig(config *rest.Config) error {
	lead.RestConfig = config
	return nil
}

func (lead *PolarOptLead) InjectClient(client client.Client) error {
	lead.KubeClient = client
	return nil
}

func (lead *PolarOptLead) InjectAPIReader(client.Reader) error {

	return nil
}

func (lead *PolarOptLead) InjectScheme(scheme *runtime.Scheme) error {
	return nil
}

func (lead *PolarOptLead) InjectCache(cache cache.Cache) error {
	return nil
}

func (lead *PolarOptLead) InjectFunc(f inject.Func) error {
	lead.InjectFunctions = f
	return nil
}

func (lead *PolarOptLead) InjectStopChannel(<-chan struct{}) error {
	return nil
}

func (lead *PolarOptLead) InjectMapper(mapper meta.RESTMapper) error {
	lead.RsMapper = mapper
	return nil
}
