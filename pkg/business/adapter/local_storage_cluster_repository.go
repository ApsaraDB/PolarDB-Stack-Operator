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
	"reflect"

	commondomain "github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"
	v1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/domain"
	mpddefine "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/define"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"github.com/go-logr/logr"
)

func NewLocalStorageClusterRepository(logger logr.Logger) *LocalStorageClusterRepository {
	return &LocalStorageClusterRepository{
		logger: logger,
	}
}

type LocalStorageClusterRepository struct {
	logger logr.Logger
}

func (r *LocalStorageClusterRepository) GetAll() ([]*domain.LocalStorageCluster, error) {
	panic("implement me")
}

func (r *LocalStorageClusterRepository) GetByName(name, namespace string) (*domain.LocalStorageCluster, error) {
	panic("implement me")
}

func (r *LocalStorageClusterRepository) Create(cluster *domain.LocalStorageCluster) error {
	panic("implement me")
}

func (r *LocalStorageClusterRepository) Update(cluster *domain.LocalStorageCluster) error {
	// write back to k8s status
	kubeRes, err := getKubeResource(cluster.Name, cluster.Namespace)
	if err != nil {
		return err
	}
	return r.update(cluster, kubeRes)
}

func (r *LocalStorageClusterRepository) update(model *domain.LocalStorageCluster, kubeRes *v1.MPDCluster) error {
	oldKubeRes := kubeRes.DeepCopy()
	kubeRes.Spec.ResourceAdditional = convertResourceToKube(model.Resources)

	kubeRes.Status.LogicInsId = model.LogicInsId
	kubeRes.Status.FollowerNum = kubeRes.Spec.FollowerNum
	kubeRes.Status.DBInstanceStatus = map[string]*v1.MPDClusterInstanceStatus{}

	if model.ClusterStatus != "" {
		kubeRes.Status.ClusterStatus = statemachine.State(model.ClusterStatus)
	}
	if model.ClusterManager != nil {
		kubeRes.Status.ClusterManagerStatus = v1.MPDClusterManagerStatus{
			WorkingPort: model.ClusterManager.Port,
		}
	}
	if model.Ins != nil && model.Ins.InsId != "" {
		kubeRes.Status.LeaderInstanceId = model.Ins.InsId
		insStatus := r.convertToKubeInsStatus(model.Ins, v1.MPDClusterInstanceTypeRW, oldKubeRes)
		kubeRes.Status.DBInstanceStatus[model.Ins.InsId] = insStatus
	}

	if !reflect.DeepEqual(kubeRes.Status, oldKubeRes.Status) {
		return updateKubeMpdClusterStatus(kubeRes)
	}
	return nil

}

func (r *LocalStorageClusterRepository) UpdateRunningStatus(name, namespace string) error {
	panic("implement me")
}

func (r *LocalStorageClusterRepository) GetByData(data interface{}, useModifyClass bool, useUpgradeVersion bool) *domain.LocalStorageCluster {
	obj := data.(*v1.MPDCluster)
	return r.convertToDomainModel(obj, useModifyClass, useUpgradeVersion)
}

func (r *LocalStorageClusterRepository) convertToDomainModel(k8sObj *v1.MPDCluster, useModifyClass bool, useUpgradeVersion bool) *domain.LocalStorageCluster {
	result := &domain.LocalStorageCluster{
		RoReplicas: k8sObj.Spec.FollowerNum,
		Port:       k8sObj.Spec.NetCfg.EngineStartPort,
		DbClusterBase: commondomain.DbClusterBase{
			Name:                k8sObj.Name,
			Namespace:           k8sObj.Namespace,
			Description:         k8sObj.Spec.Description,
			ClusterStatus:       string(k8sObj.Status.ClusterStatus),
			LogicInsId:          k8sObj.Status.LogicInsId,
			ImageInfo:           r.convertImageInfoToDomainModel(&k8sObj.Spec.VersionCfg),
			DbClusterType:       commondomain.DbClusterTypeMaster,
			EngineType:          commondomain.EngineTypeRwo,
			Interrupt:           false,
			InterruptMsg:        "",
			InterruptReason:     "",
			ResourceVersion:     k8sObj.ResourceVersion,
			UseModifyClass:      useModifyClass,
			UseUpgradeImageInfo: useUpgradeVersion,
		},
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
			insModel.ResourceName = ins.PodName
			insModel.ResourceNamespace = k8sObj.Namespace
			insModel.StorageHostId = ins.PolarFsHostId
			insModel.Installed = ins.Installed
			result.Ins = insModel
		}
	}

	if result.ClusterStatus == string(statemachine.StateInterrupt) {
		result.Interrupt = true
		result.InterruptMsg = k8sObj.Annotations[mpddefine.DefaultWfConf[define.WFInterruptMessage]]
		result.InterruptReason = k8sObj.Annotations[mpddefine.DefaultWfConf[define.WFInterruptReason]]
	}
	return result
}

func (r *LocalStorageClusterRepository) convertImageInfoToDomainModel(imageInfo *v1.VersionInfo) *commondomain.ImageInfo {
	return &commondomain.ImageInfo{
		Version: imageInfo.VersionName,
		Images: map[string]string{
			mpddefine.EngineImageName:         imageInfo.EngineImage,
			mpddefine.ManagerImageName:        imageInfo.ManagerImage,
			mpddefine.ClusterManagerImageName: imageInfo.ClusterManagerImage,
		},
	}
}

func (r *LocalStorageClusterRepository) convertToKubeInsStatus(ins *commondomain.DbIns, insType v1.MPDClusterInstanceType, oldKubeRes *v1.MPDCluster) *v1.MPDClusterInstanceStatus {
	insStatus := &v1.MPDClusterInstanceStatus{
		PhysicalInsId: ins.PhysicalInsId,
		InsId:         ins.InsId,
		InsName:       ins.InsName,
		PodName:       ins.ResourceName,
		PodNameSpace:  ins.ResourceNamespace,
		NodeName:      ins.Host,
		HostClientIP:  ins.HostIP,
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
