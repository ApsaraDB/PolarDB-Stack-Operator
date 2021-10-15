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
	"sort"
	"strconv"
	"time"

	"gitlab.alibaba-inc.com/polar-as/polar-common-domain/utils/waitutil"

	"github.com/go-logr/logr"
	commondomain "gitlab.alibaba-inc.com/polar-as/polar-common-domain/business/domain"
	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/business/domain"
)

func NewShardStorageClusterService(
	repository domain.ISharedStorageClusterRepository,
	templateQuery commondomain.IEngineParamsTemplateQuery,
	paramClassQuery commondomain.IEngineParamsClassQuery,
	paramsRepo commondomain.IEngineParamsRepository,
	minorVersionQuery commondomain.IMinorVersionQuery,
	accountRepo commondomain.IAccountRepository,
	idGenerator commondomain.IIdGenerator,
	portGenerator commondomain.IPortGenerator,
	managerClient commondomain.IManagerClient,
	storageManager commondomain.IStorageManager,
	podManager commondomain.IEnginePodManager,
	classQuery commondomain.IClassQuery,
	cmClient commondomain.IClusterManagerClient,
	clusterManagerRemover commondomain.IClusterManagerRemover,
	logger logr.Logger,
) *SharedStorageClusterService {
	return &SharedStorageClusterService{
		Repository:            repository,
		ParamsTemplateQuery:   templateQuery,
		ParamsClassQuery:      paramClassQuery,
		ParamsRepo:            paramsRepo,
		MinorVersionQuery:     minorVersionQuery,
		AccountRepo:           accountRepo,
		IdGenerator:           idGenerator,
		PortGenerator:         portGenerator,
		ManagerClient:         managerClient,
		StorageManager:        storageManager,
		PodManager:            podManager,
		ClassQuery:            classQuery,
		ClusterManagerClient:  cmClient,
		ClusterManagerRemover: clusterManagerRemover,
		logger:                logger,
	}
}

type SharedStorageClusterService struct {
	Repository            domain.ISharedStorageClusterRepository
	ParamsTemplateQuery   commondomain.IEngineParamsTemplateQuery
	ParamsClassQuery      commondomain.IEngineParamsClassQuery
	ParamsRepo            commondomain.IEngineParamsRepository
	MinorVersionQuery     commondomain.IMinorVersionQuery
	AccountRepo           commondomain.IAccountRepository
	IdGenerator           commondomain.IIdGenerator
	PortGenerator         commondomain.IPortGenerator
	ManagerClient         commondomain.IManagerClient
	StorageManager        commondomain.IStorageManager
	PodManager            commondomain.IEnginePodManager
	ClassQuery            commondomain.IClassQuery
	ClusterManagerClient  commondomain.IClusterManagerClient
	ClusterManagerRemover commondomain.IClusterManagerRemover
	logger                logr.Logger
}

func (s *SharedStorageClusterService) GetAll() ([]*domain.SharedStorageCluster, error) {
	clusters, err := s.Repository.GetAll()
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		s.setAdapters(cluster)
	}
	return clusters, nil
}

func (s *SharedStorageClusterService) GetByName(name, namespace string) (*domain.SharedStorageCluster, error) {
	cluster, err := s.Repository.GetByName(name, namespace)
	if err != nil {
		return nil, err
	}
	s.setAdapters(cluster)
	return cluster, nil
}

func (s *SharedStorageClusterService) GetByData(data interface{}, useModifyClass bool, useUpgradeVersion bool) *domain.SharedStorageCluster {
	cluster := s.Repository.GetByData(data, useModifyClass, useUpgradeVersion)
	s.setAdapters(cluster)
	return cluster
}

func (s *SharedStorageClusterService) InitMeta(cluster *domain.SharedStorageCluster) error {
	if err := cluster.InitEngineParams(); err != nil {
		return err
	}
	if err := cluster.InitMeta(); err != nil {
		return err
	}
	if err := s.Repository.Update(cluster); err != nil {
		return err
	}
	if err := cluster.InitAccount(); err != nil {
		return err
	}
	return nil
}

