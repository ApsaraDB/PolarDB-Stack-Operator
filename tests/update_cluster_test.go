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

package tests

import (
	"context"
	commondomain "github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"
	mpdv1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/adapter"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/domain"
	"gotest.tools/v3/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func TestUpdateCluster(t *testing.T) {

	ctx := context.Background()
	cc := prepareEnv(t)

	clusterName := getTestMPDClusterName()
	cluster := getTestMPDCluster()

	createResource(t, ctx, cc, cluster, clusterName)

	domainCluster := &domain.SharedStorageCluster{
		SharedStorageDbClusterBase: commondomain.SharedStorageDbClusterBase{
			DbClusterBase: commondomain.DbClusterBase{
				Name:      clusterName.Name,
				Namespace: clusterName.Namespace,
				ImageInfo: &commondomain.ImageInfo{},
				Resources: map[string]*commondomain.InstanceResource{
					"engine": &commondomain.InstanceResource{
						CPUCores:    resource.MustParse("1100m"),
						LimitMemory: resource.MustParse("2Gi"),
						Config:      "",
					},
				},
			},
		},
		RoReplicas: 0,
		Port:       0,
		RwIns:      nil,
		RoInses: map[string]*commondomain.DbIns{
			"3": &commondomain.DbIns{
				DbInsId: commondomain.DbInsId{
					PhysicalInsId: "2",
					InsId:         "3",
				},
				ResourceName:      "mpdcluster-open-test-2-3",
				ResourceNamespace: "default",
			},
		},
		TempRoInses: nil,
		TempRoIds:   nil,
	}

	logger := getLogger()
	r := adapter.NewSharedStorageClusterRepository(logger)
	r.Update(domainCluster)

	resCluster := &mpdv1.MPDCluster{}
	assert.NilError(t, cc.Get(ctx, client.ObjectKey{
		Namespace: clusterName.Namespace,
		Name:      clusterName.Name,
	}, resCluster))

	engineCfg := resCluster.Spec.ResourceAdditional["engine"]
	assert.Equal(t, engineCfg.CPUCores, resource.MustParse("1100m"))
	roStatus := resCluster.Status.DBInstanceStatus["3"]
	assert.Assert(t, roStatus.PhysicalInsId == "2")
	assert.Assert(t, roStatus.PodName == "mpdcluster-open-test-2-3")
}
