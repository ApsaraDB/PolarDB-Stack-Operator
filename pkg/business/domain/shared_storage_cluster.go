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

package domain

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/ApsaraDB/PolarDB-Stack-Common/utils/waitutil"

	commondomain "github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"
	commondefine "github.com/ApsaraDB/PolarDB-Stack-Common/define"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

type SharedStorageCluster struct {
	commondomain.SharedStorageDbClusterBase
	RoReplicas  int
	Port        int
	RwIns       *commondomain.DbIns
	RoInses     map[string]*commondomain.DbIns
	TempRoInses map[string]*commondomain.DbIns
	TempRoIds   map[string]string
	inited      bool
}

var sharedStorageClusterNotInitError = errors.New("shared storage cluster not init.")

func (cluster *SharedStorageCluster) Init(
	paramsTemplateQuery commondomain.IEngineParamsTemplateQuery,
	paramsClassQuery commondomain.IEngineParamsClassQuery,
	paramsRepo commondomain.IEngineParamsRepository,
	minorVersionQuery commondomain.IMinorVersionQuery,
	accountRepo commondomain.IAccountRepository,
	idGenerator commondomain.IIdGenerator,
	portGenerator commondomain.IPortGenerator,
	storageManager commondomain.IStorageManager,
	classQuery commondomain.IClassQuery,
	clusterManagerClient commondomain.IClusterManagerClient,
	managerClient commondomain.IManagerClient,
	podManager commondomain.IEnginePodManager,
	cmRemover commondomain.IClusterManagerRemover,
	logger logr.Logger,
) {
	cluster.SharedStorageDbClusterBase.Init(
		paramsTemplateQuery,
		paramsClassQuery,
		paramsRepo,
		minorVersionQuery,
		accountRepo,
		idGenerator,
		portGenerator,
		storageManager,
		classQuery,
		clusterManagerClient,
		nil,
		cmRemover,
		logger,
	)
	managerClient.Init(cluster)
	podManager.Init(cluster)
	cluster.Logger = logger
	if cluster.RwIns != nil {
		cluster.RwIns.Init(managerClient, podManager)
	}
	for _, roIns := range cluster.RoInses {
		roIns.Init(managerClient, podManager)
	}
	for _, tempRoIns := range cluster.TempRoInses {
		tempRoIns.Init(managerClient, podManager)
	}
	cluster.inited = true
}

func (cluster *SharedStorageCluster) InitMeta() (err error) {
	if !cluster.inited {
		return sharedStorageClusterNotInitError
	}
	if err = cluster.InitClass(); err != nil {
		return err
	}
	if err = cluster.InitVersion(); err != nil {
		return err
	}
	if err = cluster.InitDbInsMeta(); err != nil {
		return err
	}
	return nil
}

func (cluster *SharedStorageCluster) InitDbInsMeta() error {
	if !cluster.inited {
		return sharedStorageClusterNotInitError
	}
	if cluster.RwIns == nil || cluster.RwIns.InsId == "" {
		return cluster.initNewDbInsMeta()
	}
	return nil
}

func (cluster *SharedStorageCluster) AddInsToClusterManager(ctx context.Context, rwInsId string, roInsIds ...string) error {
	if !cluster.inited {
		return sharedStorageClusterNotInitError
	}
	if err := cluster.ClusterManagerClient.InitWithLocalDbCluster(ctx, cluster.Namespace, cluster.Name, true); err != nil {
		return err
	}
	if rwInsId != "" {
		if err := cluster.ClusterManagerClient.AddIns(ctx, cluster.RwIns.ResourceName, cluster.RwIns.HostIP, "RW", "SYNC", cluster.RwIns.Host, cluster.RwIns.NetInfo.Port, false); err != nil {
			return err
		}
		if err := cluster.WaitForEngineReady(ctx, cluster.RwIns.ResourceName); err != nil {
			return err
		}
	}
	for _, insId := range roInsIds {
		found := false
		for _, ins := range cluster.RoInses {
			if found {
				break
			}
			if insId == ins.InsId {
				found = true
				if err := cluster.ClusterManagerClient.AddIns(ctx, ins.ResourceName, ins.HostIP, "RO", "SYNC", ins.Host, ins.NetInfo.Port, false); err != nil {
					return err
				}
				if err := cluster.WaitForEngineReady(ctx, ins.ResourceName); err != nil {
					return err
				}
			}
		}
		for _, ins := range cluster.TempRoInses {
			if found {
				break
			}
			if insId == ins.InsId {
				found = true
				if err := cluster.ClusterManagerClient.AddIns(ctx, ins.ResourceName, ins.HostIP, "RO", "SYNC", ins.Host, ins.NetInfo.Port, false); err != nil {
					return err
				}
				if err := cluster.WaitForEngineReady(ctx, ins.ResourceName); err != nil {
					return err
				}
			}
		}
		if !found {
			return errors.New(fmt.Sprintf("insId %s not found", insId))
		}
	}
	return nil
}