func (s *SharedStorageClusterService) DeleteAllInsPod(cluster *domain.SharedStorageCluster, ctx context.Context) error {
	if err := cluster.RwIns.DeletePod(ctx); err != nil {
		return err
	}
	cluster.RwIns.EngineState = nil
	for roInsId, roIns := range cluster.RoInses {
		if err := roIns.DeletePod(ctx); err != nil {
			return err
		}
		if roIns.PhysicalInsId == cluster.RwIns.PhysicalInsId {
			delete(cluster.RoInses, roInsId)
		}
		roIns.EngineState = nil
	}
	for tempRoInsId, tempRoIns := range cluster.TempRoInses {
		if err := tempRoIns.DeletePod(ctx); err != nil {
			return err
		}
		delete(cluster.TempRoInses, tempRoInsId)
	}
	if err := s.Repository.UpdateInsStatus(cluster); err != nil {
		return err
	}
	if err := s.Repository.Update(cluster); err != nil {
		return err
	}
	return nil
}

func (s *SharedStorageClusterService) InitModifyClassMeta(cluster *domain.SharedStorageCluster) error {
	if err := cluster.ModifyClassParams(); err != nil {
		return err
	}
	if err := cluster.InitClass(); err != nil {
		return err
	}
	return s.Repository.Update(cluster)
}

func (s *SharedStorageClusterService) FlushParamsIfNecessary(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	if err := cluster.FlushParamsIfNecessary(ctx); err != nil {
		return err
	}
	return nil
}

func (s *SharedStorageClusterService) InitImages(cluster *domain.SharedStorageCluster) error {
	if err := cluster.InitVersion(); err != nil {
		return err
	}
	if err := s.Repository.Update(cluster); err != nil {
		return err
	}
	return nil
}

func (s *SharedStorageClusterService) ReleaseStorage(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	return cluster.ReleaseStorage(ctx)
}

func (s *SharedStorageClusterService) SaveParamsLastUpdateTime(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	return cluster.EngineParamsRepo.SaveLatestFlushTime(&cluster.DbClusterBase)
}

func (s *SharedStorageClusterService) PrepareStorage(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	return cluster.UseStorage(ctx, true)
}

func (s *SharedStorageClusterService) CreateRoIns(ctx context.Context, cluster *domain.SharedStorageCluster, ins *commondomain.DbIns) error {
	return s.CreateAndInstallIns(ctx, cluster, ins, false)
}

func (s *SharedStorageClusterService) CreateRwIns(ctx context.Context, cluster *domain.SharedStorageCluster, ins *commondomain.DbIns) error {
	err := s.CreateAndInstallIns(ctx, cluster, ins, true)
	if err != nil {
		return err
	}
	return ins.CreateAccount(ctx)
}

func (s *SharedStorageClusterService) GenerateNewRoTempId(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	ids, err := cluster.IdGenerator.GetNextClusterScopeHostInsIds(2)
	if err != nil {
		return err
	}
	physicalId := strconv.Itoa(ids[0])
	insId := strconv.Itoa(ids[1])
	cluster.TempRoIds = map[string]string{
		physicalId: insId,
	}
	return s.Repository.Update(cluster)
}

func (s *SharedStorageClusterService) GenerateTempRoId(ctx context.Context, cluster *domain.SharedStorageCluster, insId string) error {
	var physicalIds []string
	if cluster.RwIns.InsId == insId {
		physicalIds = append(physicalIds, cluster.RwIns.PhysicalInsId)
	} else {
		for roInsId, roIns := range cluster.RoInses {
			if roInsId == insId {
				physicalIds = append(physicalIds, roIns.PhysicalInsId)
				break
			}
		}
	}
	err := cluster.GenerateTempRoIds(ctx, physicalIds...)
	if err != nil {
		return err
	}
	return s.Repository.Update(cluster)
}

func (s *SharedStorageClusterService) GenerateTempRoIds(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	var physicalIds = []string{
		cluster.RwIns.PhysicalInsId,
	}
	for _, dbIns := range cluster.RoInses {
		physicalIds = append(physicalIds, dbIns.PhysicalInsId)
	}
	err := cluster.GenerateTempRoIds(ctx, physicalIds...)
	if err != nil {
		return err
	}
	return s.Repository.Update(cluster)
}

func (s *SharedStorageClusterService) InitTempRoMeta(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	err := cluster.InitTempRoInsMeta(ctx)
	if err != nil {
		return err
	}
	return s.Repository.Update(cluster)
}

func (s *SharedStorageClusterService) CreateTempRoIns(ctx context.Context, cluster *domain.SharedStorageCluster, ins *commondomain.DbIns) error {
	if err := s.CreateAndInstallIns(ctx, cluster, ins, false); err != nil {
		return err
	}
	return nil
}

