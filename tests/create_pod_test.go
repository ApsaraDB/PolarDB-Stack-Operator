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
	commonconfig "github.com/ApsaraDB/PolarDB-Stack-Common/configuration"
	"github.com/ApsaraDB/PolarDB-Stack-Common/utils/k8sutil"
	workflow_shared "github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/business/workflow/shared"
	"gotest.tools/v3/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes/fake"
	testcore "k8s.io/client-go/testing"
	"testing"
)

func TestCreatePod(t *testing.T) {

	ctx := context.Background()

	cc := prepareEnv(t)

	logger := getLogger()

	workflow_shared.Register()
	conf := commonconfig.GetConfig()
	conf.HwCheck = false
	commonconfig.SetConfig(conf)

	prepareConfigMap(t, ctx, cc)

	clusterName := getTestMPDClusterName()
	cluster := getTestMPDCluster()
	createResourceWithStatusUpdate(t, ctx, cc, cluster, clusterName)

	clientSet := fake.NewSimpleClientset(&v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hello",
			Namespace: "default",
		}},
		&v1.Node{
			TypeMeta:   metav1.TypeMeta{},
			ObjectMeta: metav1.ObjectMeta{
				Name: "test-01",
			},
			Spec:       v1.NodeSpec{},
			Status:     v1.NodeStatus{},
		},
	)
	k8sutil.ClientForTest = clientSet.CoreV1()

	watcher := watch.NewFake()
	clientSet.PrependWatchReactor("pods", testcore.DefaultWatchReactor(watcher, nil))
	go func() {
		pod := &v1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name: "mpdcluster-open-test-0-1",
				Namespace: "default",
			},
			Status: v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Ready: true,
						Name: "pfsd",
					},
				},
				Conditions: []v1.PodCondition{
					{
						Type: v1.PodReady,
						Status: v1.ConditionTrue,
					},
				},
			},
		}
		watcher.Add(pod)
		watcher1 := watch.NewFake()
		clientSet.PrependWatchReactor("pods", testcore.DefaultWatchReactor(watcher1, nil))
		watcher1.Add(pod)
	}()

	createRwPodStep := workflow_shared.CreateRwPod{}
	createRwPodStep.Init(map[string]interface{}{
		"_resourceName": "mpdcluster-open-test",
		"_resourceNameSpace": "default",
	}, logger)
	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	err := createRwPodStep.Model.RwIns.CreatePod(cancelCtx)

	assert.NilError(t, err)
}
