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


package monitor

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	commondefine "gitlab.alibaba-inc.com/polar-as/polar-common-domain/define"

	"github.com/go-logr/logr"
	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/business"
	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/business/service"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
)

func CreateSharedStorageClusterStatusSyncMonitor(logger logr.Logger, period time.Duration) *SharedStorageClusterStatusSyncMonitor {
	log := logger.WithName("SharedStorageClusterStatusSyncMonitor")
	return &SharedStorageClusterStatusSyncMonitor{
		logger:  log,
		period:  period,
		service: business.NewSharedStorageClusterService(log),
	}
}

type SharedStorageClusterStatusSyncMonitor struct {
	logger  logr.Logger
	period  time.Duration
	service *service.SharedStorageClusterService
}

func (m *SharedStorageClusterStatusSyncMonitor) Start(stop <-chan struct{}) {
	defer utilruntime.HandleCrash()
	m.logger.Info("start SharedStorageClusterStatusSyncMonitor")
	defer m.logger.Info("shutting down SharedStorageClusterStatusSyncMonitor")
	go wait.Until(m.doMonitor, m.period, stop)
	<-stop
}

func (m *SharedStorageClusterStatusSyncMonitor) doMonitor() {
	if !commondefine.IsLeader {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			for _, fn := range utilruntime.PanicHandlers {
				fn(r)
			}
			debug.PrintStack()
			err := fmt.Errorf("get panic, err: %v", r)
			m.logger.Error(err, "SharedStorageClusterStatusSyncMonitor panic recovered!")
		}
	}()
	m.onAction()
}

func (m *SharedStorageClusterStatusSyncMonitor) onAction() {
	begin := time.Now()
	defer func() {
		m.logger.V(5).Info(fmt.Sprintf("monitor action spend: [%v]s", time.Now().Sub(begin).Seconds()))
	}()
	m.syncSharedStorageClusterStatus()
}

func (m *SharedStorageClusterStatusSyncMonitor) syncSharedStorageClusterStatus() {
	list, err := m.service.GetAll()
	if err != nil {
		m.logger.Error(err, "GetAll SharedStorageCluster error")
		return
	}
	for _, item := range list {
		if err := m.service.SyncInsStateFromClusterManager(context.TODO(), item); err != nil {
			m.logger.Error(err, "SyncInsStateFromClusterManager failed", "cluster", item)
		}
	}
}
