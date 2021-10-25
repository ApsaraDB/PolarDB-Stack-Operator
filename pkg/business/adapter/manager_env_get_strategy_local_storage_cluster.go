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

	commonadapter "github.com/ApsaraDB/PolarDB-Stack-Common/business/adapter"
	commondomain "github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/domain"
	"github.com/go-logr/logr"
)

func NewLocalStorageClusterEnvGetStrategy(logger logr.Logger, accountRepository commondomain.IAccountRepository) *LocalStorageClusterEnvGetStrategy {
	result := &LocalStorageClusterEnvGetStrategy{
		EnvGetStrategyBase: commonadapter.EnvGetStrategyBase{
			AccountRepository: accountRepository,
			Logger:            logger.WithValues("component", "LocalStorageClusterEnvGetStrategy"),
		},
	}
	result.EnvGetStrategyBase.GetClusterName = result.GetClusterName
	result.EnvGetStrategyBase.GetFlushEnvConfigMap = result.GetFlushEnvConfigMap
	return result
}

type LocalStorageClusterEnvGetStrategy struct {
	commonadapter.EnvGetStrategyBase
	cluster *domain.LocalStorageCluster
}

func (d *LocalStorageClusterEnvGetStrategy) Load(domainModel interface{}, ins *commondomain.DbIns) error {
	d.cluster = domainModel.(*domain.LocalStorageCluster)
	d.Ins = ins
	d.Logger = d.Logger.WithValues("name", ins.ResourceName)
	return nil
}

func (d *LocalStorageClusterEnvGetStrategy) GetInstallEngineEnvirons(ctx context.Context) (string, error) {
	baseEnvsStr := ""

	portMap, err := getEnvPortMap(d.cluster)
	if err != nil {
		return "", err
	}
	portMapStr, err := portMap.ToString()
	if err != nil {
		return "", err
	}

	// tmp table
	addEnvs := map[string]string{
		// TODO: "mycnf_dict" from configmap
		"srv_opr_type":          "hostins_ops",
		"srv_opr_action":        "setup_install_instance",
		"cluster_custins_info":  "",
		"port":                  portMapStr,
		"logic_ins_id":          d.cluster.Ins.InsId,
		"cust_ins_id":           d.cluster.Ins.PhysicalInsId,
		"ins_id":                d.cluster.Ins.InsId,
		"storage_type":          "local",
		"pod_name":              d.cluster.Ins.ResourceName,
		"create_tablespace_env": `{"tablespace_name":"polar_tmp","tablespace_path":"/log/pg_tmp"}`,
	}

	envs := d.TranslateEnvMap2StringWithBase(addEnvs, baseEnvsStr)
	d.Logger.Info("envs: ", envs)
	return envs, nil
}