func (cluster *SharedStorageCluster) WaitForEngineReady(ctx context.Context, resourceName string) error {
	if err := cluster.ClusterManagerClient.InitWithLocalDbCluster(ctx, cluster.Namespace, cluster.Name, true); err != nil {
		return err
	}
	return waitutil.PollImmediateWithContext(ctx, time.Second, 3*time.Minute, func() (bool, error) {
		if clusterStatus, err := cluster.ClusterManagerClient.GetClusterStatus(ctx); err == nil {
			if clusterStatus.Rw.PodName == resourceName {
				if clusterStatus.Rw.Phase == commondefine.EnginePhaseRunning {
					return true, nil
				} else if clusterStatus.Rw.Phase == commondefine.EnginePhaseFailed {
					return false, errors.New(fmt.Sprintf("%s rw start failed", resourceName))
				}
			} else {
				for _, ro := range clusterStatus.Ro {
					if ro.PodName == resourceName {
						if ro.Phase == commondefine.EnginePhaseRunning {
							return true, nil
						} else if clusterStatus.Rw.Phase == commondefine.EnginePhaseFailed {
							return false, errors.New(fmt.Sprintf("%s ro start failed", resourceName))
						}
					}
				}
			}
		}
		return false, nil
	})
}

func (cluster *SharedStorageCluster) RemoveInsFromClusterManager(ctx context.Context, insIds ...string) error {
	if !cluster.inited {
		return sharedStorageClusterNotInitError
	}
	if err := cluster.ClusterManagerClient.InitWithLocalDbCluster(ctx, cluster.Namespace, cluster.Name, true); err != nil {
		return err
	}
	for _, insId := range insIds {
		if insId == cluster.RwIns.InsId {
			if err := cluster.ClusterManagerClient.RemoveIns(ctx, cluster.RwIns.ResourceName, cluster.RwIns.HostIP, "RW", "SYNC", cluster.RwIns.Host, cluster.RwIns.NetInfo.Port, false); err != nil {
				return err
			}
		} else {
			found := false
			for _, ins := range cluster.RoInses {
				if found {
					break
				}
				if insId == ins.InsId {
					found = true
					if err := cluster.ClusterManagerClient.RemoveIns(ctx, ins.ResourceName, ins.HostIP, "RO", "SYNC", ins.Host, ins.NetInfo.Port, false); err != nil {
						return err
					}
				}
			}
			for _, ins := range cluster.TempRoInses {
				if found {
					break
				}
				if insId == ins.InsId {
					found = true
					if err := cluster.ClusterManagerClient.RemoveIns(ctx, ins.ResourceName, ins.HostIP, "RO", "SYNC", ins.Host, ins.NetInfo.Port, false); err != nil {
						return err
					}
				}
			}
			if !found {
				return errors.New(fmt.Sprintf("insId %s not found", insId))
			}
		}
	}
	return nil
}

func (cluster *SharedStorageCluster) GenerateTempRoIds(ctx context.Context, physicalInsIds ...string) error {
	if cluster.TempRoIds != nil && len(cluster.TempRoIds) > 0 {
		return nil
	}
	result := map[string]string{}
	if len(physicalInsIds) == 0 {
		physicalInsIds := append(physicalInsIds, cluster.RwIns.PhysicalInsId)
		for _, roIns := range cluster.RoInses {
			physicalInsIds = append(physicalInsIds, roIns.PhysicalInsId)
		}
	}
	insIds, err := cluster.IdGenerator.GetNextClusterScopeHostInsIds(len(physicalInsIds))
	if err != nil {
		return err
	}
	for i, physicalInsId := range physicalInsIds {
		result[physicalInsId] = strconv.Itoa(insIds[i])
	}
	cluster.TempRoIds = result
	return nil
}

