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

package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/ApsaraDB/PolarDB-Stack-Common/utils/k8sutil"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	commondomain "github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"
	mgr "github.com/ApsaraDB/PolarDB-Stack-Common/manager"
	commonutils "github.com/ApsaraDB/PolarDB-Stack-Common/utils"
	v1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/domain"
	mpddefine "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/define"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var oldResourceVersionError = errors.New("domain model resource version is too old.")

func NewSharedStorageClusterRepository(logger logr.Logger) *SharedStorageClusterRepository {
	return &SharedStorageClusterRepository{
		logger: logger,
	}
}

type SharedStorageClusterRepository struct {
	logger logr.Logger
}

func (r *SharedStorageClusterRepository) GetAll() ([]*domain.SharedStorageCluster, error) {
	var result []*domain.SharedStorageCluster
	list, err := r.getAllKubeResource()
	if err != nil {
		return nil, err
	}
	for _, item := range list.Items {
		result = append(result, r.convertToDomainModel(&item, false, false))
	}
	return result, nil
}

func (r *SharedStorageClusterRepository) GetByName(name, namespace string) (*domain.SharedStorageCluster, error) {
	item, err := getKubeResource(name, namespace)
	if err != nil {
		return nil, err
	}
	return r.GetByData(item, false, false), nil
}

func (r *SharedStorageClusterRepository) GetByData(data interface{}, useModifyClass bool, useUpgradeVersion bool) *domain.SharedStorageCluster {
	obj := data.(*v1.MPDCluster)
	return r.convertToDomainModel(obj, useModifyClass, useUpgradeVersion)
}

func (r *SharedStorageClusterRepository) Create(cluster *domain.SharedStorageCluster) error {
	kubeRes := &v1.MPDCluster{}
	kubeRes.Namespace = cluster.Namespace
	kubeRes.Name = cluster.Name
	kubeRes.Annotations = map[string]string{}
	kubeRes.Spec = v1.MPDClusterSpec{}
	return mgr.GetSyncClient().Create(context.TODO(), kubeRes)
}

func (r *SharedStorageClusterRepository) UpdateWithResourceVersion(cluster *domain.SharedStorageCluster) error {
	kubeObj, err := getKubeResource(cluster.Name, cluster.Namespace)
	if err != nil {
		return err
	}
	if kubeObj.ResourceVersion != cluster.ResourceVersion {
		return oldResourceVersionError
	}
	return r.update(cluster, kubeObj)
}

func (r *SharedStorageClusterRepository) Update(cluster *domain.SharedStorageCluster) error {
	kubeRes, err := getKubeResource(cluster.Name, cluster.Namespace)
	if err != nil {
		return err
	}
	return r.update(cluster, kubeRes)
}

func (r *SharedStorageClusterRepository) UpdateRunningStatus(name, namespace string) error {
	kubeRes, err := getKubeResource(name, namespace)
	if err != nil {
		return err
	}
	kubeRes.Status.ClusterStatus = statemachine.StateRunning
	return updateKubeMpdClusterStatus(kubeRes)
}

func (r *SharedStorageClusterRepository) UpdateInsStatus(cluster *domain.SharedStorageCluster) error {
	kubeRes, err := getKubeResource(cluster.Name, cluster.Namespace)
	if err != nil {
		return err
	}

	if cluster.RwIns != nil && cluster.RwIns.InsId != "" {
		updateEngineState(cluster.RwIns, kubeRes, v1.MPDClusterInstanceTypeRW)
	}

	if cluster.RoInses != nil && len(cluster.RoInses) > 0 {
		for _, ins := range cluster.RoInses {
			updateEngineState(ins, kubeRes, v1.MPDClusterInstanceTypeRO)
		}
	}

	return updateKubeMpdClusterStatus(kubeRes)
}

