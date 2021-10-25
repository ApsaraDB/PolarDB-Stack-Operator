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
	"strconv"

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/domain"

	"github.com/go-logr/logr"

	commonadapter "github.com/ApsaraDB/PolarDB-Stack-Common/business/adapter"
	commondomain "github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"
	commondefine "github.com/ApsaraDB/PolarDB-Stack-Common/define"
	mgr "github.com/ApsaraDB/PolarDB-Stack-Common/manager"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func NewSharedStorageClusterEnvGetStrategy(logger logr.Logger, accountRepository commondomain.IAccountRepository) *SharedStorageClusterEnvGetStrategy {
	result := &SharedStorageClusterEnvGetStrategy{
		EnvGetStrategyBase: commonadapter.EnvGetStrategyBase{
			AccountRepository: accountRepository,
			Logger:            logger.WithValues("component", "SharedStorageClusterEnvGetStrategy"),
		},
	}
	result.EnvGetStrategyBase.GetClusterName = result.GetClusterName
	result.EnvGetStrategyBase.GetFlushEnvConfigMap = result.GetFlushEnvConfigMap
	return result
}

type SharedStorageClusterEnvGetStrategy struct {
	commonadapter.EnvGetStrategyBase
	cluster *domain.SharedStorageCluster
}

func (d *SharedStorageClusterEnvGetStrategy) Load(domainModel interface{}, ins *commondomain.DbIns) error {
	d.cluster = domainModel.(*domain.SharedStorageCluster)
	d.Ins = ins
	d.Logger = d.Logger.WithValues("name", ins.ResourceName)
	return nil
}

func (d *SharedStorageClusterEnvGetStrategy) GetClusterName() string {
	return d.cluster.Name
}

func (d *SharedStorageClusterEnvGetStrategy) GetInstallEngineEnvirons(ctx context.Context) (string, error) {
	baseEnvs, err := d.GetCommonEnv(commondefine.ManagerSrvTypeIns, commondefine.ManagerActionSetupInstall, true)
	if err != nil {
		return "", err
	}

	addEnvs := map[string]string{
		"create_tablespace_env": `{"tablespace_name":"polar_tmp","tablespace_path":"/log/pg_tmp"}`,
	}
	isRoParam := ctx.Value("isRo")
	if isRoParam != nil && isRoParam.(bool) == true {
		addEnvs["lock_install_ins"] = "True"
	}

	envs := d.TranslateEnvMap2StringWithBase(addEnvs, baseEnvs)
	return envs, nil
}

func (d *SharedStorageClusterEnvGetStrategy) GetFlushEnvConfigMap() (*v1.ConfigMap, error) {
	resName := d.Ins.ResourceName
	insId := d.Ins.InsId
	insConfig, _, _, err := d.cluster.GetEffectiveParams(false, false)
	if err != nil {
		return nil, err
	}

	params, err := json.Marshal(insConfig)
	if err != nil {
		return nil, err
	}

	clusterInfo, err := d.getEnvClusterInfo()
	if err != nil {
		return nil, err
	}
	strClusterInfo, err := clusterInfo.ToString()
	if err != nil {
		return nil, err
	}

	port, err := d.GetEnvPort()
	if err != nil {
		return nil, err
	}
	strPort, err := port.ToString()
	if err != nil {
		return nil, err
	}

	uniqueId, err := d.GetUniqueId(d.Ins.ResourceName)
	if err != nil {
		return nil, err
	}

	serviceType := define.RwEnvServiceType
	if d.Ins.InsId != d.cluster.RwIns.InsId {
		serviceType = define.RoEnvServiceType
	}

	kubeObj, err := getKubeResource(d.cluster.Name, d.cluster.Namespace)
	if err != nil {
		return nil, err
	}

	podConfigMap := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      resName,
			Namespace: kubeObj.Namespace,
		},
		Data: map[string]string{
			"cluster_custins_info": strClusterInfo,
			"cust_ins_id":          d.Ins.PhysicalInsId,
			"mycnf_dict":           string(params),
			"storage_type":         define.SharedStorageEnvStorageType,
			"port":                 strPort,
			"logic_ins_id":         d.cluster.LogicInsId,
			"slot_unique_name":     uniqueId,
			"on_pfs":               "True",
			"san_device_name":      "/dev/mapper/" + d.cluster.StorageInfo.VolumeId,
			"polarfs_host_id":      d.Ins.StorageHostId,
			"ins_id":               insId,
			"pod_name":             resName,
			"service_type":         serviceType,
		},
	}

	if err := controllerutil.SetControllerReference(kubeObj, podConfigMap, mgr.GetManager().GetScheme()); err != nil {
		return nil, err
	}

	if _, err = d.GetEnvConfigMap(); err != nil {
		if apierrors.IsNotFound(err) {
			err = mgr.GetSyncClient().Create(context.TODO(), podConfigMap)
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {
		// 每次均重新生成configMap. 保证configMap中数据是准确.
		err = mgr.GetSyncClient().Update(context.TODO(), podConfigMap)
		if err != nil {
			return nil, err
		}
	}

	return podConfigMap, nil
}

func (d SharedStorageClusterEnvGetStrategy) getEnvClusterInfo() (*commonadapter.EnvClusterInfo, error) {
	accounts, err := d.GetAccountsFromMeta()
	if err != nil {
		return nil, err
	}
	rwIns := d.cluster.RwIns
	rwPhyId, err := strconv.Atoi(rwIns.PhysicalInsId)
	if err != nil {
		return nil, err
	}
	rwPbd := commonadapter.Pbd{
		Label:      "primary",
		EngineType: "san",
		CustinsId:  rwPhyId,
	}
	rwInfo := map[string]*commonadapter.PhysicalCustInstance{
		rwIns.PhysicalInsId: d.GetPhysicalCustInstanceEnv(accounts, d.cluster.RwIns.InsId, rwIns.HostIP, d.cluster.Port, &rwPbd, "master", 0),
	}
	roInfo := map[string]*commonadapter.PhysicalCustInstance{}
	for _, ins := range d.cluster.RoInses {
		roInfo[ins.PhysicalInsId] = d.GetPhysicalCustInstanceEnv(accounts, ins.InsId, ins.HostIP, d.cluster.Port, &rwPbd, "slave", 3)
	}
	for _, ins := range d.cluster.TempRoInses {
		roInfo[ins.PhysicalInsId] = d.GetPhysicalCustInstanceEnv(accounts, ins.InsId, ins.HostIP, d.cluster.Port, &rwPbd, "slave", 3)
	}

	envClusterInfo := &commonadapter.EnvClusterInfo{
		RwClusterInfo: rwInfo,
		RoClusterInfo: roInfo,
	}

	return envClusterInfo, nil
}