func (s *SharedStorageClusterService) ConvertTempRoForRwToRo(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	if tempRoInsId := cluster.ConvertTempRoForRwToRo(); tempRoInsId != "" {
		if err := cluster.AddInsToClusterManager(ctx, "", tempRoInsId); err != nil {
			return err
		}
		return s.Repository.Update(cluster)
	}
	return nil
}

func (s *SharedStorageClusterService) SwitchNewRoToRw(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	newRwId, err := cluster.SwitchNewRoToRw(ctx)
	if err != nil {
		return err
	}
	return s.waitNewRwSwitchover(ctx, cluster, newRwId)
}

func (s *SharedStorageClusterService) Switchover(ctx context.Context, cluster *domain.SharedStorageCluster, newRwInsId string) error {
	err := cluster.Switchover(ctx, newRwInsId)
	if err != nil {
		return err
	}
	return s.waitNewRwSwitchover(ctx, cluster, newRwInsId)
}

func (s *SharedStorageClusterService) waitNewRwSwitchover(ctx context.Context, cluster *domain.SharedStorageCluster, newRwId string) error {
	err := waitutil.PollImmediateWithContext(ctx, 2*time.Second, 30*time.Second, func() (done bool, err error) {
		cluster, err = s.GetByName(cluster.Name, cluster.Namespace)
		if err != nil {
			return false, err
		}
		if cluster.RwIns.InsId == newRwId {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	cluster, err = s.GetByName(cluster.Name, cluster.Namespace)
	if err != nil {
		return err
	}
	if err := cluster.RwIns.DoHealthCheck(ctx); err != nil {
		return err
	}
	return nil
}

func (s *SharedStorageClusterService) DeleteOldRw(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	if err := cluster.DeleteOldIns(ctx, cluster.RwIns.PhysicalInsId, cluster.RwIns.InsId, true, false); err != nil {
		return err
	}

	return s.Repository.Update(cluster)
}

func (s *SharedStorageClusterService) DeleteOldIns(ctx context.Context, cluster *domain.SharedStorageCluster, phyId, newInsId string, deleteOldInsMeta, minusRoReplicas bool) error {
	if err := cluster.DeleteOldIns(ctx, phyId, newInsId, deleteOldInsMeta, minusRoReplicas); err != nil {
		return err
	}
	if deleteOldInsMeta {
		return s.Repository.Update(cluster)
	}
	return nil
}

func (s *SharedStorageClusterService) EnsureNewRoUpToDate(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	var tempRoIds []string
	for k, _ := range cluster.TempRoInses {
		tempRoIds = append(tempRoIds, k)
	}
	sort.Strings(tempRoIds)
	for _, tempRoInsId := range tempRoIds {
		tempRoIns := cluster.TempRoInses[tempRoInsId]
		if tempRoIns.PhysicalInsId == cluster.RwIns.PhysicalInsId {
			continue
		}
		if newTempInsId, ok := cluster.TempRoIds[tempRoIns.PhysicalInsId]; !ok || newTempInsId != tempRoInsId {
			continue
		}
		if err := s.CreateTempRoIns(ctx, cluster, tempRoIns); err != nil {
			return err
		}
		var deleteRo bool
		for roInsId, roIns := range cluster.RoInses {
			if tempRoIns.PhysicalInsId == roIns.PhysicalInsId {
				if tempRoInsId == roInsId {
					// already switched
					continue
				}
				deleteRo = true
				if err := cluster.RemoveInsFromClusterManager(ctx, roInsId); err != nil {
					return err
				}
				if err := roIns.StopEngineAndDeletePod(ctx, true); err != nil {
					return err
				}
				delete(cluster.RoInses, roInsId)
				cluster.RoInses[tempRoInsId] = tempRoIns
				delete(cluster.TempRoInses, tempRoInsId)
				if err := cluster.AddInsToClusterManager(ctx, "", tempRoInsId); err != nil {
					return err
				}
				if err := s.Repository.Update(cluster); err != nil {
					return err
				}
				break
			}
		}
		if !deleteRo {
			cluster.RoInses[tempRoInsId] = tempRoIns
			delete(cluster.TempRoInses, tempRoInsId)
			if err := cluster.AddInsToClusterManager(ctx, "", tempRoInsId); err != nil {
				return err
			}
			if err := s.Repository.Update(cluster); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *SharedStorageClusterService) SyncInsStateFromClusterManager(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	changed, err := cluster.SyncInsStateFromClusterManager(ctx)
	if err != nil {
		return err
	}
	if changed {
		if err := s.Repository.UpdateInsStatus(cluster); err != nil {
			return err
		}
		return cluster.EnsureInsTypeMeta(ctx)
	}
	return nil
}

func (s *SharedStorageClusterService) SetInsState(ctx context.Context, endpoint, state, startAt, reason string, cluster *domain.SharedStorageCluster) error {
	changed, err := cluster.SetInsState(ctx, endpoint, state, startAt, reason)
	if err != nil {
		return err
	}
	if changed {
		return s.Repository.UpdateInsStatus(cluster)
	}
	return nil
}

func (s *SharedStorageClusterService) SetRw(ctx context.Context, endpoint string, cluster *domain.SharedStorageCluster) error {
	changed, err := cluster.SetRw(ctx, endpoint)
	if err != nil {
		return err
	}
	if changed {
		if err := s.Repository.UpdateInsStatus(cluster); err != nil {
			return err
		}
		return cluster.EnsureInsTypeMeta(ctx)
	}
	return nil
}

func (s *SharedStorageClusterService) EnsureInsTypeMeta(ctx context.Context, cluster *domain.SharedStorageCluster) error {
	return cluster.EnsureInsTypeMeta(ctx)
}

func (s *SharedStorageClusterService) CreateAndInstallIns(ctx context.Context, cluster *domain.SharedStorageCluster, ins *commondomain.DbIns, writeLock bool) error {
	if !writeLock {
		if err := cluster.RwIns.CreateReplicationSlot(ctx, ins.ResourceName); err != nil {
			return err
		}
	}
	if err := ins.CreatePod(ctx); err != nil {
		return err
	}
	if err := s.Repository.Update(cluster); err != nil {
		return err
	}
	if writeLock {
		if err := cluster.SetStorageWriteLock(ctx, ins.Host); err != nil {
			return err
		}
	}
	if err := ins.InstallDbIns(ctx); err != nil {
		return err
	}
	if err := ins.DoHealthCheck(ctx); err != nil {
		return err
	}
	if err := ins.SetupLogAgent(ctx); err != nil {
		return err
	}
	if err := ins.SetDbInsInstalled(); err != nil {
		return err
	}
	if err := s.Repository.Update(cluster); err != nil {
		return err
	}
	return nil
}

func (s *SharedStorageClusterService) FlushClusterParams(ctx context.Context, sb *domain.SharedStorageCluster) (bool, error) {
	return sb.FlushClusterParams(ctx)
}

func (s *SharedStorageClusterService) UpdateRunningStatus(name, namespace string) error {
	return s.Repository.UpdateRunningStatus(name, namespace)
}

func (s *SharedStorageClusterService) EnableHA(ctx context.Context, sb *domain.SharedStorageCluster) error {
	return sb.EnableHA(ctx)
}

func (s *SharedStorageClusterService) DisableHA(ctx context.Context, sb *domain.SharedStorageCluster) error {
	return sb.DisableHA(ctx)
}

func (s *SharedStorageClusterService) EnsureCmAffinity(ctx context.Context, sb *domain.SharedStorageCluster) error {
	return sb.EnsureCmAffinity(ctx, sb.RwIns.Host)
}

func (s *SharedStorageClusterService) UpgradeCmVersion(ctx context.Context, sb *domain.SharedStorageCluster, cmImage string) error {
	return sb.UpgradeCmVersion(ctx, cmImage)
}

func (s *SharedStorageClusterService) RestartIns(ctx context.Context, sb *domain.SharedStorageCluster, insId string) error {
	return sb.RestartIns(ctx, insId)
}

func (s *SharedStorageClusterService) RestartCluster(ctx context.Context, sb *domain.SharedStorageCluster) error {
	return sb.RestartCluster(ctx)
}

func (s *SharedStorageClusterService) DeleteCm(ctx context.Context, sb *domain.SharedStorageCluster) error {
	return sb.DeleteCm(ctx)
}

func (s *SharedStorageClusterService) GrowStorage(ctx context.Context, sb *domain.SharedStorageCluster) error {
	return sb.GrowStorage(ctx)
}

func (s *SharedStorageClusterService) setAdapters(cluster *domain.SharedStorageCluster) {
	cluster.Init(
		s.ParamsTemplateQuery,
		s.ParamsClassQuery,
		s.ParamsRepo,
		s.MinorVersionQuery,
		s.AccountRepo,
		s.IdGenerator,
		s.PortGenerator,
		s.StorageManager,
		s.ClassQuery,
		s.ClusterManagerClient,
		s.ManagerClient,
		s.PodManager,
		s.ClusterManagerRemover,
		s.logger,
	)
}