func (cluster *SharedStorageCluster) InitTempRoInsMeta(ctx context.Context) error {
	if !cluster.inited {
		return sharedStorageClusterNotInitError
	}
	for physicalId, insId := range cluster.TempRoIds {
		if cluster.RwIns.InsId == insId {
			continue
		}
		if _, ok := cluster.RoInses[insId]; ok {
			continue
		}
		fsId, err := FsIdGenerator(cluster)
		if err != nil {
			return err
		}
		tempRoDbInsId := commondomain.DbInsId{
			PhysicalInsId: physicalId,
			InsId:         insId,
			InsName:       insId,
		}
		tempRoIns := &commondomain.DbIns{
			DbInsId:           tempRoDbInsId,
			ResourceName:      cluster.Name + "-" + tempRoDbInsId.PhysicalInsId + "-" + tempRoDbInsId.InsId,
			ResourceNamespace: cluster.Namespace,
			StorageHostId:     fsId,
			NetInfo: &commondomain.NetInfo{
				Port: cluster.Port,
			},
		}
		if cluster.TempRoInses == nil {
			cluster.TempRoInses = make(map[string]*commondomain.DbIns)
		}
		cluster.TempRoInses[insId] = tempRoIns
	}
	return nil
}

func (cluster *SharedStorageCluster) ConvertTempRoForRwToRo() string {
	for tempRoInsId, tempRoIns := range cluster.TempRoInses {
		if tempRoIns.PhysicalInsId == cluster.RwIns.PhysicalInsId {
			if tempRoInsId == cluster.RwIns.InsId {
				continue
			}
		}
		if _, ok := cluster.RoInses[tempRoInsId]; !ok {
			cluster.RoInses[tempRoInsId] = tempRoIns
			delete(cluster.TempRoInses, tempRoInsId)
			return tempRoInsId
		}
	}
	return ""
}

func (cluster *SharedStorageCluster) SwitchNewRoToRw(ctx context.Context) (newRwId string, err error) {
	newRwId = cluster.TempRoIds[cluster.RwIns.PhysicalInsId]
	if newRwId == "" {
		return "", commondefine.CreateInterruptError(define.TempRoInsIdNotFound, nil)
	}

	return newRwId, cluster.Switchover(ctx, newRwId)
}

func (cluster *SharedStorageCluster) Switchover(ctx context.Context, newRwInsId string) error {
	if err := cluster.ClusterManagerClient.InitWithLocalDbCluster(ctx, cluster.Namespace, cluster.Name, true); err != nil {
		return err
	}
	if roIns, ok := cluster.RoInses[newRwInsId]; ok {
		return cluster.ClusterManagerClient.Switchover(ctx, cluster.RwIns.HostIP, roIns.HostIP, strconv.Itoa(cluster.Port), false)
	}
	return nil
}

func (cluster *SharedStorageCluster) DeleteOldIns(ctx context.Context, physicalId string, newInsId string, deleteOldInsMeta, minusRoReplicas bool) error {
	for roInsId, roIns := range cluster.RoInses {
		if roIns.PhysicalInsId == physicalId {
			if roInsId == newInsId {
				return commondefine.CreateInterruptError(define.RoMetaInvalid, nil)
			}
			if err := cluster.RemoveInsFromClusterManager(ctx, roInsId); err != nil {
				return err
			}
			if err := roIns.StopEngineAndDeletePod(ctx, true); err != nil {
				return err
			}
			if deleteOldInsMeta {
				if minusRoReplicas {
					cluster.RoReplicas--
				}
				delete(cluster.RoInses, roInsId)
			}
		}
	}
	return nil
}

func (cluster *SharedStorageCluster) SyncInsStateFromClusterManager(ctx context.Context) (bool, error) {
	if !cluster.inited {
		return false, sharedStorageClusterNotInitError
	}
	if err := cluster.ClusterManagerClient.InitWithLocalDbCluster(ctx, cluster.Namespace, cluster.Name, false); err != nil {
		return false, err
	}

	status, err := cluster.ClusterManagerClient.GetClusterStatus(context.TODO())
	if err != nil {
		cluster.Logger.Error(err, "cluster.ClusterManagerClient.GetClusterStatus error")
		return false, err
	}
	rwChanged, err := cluster.checkRwSwitched(status.Rw.Endpoint)
	if err != nil {
		return false, err
	}
	statusChanged, err := cluster.checkStatusChanged(status)
	if err != nil {
		return false, err
	}
	return rwChanged || statusChanged, nil
}

