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

	commondomain "github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"
	mgr "github.com/ApsaraDB/PolarDB-Stack-Common/manager"
	v1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/domain"
	"k8s.io/apimachinery/pkg/types"
)

func getKubeResource(name string, namespace string) (*v1.MPDCluster, error) {
	cluster := &v1.MPDCluster{}
	err := mgr.GetSyncClient().Get(context.TODO(),
		types.NamespacedName{Name: name, Namespace: namespace}, cluster)
	if err != nil {
		return nil, err
	}
	return cluster, nil
}

func updateKubeMpdClusterStatus(cluster *v1.MPDCluster) error {
	return mgr.GetSyncClient().Status().Update(context.TODO(), cluster)
}

func convertResourceToKube(resources map[string]*commondomain.InstanceResource) map[string]v1.AdditionalResourceCfg {
	var result = make(map[string]v1.AdditionalResourceCfg)
	for key, resource := range resources {
		result[key] = convertResourceItemToKube(resource)
	}
	return result
}

func convertResourceItemToKube(resource *commondomain.InstanceResource) v1.AdditionalResourceCfg {
	return v1.AdditionalResourceCfg{
		CPUCores:    resource.CPUCores,
		LimitMemory: resource.LimitMemory,
		Config:      resource.Config,
	}
}

type Port struct {
	Link        []int `json:"link"`
	Access_port []int `json:"access_port"`
	Perf_port   []int `json:"perf_port"`
}

// EnvPort  map[string]*Port
// hostins id -> _port
type EnvPort map[string]*Port

func (r *EnvPort) ToString() (string, error) {
	str, err := json.Marshal(r)
	if err != nil {
		return "", err
	}
	return string(str), nil
}

func getEnvPortMap(cluster *domain.LocalStorageCluster) (*EnvPort, error) {
	envPort := EnvPort{}

	port := Port{
		Link:        []int{cluster.Port},
		Access_port: []int{cluster.Port},
		Perf_port:   []int{0},
	}
	envPort[cluster.Ins.InsId] = &port

	return &envPort, nil
}
