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

	v1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"

	commonadapter "github.com/ApsaraDB/PolarDB-Stack-Common/business/adapter"
	"github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"
	mgr "github.com/ApsaraDB/PolarDB-Stack-Common/manager"
	"github.com/go-logr/logr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func NewMdpAccountRepository(logger logr.Logger) *commonadapter.AccountRepository {
	return &commonadapter.AccountRepository{
		Logger:              logger,
		GetKubeResourceFunc: GetKubeResourceByName,
	}
}

func GetKubeResourceByName(name, namespace string, clusterType domain.DbClusterType) (metav1.Object, error) {
	sbl := &v1.MPDCluster{}
	err := mgr.GetSyncClient().Get(context.TODO(),
		types.NamespacedName{Name: name, Namespace: namespace}, sbl)
	if err != nil {
		return nil, err
	}
	return sbl, nil
}