func updateEngineState(ins *commondomain.DbIns, kubeRes *v1.MPDCluster, insType v1.MPDClusterInstanceType) {
	if insType == v1.MPDClusterInstanceTypeRW {
		kubeRes.Status.LeaderInstanceId = ins.InsId
		kubeRes.Status.LeaderInstanceHost = ins.Host
	}

	currentStartAt := metav1.Now()
	if ins.EngineState != nil && ins.EngineState.CurrentState.StartedAt != nil {
		currentStartAt = metav1.NewTime(*ins.EngineState.CurrentState.StartedAt)
	}

	var lastStartAt *metav1.Time
	if ins.EngineState != nil && ins.EngineState.LastState.StartedAt != nil {
		startAt := metav1.NewTime(*ins.EngineState.LastState.StartedAt)
		lastStartAt = &startAt
	}

	if insMeta, ok := kubeRes.Status.DBInstanceStatus[ins.InsId]; ok {
		insMeta.InsType = insType
		if ins.EngineState != nil {
			insMeta.CurrentState = v1.MPDClusterInstanceState{
				Reason:    ins.EngineState.CurrentState.Reason,
				State:     ins.EngineState.CurrentState.State,
				StartedAt: &currentStartAt,
			}
			insMeta.LastState = v1.MPDClusterInstanceState{
				Reason:    ins.EngineState.LastState.Reason,
				State:     ins.EngineState.LastState.State,
				StartedAt: lastStartAt,
			}
		} else {
			insMeta.CurrentState = v1.MPDClusterInstanceState{}
			insMeta.LastState = v1.MPDClusterInstanceState{}
		}
	}
}

