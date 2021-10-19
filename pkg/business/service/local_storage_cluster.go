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


package service

import (
	"context"
	"github.com/go-logr/logr"
	commondomain "github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/domain"
)

type LocalStorageClusterService struct {
	Repository            domain.ILocalStorageClusterRepository
	PodManager            commondomain.IEnginePodManager
	IDGenerator			  commondomain.IIdGenerator
	PortGenerator		  commondomain.IPortGenerator
	ManagerClient         commondomain.IManagerClient
	logger                logr.Logger
}

func NewLocalStorageClusterService(
	repository domain.ILocalStorageClusterRepository,
	podManager commondomain.IEnginePodManager,
	idGenerator commondomain.IIdGenerator,
	portGenerator commondomain.IPortGenerator,
	managerClient commondomain.IManagerClient,
	logger logr.Logger,
) *LocalStorageClusterService {
	return &LocalStorageClusterService{
		Repository:			   repository,
		PodManager:            podManager,
		IDGenerator: 		   idGenerator,
		PortGenerator: 		   portGenerator,
		ManagerClient:		   managerClient,
		logger:                logger,
	}
}

func (s *LocalStorageClusterService) InitStatusInfo(cluster *domain.LocalStorageCluster) error {
	// produce db info: id, name etc.
	if err := cluster.InitDbInsMeta(); err != nil {
		return err
	}
	// instance info sync to k8s mpd cluster
	if err := s.Repository.Update(cluster); err != nil {
		return err
	}
	return nil
}

func (s *LocalStorageClusterService) CreatePod(ctx context.Context, cluster *domain.LocalStorageCluster, ins *commondomain.DbIns) error {
	err := ins.CreatePod(ctx)
	return err
}

func (s *LocalStorageClusterService) InstallDBEngine(ctx context.Context, ins *commondomain.DbIns) error {
	err := ins.InstallDbIns(ctx)
	return err
}

func (s *LocalStorageClusterService) GetByData(data interface{}, useModifyClass bool, useUpgradeVersion bool) *domain.LocalStorageCluster {
	cluster := s.Repository.GetByData(data, useModifyClass, useUpgradeVersion)
	s.setAdapters(cluster)
	return cluster
}

func (s *LocalStorageClusterService) setAdapters(cluster *domain.LocalStorageCluster) {
	cluster.Init(
		s.PodManager,
		s.IDGenerator,
		s.PortGenerator,
		s.ManagerClient,
		s.logger,
	)
}