func (cluster *SharedStorageCluster) SetInsState(ctx context.Context, clientEndpoint, state, startAt, reason string) (bool, error) {
	if !cluster.inited {
		return false, sharedStorageClusterNotInitError
	}
	return cluster.checkInsStateChanged(clientEndpoint, state, startAt, reason)
}

func (cluster *SharedStorageCluster) SetRw(ctx context.Context, newRwEndpoint string) (bool, error) {
	if !cluster.inited {
		return false, sharedStorageClusterNotInitError
	}
	return cluster.checkRwSwitched(newRwEndpoint)
}

func (cluster *SharedStorageCluster) EnsureInsTypeMeta(ctx context.Context) error {
	if !cluster.inited {
		return sharedStorageClusterNotInitError
	}
	if err := cluster.RwIns.EnsureInsTypeMeta(ctx, define.InsTypeRw); err != nil {
		return err
	}
	for _, ins := range cluster.RoInses {
		if err := ins.EnsureInsTypeMeta(ctx, define.InsTypeRo); err != nil {
			return err
		}
	}
	return nil
}

func (cluster *SharedStorageCluster) FlushParamsIfNecessary(ctx context.Context) error {
	if !cluster.inited {
		return sharedStorageClusterNotInitError
	}
	userParams, _, err := cluster.GetUserParams()
	if err != nil {
		return err
	}
	runningParams, err := cluster.GetRunningParams()
	if err != nil {
		return err
	}
	oldMaxConnParam, err := GetParamValue(runningParams, "max_connections")
	if err != nil {
		return err
	}
	newMaxConnParam, err := GetParamValue(userParams, "max_connections")
	if err != nil {
		return nil
	}

	oldIntMaxConn, err := strconv.Atoi(oldMaxConnParam.Value)
	if err != nil {
		cluster.Logger.Error(err, "max_connections param format illegal.", "oldIntMaxConn", oldIntMaxConn)
		return err
	}
	newIntMaxConn, err := strconv.Atoi(newMaxConnParam.Value)
	if err != nil {
		cluster.Logger.Error(err, "max_connections param format illegal.", "newIntMaxConn", newIntMaxConn)
		return err
	}

	oldIntMaxTrans := 0
	oldMaxTransParam, err := GetParamValue(runningParams, "max_prepared_transactions")
	if err == nil {
		oldIntMaxTrans, err = strconv.Atoi(oldMaxTransParam.Value)
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("running params [max_prepared_transactions] value illegal, %v", oldMaxTransParam.Value))
			cluster.Logger.Error(err, "")
			return err
		}
	}

	canUpdateNotChangeableParams := commondomain.GetCanUpdateNotChangeableParams(cluster.Logger)
	if oldIntMaxConn > newIntMaxConn || oldIntMaxTrans > 800 {
		rwParams, _, _, err := cluster.GetEffectiveParams(false, canUpdateNotChangeableParams)
		if err != nil {
			return err
		}
		if oldIntMaxConn > newIntMaxConn {
			// ro连接数必须 >= rw，否则ro会crash，所以降配要把rw先降下来，保证创建的新temp ro=rw，ro > rw
			rwParams["max_connections"] = newMaxConnParam.Value
		}
		if oldIntMaxTrans > 800 {
			// 参数模板中此项不可更改，固定值是800
			rwParams["max_prepared_transactions"] = "800"
		}
		if err = cluster.RwIns.FlushParams(ctx, rwParams, true); err != nil {
			return errors.Wrap(err, "flush rw params error.")
		}
		if err = cluster.RwIns.DoHealthCheck(ctx); err != nil {
			return errors.Wrap(err, "wait rw health error.")
		}
	}
	if oldIntMaxConn < newIntMaxConn {
		// 升配要先把ro的连接数升上去，防止ro crash
		params, _, _, err := cluster.GetEffectiveParams(false, canUpdateNotChangeableParams)
		if err != nil {
			return err
		}
		params["max_connections"] = newMaxConnParam.Value
		for _, roIns := range cluster.RoInses {
			if err = roIns.FlushParams(ctx, params, true); err != nil {
				return errors.Wrap(err, fmt.Sprintf("flush ro %s params error.", roIns.ResourceName))
			}
			if err = roIns.DoHealthCheck(ctx); err != nil {
				return errors.Wrap(err, fmt.Sprintf("wait ro %s health error.", roIns.ResourceName))
			}
		}
	}

	return nil
}

