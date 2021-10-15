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
	"fmt"
	"reflect"

	mpddefine "gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/define"

	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/business/domain"

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"gitlab.alibaba-inc.com/polar-as/polar-common-domain/define"

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	commonadapter "gitlab.alibaba-inc.com/polar-as/polar-common-domain/business/adapter"
	commondomain "gitlab.alibaba-inc.com/polar-as/polar-common-domain/business/domain"
	mgr "gitlab.alibaba-inc.com/polar-as/polar-common-domain/manager"
	"gitlab.alibaba-inc.com/polar-as/polar-common-domain/utils"
	"gitlab.alibaba-inc.com/polar-as/polar-common-domain/utils/k8sutil"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
)

func NewSharedStoragePodManager(logger logr.Logger) *SharedStoragePodManager {
	return &SharedStoragePodManager{
		PodManagerBase: commonadapter.PodManagerBase{
			Logger: logger,
		},
	}
}

type SharedStoragePodManager struct {
	commonadapter.PodManagerBase
	cluster *domain.SharedStorageCluster
}

func (q *SharedStoragePodManager) Init(domainModel interface{}) {
	q.cluster = domainModel.(*domain.SharedStorageCluster)
	q.Inited = true
}

func (q *SharedStoragePodManager) CreatePodAndWaitReady(ctx context.Context) error {
	if !q.Inited {
		return commonadapter.PodManagerNotInitError
	}
	insPod, err := k8sutil.GetPod(q.Ins.ResourceName, q.cluster.Namespace, q.Logger)
	if err != nil {
		if apierrors.IsNotFound(err) {
			// 创建Pod前清理所有主机上相同insId的data目录
			// todo: 创建pod前某节点重启中，没有进行清理，pod创建时节点恢复，被调度到该节点。
			if err := q.CleanDataFiles(); err != nil {
				err = errors.Wrap(err, fmt.Sprintf("clean up instance data fail, %v", q.cluster.Name))
				q.Logger.Error(err, "cleanDataFiles failed")
				return err
			}
			q.Logger.Info("cleanDataFiles succeed")
			newPod, err := q.generatePod()
			if err != nil {
				return err
			}
			err = k8sutil.CreatePod(newPod, q.Logger)
			if err != nil {
				return errors.Wrap(err, "create pod fail")
			}
			insPod = newPod
		} else {
			return err
		}
	} else {
		// pod已经存在，上一次可能由于节点故障导致没有调度成功，重试要重新检查节点状态和亲和性。
		// 比较新的亲和性和旧的亲和性，如果不一致，则删掉Pod重建，一直则直接进入等待。
		if insPod.Spec.Affinity != nil &&
			insPod.Spec.Affinity.NodeAffinity != nil &&
			insPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil &&
			insPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms != nil {
			for _, nodeSelectorTerm := range insPod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
				if nodeSelectorTerm.MatchExpressions != nil {
					for _, expression := range nodeSelectorTerm.MatchExpressions {
						if expression.Key == "kubernetes.io/hostname" && expression.Operator == corev1.NodeSelectorOpNotIn {
							taintNodeList := utils.SortStrings(expression.Values)
							newTaintNodeList, err := commonadapter.GetUnAvailableNode(q.Logger)
							if err != nil {
								q.Logger.Error(err, "GetUnAvailableNode error")
								return err
							}
							newTaintNodeList = utils.SortStrings(newTaintNodeList)
							if !reflect.DeepEqual(taintNodeList, newTaintNodeList) {
								q.Logger.Info("taint nodes changed, delete pod and recreate")
								err = k8sutil.DeletePod(insPod.Name, insPod.Namespace, ctx, q.Logger)
								if err != nil {
									err := errors.Wrapf(err, "delete pod %s error", q.Ins.ResourceName)
									q.Logger.Error(err, "")
									return err
								}
								return q.CreatePodAndWaitReady(ctx)
							}
						}
					}
				}
			}
		}
	}

	if insPod, err = k8sutil.WaitForContainerReady(insPod, define.ContainerNamePfsd, ctx, q.Logger); err != nil {
		return err
	}

	pfsdToolClient := commonadapter.NewPfsdToolClient(q.Logger)
	pfsdToolClient.Init(q.Ins, q.cluster.Resources, q.cluster.StorageInfo.VolumeId)
	if err := pfsdToolClient.StartPfsd(ctx); err != nil {
		return err
	}

	insPod, err = k8sutil.WaitForPodReady(insPod, true, ctx, q.Logger)
	if err != nil {
		return err
	}
	return q.setHostName(insPod)
}

