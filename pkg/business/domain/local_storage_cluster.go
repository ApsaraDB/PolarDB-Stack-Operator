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
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	commondomain "github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	"strconv"
)

type LocalStorageCluster struct {
	commondomain.DbClusterBase
	RoReplicas  int
	Port        int
	inited      bool
	Ins         *commondomain.DbIns
	FollowerIns []*commondomain.DbIns
}

func (cluster *LocalStorageCluster) Init(
	podManager commondomain.IEnginePodManager,
	idGeneragor commondomain.IIdGenerator,
	portGenerator commondomain.IPortGenerator,
	managerClient commondomain.IManagerClient,
	logger logr.Logger,
) {
	cluster.DbClusterBase.Init(
		nil,
		nil,
		nil,
		nil,
		nil,
		idGeneragor,
		portGenerator,
		nil,
		nil,
		nil,
		nil,
		logger,
	)
	managerClient.Init(cluster)
	podManager.Init(cluster)
	cluster.Logger = logger
	if cluster.Ins != nil {
		cluster.Ins.Init(managerClient, podManager)
	}
	cluster.inited = true
}

func (cluster *LocalStorageCluster) InitDbInsMeta() error {
	if !cluster.inited {
		return errors.New("local storage cluster not init.")
	}
	if cluster.Ins == nil || cluster.Ins.InsId == "" {
		return cluster.initNewDbInsMeta()
	}
	return nil
}

func (cluster *LocalStorageCluster) initNewDbInsMeta() error {
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
	if err != nil {
		return err
	}
	cluster.Ins = &commondomain.DbIns{
		DbInsId:           rwDbInsId,
		ResourceName:      cluster.Name + "-" + rwDbInsId.PhysicalInsId + "-" + rwDbInsId.InsId,
		ResourceNamespace: cluster.Namespace,
		NetInfo: &commondomain.NetInfo{
			Port: cluster.Port,
		},
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
