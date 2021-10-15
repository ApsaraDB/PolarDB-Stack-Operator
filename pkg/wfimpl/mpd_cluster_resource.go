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


package wfimpl

import (
	"context"

	apierrors "k8s.io/apimachinery/pkg/api/errors"

	v1 "gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/apis/mpd/v1"

	"github.com/go-logr/logr"
	mgr "gitlab.alibaba-inc.com/polar-as/polar-common-domain/manager"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/implement"
	"gitlab.alibaba-inc.com/polar-as/polar-wf-engine/statemachine"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
)

type MpdClusterResource struct {
	implement.KubeResource
	Logger logr.Logger
}

func (s *MpdClusterResource) GetMpdCluster() *v1.MPDCluster {
	return s.Resource.(*v1.MPDCluster)
}

// GetState 获取资源当前状态
func (s *MpdClusterResource) GetState() statemachine.State {
	return s.GetMpdCluster().Status.ClusterStatus
}

// UpdateState 更新资源当前状态(string)
func (s *MpdClusterResource) UpdateState(state statemachine.State) (statemachine.StateResource, error) {
	so, err := s.fetch()
	mpdCluster := so.Resource.(*v1.MPDCluster)
	mpdCluster.Status.ClusterStatus = state
	if mgr.GetSyncClient().Status().Update(context.TODO(), mpdCluster); err != nil {
		s.Logger.Error(err, "update mpd cluster status error")
		return nil, err
	}
	return so, nil
}

// 更新资源信息
func (s *MpdClusterResource) Update() error {
	if err := mgr.GetSyncClient().Update(context.TODO(), s.GetMpdCluster()); err != nil {
		s.Logger.Error(err, "update mpd cluster error")
		return err
	}
	return nil
}

// Fetch 重新获取资源
func (s *MpdClusterResource) Fetch() (statemachine.StateResource, error) {
	return s.fetch()
}

// GetScheme ...
func (s *MpdClusterResource) GetScheme() *runtime.Scheme {
	return mgr.GetManager().GetScheme()
}

func (s *MpdClusterResource) IsCancelled() bool {
	mpd, err := s.fetch()
	if err != nil {
		if apierrors.IsNotFound(err) {
			return true
		}
		return false
	}
	return mpd.Resource.GetAnnotations()["cancelled"] == "true" || mpd.Resource.GetDeletionTimestamp() != nil
}

func (s *MpdClusterResource) fetch() (*MpdClusterResource, error) {
	kubeRes := &v1.MPDCluster{}
	err := mgr.GetSyncClient().Get(
		context.TODO(), types.NamespacedName{Name: s.Resource.GetName(), Namespace: s.Resource.GetNamespace()}, kubeRes)
	if err != nil {
		s.Logger.Error(err, "mpd cluster not found")
		return nil, err
	}
	return &MpdClusterResource{
		KubeResource: implement.KubeResource{
			Resource: kubeRes,
		},
		Logger: s.Logger,
	}, nil
}
