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
	adapter2 "gitlab.alibaba-inc.com/polar-as/polar-common-domain/business/adapter"
	workflow_shared "gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/business/workflow/shared"
	"gotest.tools/v3/assert"
	"strings"
	"testing"
)

func TestGetCreateEngineAccountEnvirons(t *testing.T) {

	ctx := context.Background()
	cc := prepareEnv(t)

	prepareConfigMap(t, ctx, cc)
	clusterName := getTestMPDClusterName()
	cluster := getTestMPDCluster()
	createResourceWithStatusUpdate(t, ctx, cc, cluster, clusterName)
	createUserParams(t, ctx, cc, clusterName.Name)
	createRunningParams(t, ctx, cc, clusterName.Name)
	createAccountAuroraSecret(t, ctx, cc)
	createAccountReplicatorSecret(t, ctx, cc)

	logger := getLogger()
	createRwPodStep := workflow_shared.CreateRwPod{}
	createRwPodStep.Init(map[string]interface{}{
		"_resourceName": "mpdcluster-open-test",
		"_resourceNameSpace": "default",
	}, logger)
	ins := createRwPodStep.Model.RwIns
	ins.ManagerClient.SetIns(ins)
	envsList, err := ins.ManagerClient.(*adapter2.ManagerClient).EnvGetStrategy.GetCreateEngineAccountEnvirons()

	assert.NilError(t, err)
	assert.Assert(t, len(envsList) == 2)
	assert.Assert(t, strings.Contains(envsList[0], "aurora-test"))
	assert.Assert(t, strings.Contains(envsList[1], "replicator-test"))
}