func (*SharedStorageClusterRepository) getAllKubeResource() (*v1.MPDClusterList, error) {
	list := &v1.MPDClusterList{}
	err := mgr.GetSyncClient().List(context.TODO(), list, &client.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (*SharedStorageClusterRepository) updateKubeMpdClusterSpec(cluster *v1.MPDCluster) error {
	return mgr.GetSyncClient().Update(context.TODO(), cluster)
}

func (r *SharedStorageClusterRepository) update(model *domain.SharedStorageCluster, kubeRes *v1.MPDCluster) error {
	var err error
	oldKubeRes := kubeRes.DeepCopy()
	kubeRes.Spec.ResourceAdditional = convertResourceToKube(model.Resources)
	if model.UseUpgradeImageInfo {
		kubeRes.Spec.VersionCfgModifyTo = *r.convertImageInfoToKube(model.ImageInfo)
	} else {
		kubeRes.Spec.VersionCfg = *r.convertImageInfoToKube(model.ImageInfo)
	}
	if model.TempRoIds != nil && len(model.TempRoIds) > 0 {
		if byteTempRoIds, err := json.Marshal(model.TempRoIds); err != nil {
			r.logger.Error(err, "json.Marshal failed", "oriValue", model.TempRoIds)
		} else {
			kubeRes.Annotations[mpddefine.AnnotationTempRoIds] = string(byteTempRoIds)
		}
	} else {
		delete(kubeRes.Annotations, mpddefine.AnnotationTempRoIds)
	}
	if model.RoReplicas != oldKubeRes.Spec.FollowerNum {
		kubeRes.Spec.FollowerNum = model.RoReplicas
	}
	if !reflect.DeepEqual(kubeRes.Spec, oldKubeRes.Spec) || !reflect.DeepEqual(kubeRes.Annotations, oldKubeRes.Annotations) {
		if err := r.updateKubeMpdClusterSpec(kubeRes); err != nil {
			return err
		}
		if kubeRes, err = getKubeResource(kubeRes.Name, kubeRes.Namespace); err != nil {
			return err
		}
		oldKubeRes = kubeRes.DeepCopy()
	}

	kubeRes.Status.LogicInsId = model.LogicInsId
	kubeRes.Status.DBInstanceStatus = map[string]*v1.MPDClusterInstanceStatus{}

	if model.ClusterStatus != "" {
		kubeRes.Status.ClusterStatus = statemachine.State(model.ClusterStatus)
	}
	if model.ClusterManager != nil {
		kubeRes.Status.ClusterManagerStatus = v1.MPDClusterManagerStatus{
			WorkingPort: model.ClusterManager.Port,
		}
	}
	kubeRes.Status.DBInstanceStatus = map[string]*v1.MPDClusterInstanceStatus{}
	var allInsId []string
	if model.RwIns != nil && model.RwIns.InsId != "" {
		allInsId = append(allInsId, model.RwIns.InsId)
		kubeRes.Status.LeaderInstanceId = model.RwIns.InsId
		insStatus := r.convertToKubeInsStatus(model.RwIns, v1.MPDClusterInstanceTypeRW, oldKubeRes)
		kubeRes.Status.DBInstanceStatus[model.RwIns.InsId] = insStatus
	}
	for insId, ins := range model.RoInses {
		allInsId = append(allInsId, insId)
		insStatus := r.convertToKubeInsStatus(ins, v1.MPDClusterInstanceTypeRO, oldKubeRes)
		kubeRes.Status.DBInstanceStatus[insId] = insStatus
	}
	for insId, ins := range model.TempRoInses {
		allInsId = append(allInsId, insId)
		insStatus := r.convertToKubeInsStatus(ins, v1.MPDClusterInstanceTypeTempRO, oldKubeRes)
		kubeRes.Status.DBInstanceStatus[insId] = insStatus
	}
	for insId, _ := range kubeRes.Status.DBInstanceStatus {
		if !commonutils.ContainsString(allInsId, insId, nil) {
			delete(kubeRes.Status.DBInstanceStatus, insId)
		}
	}

	if !reflect.DeepEqual(kubeRes.Status, oldKubeRes.Status) {
		return updateKubeMpdClusterStatus(kubeRes)
	}
	return nil

}

func (r *SharedStorageClusterRepository) convertToKubeInsStatus(ins *commondomain.DbIns, insType v1.MPDClusterInstanceType, oldKubeRes *v1.MPDCluster) *v1.MPDClusterInstanceStatus {
	clientIP, err := k8sutil.GetClientIP(ins.HostIP, r.logger)
	if err != nil {
		r.logger.Error(err, "")
	}
	insStatus := &v1.MPDClusterInstanceStatus{
		PhysicalInsId: ins.PhysicalInsId,
		InsId:         ins.InsId,
		InsName:       ins.InsName,
		PodName:       ins.ResourceName,
		PodNameSpace:  ins.ResourceNamespace,
		NodeName:      ins.Host,
		HostClientIP:  clientIP,
		PolarFsHostId: ins.StorageHostId,
		Installed:     ins.Installed,
		Role:          "",
		InsType:       insType,
		Status:        "",
		VersionInfo:   v1.VersionInfo{},
		InsClassInfo:  v1.InstanceClassInfo{},
		NetInfo:       v1.DBInstanceNetInfo{},
		CurrentState:  v1.MPDClusterInstanceState{},
		LastState:     v1.MPDClusterInstanceState{},
	}
	if oldInsStatus, ok := oldKubeRes.Status.DBInstanceStatus[ins.InsId]; ok {
		// 状态只能cm更新
		insStatus.CurrentState = oldInsStatus.CurrentState
		insStatus.LastState = oldInsStatus.LastState
	}
	if ins.NetInfo != nil {
		port := ins.NetInfo.Port
		insStatus.NetInfo = v1.DBInstanceNetInfo{
			NetType:              "",
			WorkingPort:          port,
			WorkingHostIP:        ins.HostIP,
			EnableWorkingAdminIP: false,
			WorkingAdminIP:       "",
		}
	}
	return insStatus
}

func (r *SharedStorageClusterRepository) convertToDomainModel(k8sObj *v1.MPDCluster, useModifyClass bool, useUpgradeVersion bool) *domain.SharedStorageCluster {
	result := &domain.SharedStorageCluster{
		RoReplicas: k8sObj.Spec.FollowerNum,
		Port:       k8sObj.Spec.NetCfg.EngineStartPort,
		SharedStorageDbClusterBase: commondomain.SharedStorageDbClusterBase{
			DbClusterBase: commondomain.DbClusterBase{
				Name:                k8sObj.Name,
				Namespace:           k8sObj.Namespace,
				Description:         k8sObj.Spec.Description,
				ClusterStatus:       string(k8sObj.Status.ClusterStatus),
				LogicInsId:          k8sObj.Status.LogicInsId,
				ImageInfo:           r.convertImageInfoToDomainModel(&k8sObj.Spec.VersionCfg),
				ClassInfo:           r.convertClassInfoToDomainModel(&k8sObj.Spec.ClassInfo),
				Resources:           r.convertResourceToDomainModel(k8sObj.Spec.ResourceAdditional),
				DbClusterType:       commondomain.DbClusterTypeMaster,
				EngineType:          commondomain.EngineTypeRwo,
				Interrupt:           false,
				InterruptMsg:        "",
				InterruptReason:     "",
				ResourceVersion:     k8sObj.ResourceVersion,
				UseModifyClass:      useModifyClass,
				UseUpgradeImageInfo: useUpgradeVersion,
			},
			StorageInfo: r.convertStorageInfoToDomainModel(k8sObj.Spec.ShareStore),
		},
		RoInses:     map[string]*commondomain.DbIns{},
		TempRoInses: map[string]*commondomain.DbIns{},
	}

	var strTempRoIds string
	if k8sObj.Annotations != nil {
		strTempRoIds = k8sObj.Annotations[mpddefine.AnnotationTempRoIds]
	}
	var tempRoIds = &map[string]string{}
	if strTempRoIds != "" {
		if err := json.Unmarshal([]byte(strTempRoIds), tempRoIds); err != nil {
			r.logger.Error(err, fmt.Sprintf("annotation [%s] is invaild json", mpddefine.AnnotationTempRoIds))
		}
		result.TempRoIds = *tempRoIds
	}

	if useModifyClass && k8sObj.Spec.ClassInfoModifyTo.ClassName != "" {
		result.ClassInfo = r.convertClassInfoToDomainModel(&k8sObj.Spec.ClassInfoModifyTo)
		result.OldClassInfo = r.convertClassInfoToDomainModel(&k8sObj.Spec.ClassInfo)
	}
	if useUpgradeVersion && k8sObj.Spec.VersionCfgModifyTo.VersionName != "" {
		result.ImageInfo = r.convertImageInfoToDomainModel(&k8sObj.Spec.VersionCfgModifyTo)
		result.OldImageInfo = r.convertImageInfoToDomainModel(&k8sObj.Spec.VersionCfg)
	}

	if k8sObj.Status.DBInstanceStatus != nil && len(k8sObj.Status.DBInstanceStatus) > 0 {
		for _, ins := range k8sObj.Status.DBInstanceStatus {
			insModel := &commondomain.DbIns{}
			insModel.DbInsId = commondomain.DbInsId{
				PhysicalInsId: ins.PhysicalInsId,
				InsId:         ins.InsId,
				InsName:       ins.InsName,
			}
			insModel.Host = ins.NodeName
			insModel.HostIP = ins.NetInfo.WorkingHostIP
			insModel.ClientIP = ins.HostClientIP
			insModel.ResourceName = ins.PodName
			insModel.ResourceNamespace = k8sObj.Namespace
			insModel.StorageHostId = ins.PolarFsHostId
			insModel.Installed = ins.Installed
			netInfo := r.convertNetInfoToDomainModel(&ins.NetInfo)
			if netInfo != nil {
				insModel.NetInfo = netInfo
			}
			insModel.EngineState = r.convertStateToDomainModel(&ins.CurrentState, &ins.LastState)
			if ins.InsType == v1.MPDClusterInstanceTypeRW {
				result.RwIns = insModel
			} else if ins.InsType == v1.MPDClusterInstanceTypeRO {
				result.RoInses[ins.InsId] = insModel
			} else if ins.InsType == v1.MPDClusterInstanceTypeTempRO {
				result.TempRoInses[ins.InsId] = insModel
			}
		}
	}
	result.ClusterManager = r.convertCmToDomainModel(&k8sObj.Status.ClusterManagerStatus)

	if result.ClusterStatus == string(statemachine.StateInterrupt) {
		result.Interrupt = true
		result.InterruptMsg = k8sObj.Annotations[mpddefine.DefaultWfConf[define.WFInterruptMessage]]
		result.InterruptReason = k8sObj.Annotations[mpddefine.DefaultWfConf[define.WFInterruptReason]]
	}
	return result
}

func (r *SharedStorageClusterRepository) convertCmToDomainModel(info *v1.MPDClusterManagerStatus) *commondomain.ClusterManagerInfo {
	if info == nil || info.WorkingPort == 0 {
		return nil
	}
	return &commondomain.ClusterManagerInfo{
		Port: info.WorkingPort,
	}
}

func (r *SharedStorageClusterRepository) convertStateToDomainModel(currentState, lastState *v1.MPDClusterInstanceState) *commondomain.EngineStatus {
	result := &commondomain.EngineStatus{
		CurrentState: commondomain.EngineState{
			Reason: currentState.Reason,
			State:  currentState.State,
		},
		LastState: commondomain.EngineState{
			Reason: lastState.Reason,
			State:  lastState.State,
		},
	}
	if currentState.StartedAt != nil {
		result.CurrentState.StartedAt = &currentState.StartedAt.Time
	}
	if lastState.StartedAt != nil {
		result.LastState.StartedAt = &lastState.StartedAt.Time
	}
	return result
}

func (r *SharedStorageClusterRepository) convertNetInfoToDomainModel(netInfo *v1.DBInstanceNetInfo) *commondomain.NetInfo {
	var result = &commondomain.NetInfo{
		// FloatingIP: netInfo.Vip,
		Port: netInfo.WorkingPort,
	}

	return result
}

func (r *SharedStorageClusterRepository) convertStorageInfoToDomainModel(info *v1.ShareStoreConfig) *commondomain.StorageInfo {
	return &commondomain.StorageInfo{
		DiskID:     info.SharePvcName,
		DiskQuota:  info.DiskQuota,
		VolumeId:   info.VolumeId,
		VolumeType: info.VolumeType,
	}
}

func (r *SharedStorageClusterRepository) convertResourceToDomainModel(info map[string]v1.AdditionalResourceCfg) map[string]*commondomain.InstanceResource {
	var result = make(map[string]*commondomain.InstanceResource)
	for key, resource := range info {
		result[key] = r.convertResourceItemToDomainModel(&resource)
	}
	return result
}

func (r *SharedStorageClusterRepository) convertResourceItemToDomainModel(info *v1.AdditionalResourceCfg) *commondomain.InstanceResource {
	return &commondomain.InstanceResource{
		CPUCores:    info.CPUCores,
		LimitMemory: info.LimitMemory,
		Config:      info.Config,
	}
}

func (r *SharedStorageClusterRepository) convertClassInfoToDomainModel(info *v1.InstanceClassInfo) *commondomain.ClassInfo {
	return &commondomain.ClassInfo{
		ClassName: info.ClassName,
		CPU:       info.Cpu,
		Memory:    info.Memory,
	}
}

func (r *SharedStorageClusterRepository) convertImageInfoToDomainModel(imageInfo *v1.VersionInfo) *commondomain.ImageInfo {
	return &commondomain.ImageInfo{
		Version: imageInfo.VersionName,
		Images: map[string]string{
			mpddefine.EngineImageName:         imageInfo.EngineImage,
			mpddefine.ManagerImageName:        imageInfo.ManagerImage,
			mpddefine.PfsdImageName:           imageInfo.PfsdImage,
			mpddefine.PfsdToolImageName:       imageInfo.PfsdToolImage,
			mpddefine.ClusterManagerImageName: imageInfo.ClusterManagerImage,
		},
	}
}

//

func (r *SharedStorageClusterRepository) convertImageInfoToKube(imageInfo *commondomain.ImageInfo) *v1.VersionInfo {
	if imageInfo == nil {
		return nil
	}
	return &v1.VersionInfo{
		VersionName:         imageInfo.Version,
		EngineImage:         imageInfo.Images[mpddefine.EngineImageName],
		ManagerImage:        imageInfo.Images[mpddefine.ManagerImageName],
		PfsdImage:           imageInfo.Images[mpddefine.PfsdImageName],
		PfsdToolImage:       imageInfo.Images[mpddefine.PfsdToolImageName],
		ClusterManagerImage: imageInfo.Images[mpddefine.ClusterManagerImageName],
	}
}