func (cluster *SharedStorageCluster) RestartIns(ctx context.Context, insId string) error {
	if !cluster.inited {
		return sharedStorageClusterNotInitError
	}
	if cluster.RwIns.InsId == insId {
		return cluster.RwIns.Restart(ctx)
	}
	for roId, ins := range cluster.RoInses {
		if roId == insId {
			return ins.Restart(ctx)
		}
	}
	return nil
}

func (cluster *SharedStorageCluster) RestartCluster(ctx context.Context) error {
	if !cluster.inited {
		return sharedStorageClusterNotInitError
	}
	for _, ins := range cluster.RoInses {
		if err := ins.GracefulStopEngine(ctx); err != nil {
			return err
		}
	}
	if err := cluster.RwIns.Restart(ctx); err != nil {
		return err
	}
	for _, ins := range cluster.RoInses {
		if err := ins.StartEngine(ctx); err != nil {
			return err
		}
	}
	for _, ins := range cluster.RoInses {
		if err := ins.DoHealthCheck(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (cluster *SharedStorageCluster) GrowStorage(ctx context.Context) error {
	if !cluster.inited {
		return sharedStorageClusterNotInitError
	}
	for _, ins := range cluster.RoInses {
		if err := ins.GrowStorage(ctx); err != nil {
			return err
		}
	}
	if err := cluster.RwIns.GrowStorage(ctx); err != nil {
		return err
	}
	return nil
}

/**
 * @Description: 执行刷参
 * @receiver ins
 * @return error
 */
func (cluster *SharedStorageCluster) FlushClusterParams(ctx context.Context) (needRestart bool, err error) {
	if !cluster.inited {
		return false, sharedStorageClusterNotInitError
	}
	if err := cluster.ModifyClassParams(); err != nil {
		return false, err
	}
	canUpdateNotChangeableParams := commondomain.GetCanUpdateNotChangeableParams(cluster.Logger)
	params, changed, needRestart, err := cluster.GetEffectiveParams(false, canUpdateNotChangeableParams)
	if err != nil {
		return false, err
	}
	if !changed {
		cluster.Logger.Info("params is not change, ignore")
		return false, nil
	}
	if err = cluster.RwIns.FlushParams(ctx, params, false); err != nil {
		return false, err
	}

	for _, roIns := range cluster.RoInses {
		if err = roIns.FlushParams(ctx, params, false); err != nil {
			return false, err
		}
	}
	return needRestart, nil
}

func GetParamValue(paramItems map[string]*commondomain.ParamItem, paramName string) (*commondomain.ParamItem, error) {
	value, ok := paramItems[paramName]
	if !ok || value == nil {
		return nil, errors.New(fmt.Sprintf("GetParamValue, [%s] not found", paramName))
	}
	return value, nil
}

func (cluster *SharedStorageCluster) checkStatusChanged(status *commondomain.ClusterStatus) (bool, error) {
	podStatus := status.Ro
	podStatus = append(podStatus, status.Rw)
	changed := false
	for _, status := range podStatus {
		c, err := cluster.checkInsStateChanged(status.Endpoint, status.Phase, status.StartAt, "")
		if err != nil {
			return false, err
		}
		if c {
			changed = true
		}
	}
	return changed, nil
}

func (cluster *SharedStorageCluster) checkInsStateChanged(clientEndpoint, state, startAt, reason string) (bool, error) {
	changed := false
	found := false
	if cluster.RwIns != nil && cluster.RwIns.ClientIP+":"+strconv.Itoa(cluster.Port) == clientEndpoint {
		found = true
		if cluster.RwIns.EngineState.CurrentState.State != state {
			cluster.RwIns.EngineState.LastState = cluster.RwIns.EngineState.CurrentState
			cluster.RwIns.EngineState.CurrentState = commondomain.EngineState{
				StartedAt: parseTime(startAt),
				State:     state,
				Reason:    reason,
			}
			changed = true
			cluster.Logger.Info("rw ins status changed", "insId", cluster.RwIns.InsId, "newState", state)
		}
	}
	for _, roIns := range cluster.RoInses {
		if roIns.ClientIP+":"+strconv.Itoa(cluster.Port) == clientEndpoint {
			found = true
			if roIns.EngineState.CurrentState.State != state {
				roIns.EngineState.LastState = roIns.EngineState.CurrentState
				roIns.EngineState.CurrentState = commondomain.EngineState{
					StartedAt: parseTime(startAt),
					State:     state,
					Reason:    reason,
				}
				changed = true
				cluster.Logger.Info("ro ins status changed", "insId", roIns.InsId, "newState", state)
			}
		}
	}
	if !found {
		return false, errors.New(fmt.Sprintf("pod %s is not found from cluster metadata.", clientEndpoint))
	}
	return changed, nil
}

func parseTime(t string) *time.Time {
	result := time.Now()
	if t != "" {
		result, _ = time.Parse("2006-01-02 15:04:05", t)
	}
	return &result
}

func (cluster *SharedStorageCluster) checkRwSwitched(newRwEndpoint string) (bool, error) {
	if cluster.RwIns != nil && newRwEndpoint != "" && newRwEndpoint != cluster.RwIns.ClientIP+":"+strconv.Itoa(cluster.Port) {
		newRwInsId := cluster.RwIns.InsId
		found := false
		for insId, ins := range cluster.RoInses {
			if ins.ClientIP+":"+strconv.Itoa(cluster.Port) == newRwEndpoint {
				newRwInsId = insId
				found = true
				break
			}
		}
		if !found {
			return false, errors.New(fmt.Sprintf("new rw pod %s is not found from cluster metadata.", newRwEndpoint))
		}
		if newRwInsId != cluster.RwIns.InsId {
			oldRwIns := cluster.RwIns
			cluster.RwIns = cluster.RoInses[newRwInsId]
			delete(cluster.RoInses, newRwInsId)
			cluster.RoInses[oldRwIns.InsId] = oldRwIns
			cluster.Logger.Info("rw ins switched", "oldRw", oldRwIns, "newRw", cluster.RwIns)
			return true, nil
		}
	}
	return false, nil
}

func (cluster *SharedStorageCluster) initNewDbInsMeta() error {
	numIdNeedGenerate := (cluster.RoReplicas+1)*2 + 1
	ids, err := cluster.IdGenerator.GetNextClusterScopeHostInsIds(numIdNeedGenerate)
	if err != nil {
		return err
	}
	isUsed, err := cluster.PortGenerator.CheckPortUsed(cluster.Port, define.EnginePortRangesAnnotation)
	if isUsed {
		return err
	}

	cluster.LogicInsId = strconv.Itoa(ids[0])
	rwDbInsId := commondomain.DbInsId{
		PhysicalInsId: strconv.Itoa(ids[1]),
		InsId:         strconv.Itoa(ids[2]),
		InsName:       strconv.Itoa(ids[2]),
	}
	fsId, err := FsIdGenerator(cluster)
	if err != nil {
		return err
	}
	cluster.RwIns = &commondomain.DbIns{
		DbInsId:           rwDbInsId,
		ResourceName:      cluster.Name + "-" + rwDbInsId.PhysicalInsId + "-" + rwDbInsId.InsId,
		ResourceNamespace: cluster.Namespace,
		StorageHostId:     fsId,
		NetInfo: &commondomain.NetInfo{
			Port: cluster.Port,
		},
	}

	for i := 3; i < numIdNeedGenerate; i = i + 2 {
		roPhysicalInsId := strconv.Itoa(ids[i])
		roInsId := strconv.Itoa(ids[i+1])
		roFsId, err := FsIdGenerator(cluster)
		if err != nil {
			return err
		}
		roDbInsId := commondomain.DbInsId{
			PhysicalInsId: roPhysicalInsId,
			InsId:         roInsId,
			InsName:       roInsId,
		}
		roIns := &commondomain.DbIns{
			DbInsId:           roDbInsId,
			ResourceName:      cluster.Name + "-" + roDbInsId.PhysicalInsId + "-" + roDbInsId.InsId,
			ResourceNamespace: cluster.Namespace,
			StorageHostId:     roFsId,
			NetInfo: &commondomain.NetInfo{
				Port: cluster.Port,
			},
		}
		cluster.RoInses[roInsId] = roIns
	}

	if cluster.ClusterManager == nil {
		cluster.ClusterManager = &commondomain.ClusterManagerInfo{}
	}
	port, err := cluster.PortGenerator.GetNextClusterExternalPort()
	if err != nil {
		return err
	}
	cluster.ClusterManager.Port = port
	cluster.ClusterStatus = string(statemachine.StateCreating)
	return nil
}