func (q *SharedStoragePodManager) setHostName(pod *corev1.Pod) error {
	q.Ins.Host = pod.Spec.NodeName
	q.Ins.HostIP = pod.Status.HostIP
	return nil
}

func (q *SharedStoragePodManager) generatePod() (*corev1.Pod, error) {
	//依据故障节点，选出污点node信息
	taintNodeList, err := commonadapter.BuildNodeAvailableInfo(q.Logger)
	if err != nil {
		return nil, err
	}
	//节点设置亲和性
	nodeAffinity, err := commonadapter.BuildPodNodeAffinity(taintNodeList, q.Ins.TargetNode, q.Logger)
	if err != nil {
		return nil, err
	}
	insType := mpddefine.InsTypeRo
	if q.Ins.InsId == q.cluster.RwIns.InsId {
		insType = mpddefine.InsTypeRw
	}
	objectMeta := commonadapter.BuildPodObjectMeta(*q.Ins, q.cluster.SharedStorageDbClusterBase, insType)

	resConfig, err := commonadapter.GetSysResConfig(q.Logger)
	if err != nil {
		return nil, err
	}

	classes, err := commondomain.NewEngineClasses(commondomain.EngineTypeRwo, q.cluster.ClassQuery)
	if err != nil {
		return nil, err
	}
	instanceClass, err := classes.GetClass(q.cluster.ClassInfo.ClassName)
	if err != nil {
		return nil, err
	}

	pfsdToolsContainer := commonadapter.BuildPfsdToolsContainer(&q.cluster.SharedStorageDbClusterBase, resConfig)

	pfsdContainer := commonadapter.BuildPfsdContainer(&q.cluster.SharedStorageDbClusterBase, resConfig)

	engineContainer := commonadapter.BuildEngineContainer(q.Logger, &q.cluster.SharedStorageDbClusterBase, q.Ins, instanceClass)

	managerContainer := commonadapter.BuildManagerContainer(&q.cluster.SharedStorageDbClusterBase, resConfig)
	t := true
	pod := &corev1.Pod{
		ObjectMeta: objectMeta,
		Spec: corev1.PodSpec{
			Affinity: &corev1.Affinity{
				NodeAffinity: nodeAffinity,
				//PodAffinity:     podAffinity,
				//PodAntiAffinity: podAntiAffinity,
			},
			Containers: []corev1.Container{
				pfsdToolsContainer,
				pfsdContainer,
				engineContainer,
				managerContainer,
			},
			HostNetwork:           true,
			DNSPolicy:             corev1.DNSClusterFirstWithHostNet,
			RestartPolicy:         corev1.RestartPolicyAlways,
			SchedulerName:         "default-scheduler",
			ShareProcessNamespace: &t,
			//volume先置空，之后填上内容
			Volumes: []corev1.Volume{},
			NodeSelector: map[string]string{
				"node.kubernetes.io/node": "",
			},
		},
	}
	pod.Spec.Tolerations = []corev1.Toleration{
		{
			Key:      "node.kubernetes.io/not-ready",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
		{
			Key:      "node.kubernetes.io/unreachable",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
		{
			Key:      "node.kubernetes.io/disk-pressure",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
		{
			Key:      "node.kubernetes.io/memory-pressure",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
		{
			Key:      "node.kubernetes.io/unschedulable",
			Operator: corev1.TolerationOpExists,
			Effect:   corev1.TaintEffectNoExecute,
		},
	}
	kubeRes, err := getKubeResource(q.cluster.Name, q.cluster.Namespace)
	if err != nil {
		return nil, err
	}
	commonadapter.BuildAndRefVolumeInfo(q.cluster.StorageInfo.DiskID, q.Ins.InsId, pod)

	if err := controllerutil.SetControllerReference(kubeRes, pod, mgr.GetManager().GetScheme()); err != nil {
		return nil, err
	}

	return pod, nil
}
