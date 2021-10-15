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

	"github.com/go-logr/logr"
	"github.com/pkg/errors"
	commonadapter "gitlab.alibaba-inc.com/polar-as/polar-common-domain/business/adapter"
	mgr "gitlab.alibaba-inc.com/polar-as/polar-common-domain/manager"
	"gitlab.alibaba-inc.com/polar-as/polar-common-domain/utils/k8sutil"
	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/business/domain"
	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/define"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type LocalStoragePodManager struct {
	commonadapter.PodManagerBase
	cluster *domain.LocalStorageCluster
}

func (q *LocalStoragePodManager) Init(domainModel interface{}) {
	q.cluster = domainModel.(*domain.LocalStorageCluster)
	q.Inited = true
}

func NewLocalStoragePodManager(logger logr.Logger) *LocalStoragePodManager {
	return &LocalStoragePodManager{
		PodManagerBase: commonadapter.PodManagerBase{
			Logger: logger,
		},
	}
}

func (q *LocalStoragePodManager) CreatePodAndWaitReady(ctx context.Context) error {
	q.Logger.Info("CreatePodAndWaitReady resourceName: %v, ctx: %v", q.Ins.ResourceName, ctx)
	if !q.Inited {
		return commonadapter.PodManagerNotInitError
	}
	insPod, err := k8sutil.GetPod(q.Ins.ResourceName, q.cluster.Namespace, q.Logger)
	if err != nil {
		if apierrors.IsNotFound(err) {
			q.Logger.Info("cleanDataFiles succeed")
			newPod, err := q.generatePodInfo()
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
	}

	insPod, err = k8sutil.WaitForPodReady(insPod, true, ctx, q.Logger)
	if err != nil {
		return err
	}
	return q.setHostName(insPod)
}

func (q *LocalStoragePodManager) setHostName(pod *corev1.Pod) error {
	q.Ins.Host = pod.Spec.NodeName
	q.Ins.HostIP = pod.Status.HostIP
	return nil
}

func (q *LocalStoragePodManager) generatePodInfo() (*corev1.Pod, error) {

	engine := q.getEngineContainerInfo(q.cluster.Port)
	manager := q.getManagerContainerInfo()
	t := true
	pod := &corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Name:        q.Ins.ResourceName,
			Namespace:   q.cluster.Namespace,
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Spec: corev1.PodSpec{
			Affinity: &corev1.Affinity{
				//NodeAffinity: nodeAffinity,
				//PodAffinity:     podAffinity,
				//PodAntiAffinity: podAntiAffinity,
			},
			Containers: []corev1.Container{
				engine,
				manager,
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
			// TODO: ct 方便调试
			NodeName: "dbm-01",
		},
	}

	// mount volume
	HOSTPATHTYPE := corev1.HostPathDirectoryOrCreate
	dbclusterLogPath := "/disk1/polardb_mpd/"
	pod.Spec.Volumes = append(pod.Spec.Volumes, []corev1.Volume{
		{
			Name: "config-data",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: dbclusterLogPath + q.Ins.InsId + "/data", // + "/log/",
					Type: &HOSTPATHTYPE,
				},
			},
		},
		{
			Name: "config-log",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: dbclusterLogPath + q.Ins.InsId + "/log", // + "/log/",
					Type: &HOSTPATHTYPE,
				},
			},
		},
		{
			Name: "common-pg-hba-cfg",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/etc/postgres/",
					Type: &HOSTPATHTYPE,
				},
			},
		},
		{
			Name: "data",
			VolumeSource: corev1.VolumeSource{
				HostPath: &corev1.HostPathVolumeSource{
					Path: "/disk2/localstore",
					Type: &HOSTPATHTYPE,
				},
			},
		},
	}...)

	if err := q.setControllerReference(pod); err != nil {
		return nil, err
	}

	return pod, nil
}

func (q *LocalStoragePodManager) getEngineContainerInfo(port int) corev1.Container {
	var MOUNTPROPAGATION = corev1.MountPropagationHostToContainer
	c := corev1.Container{
		Name:            "engine",
		Image:           q.cluster.ImageInfo.Images[define.EngineImageName],
		ImagePullPolicy: "IfNotPresent",
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config-data",
				MountPath: "/data",
			},
			{
				Name:      "config-log",
				MountPath: "/log",
			},
			{
				Name:             "common-pg-hba-cfg", // + "-log",
				MountPath:        "/etc/postgres/",
				SubPath:          "",
				MountPropagation: &MOUNTPROPAGATION,
				ReadOnly:         true,
			},
			{
				Name:      "data",
				MountPath: "/disk1",
			},
		},
		Ports: []corev1.ContainerPort{
			{
				ContainerPort: int32(port),
			},
		},
		Lifecycle: &corev1.Lifecycle{
			PreStop: &corev1.Handler{
				Exec: &corev1.ExecAction{
					Command: []string{
						"/bin/sh",
						"-c",
						"srv_opr_type=hostins_ops srv_opr_action=process_cleanup /docker_script/entry_point.py",
					},
				},
			},
		},
		SecurityContext: &corev1.SecurityContext{
			Capabilities: &corev1.Capabilities{
				Add: []corev1.Capability{
					"SYS_PTRACE",
				},
			},
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:                     resource.MustParse("2"),
				corev1.ResourceMemory:                  resource.MustParse("5Gi"),
				corev1.ResourceHugePagesPrefix + "2Mi": resource.MustParse("12Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:                     resource.MustParse("2"),
				corev1.ResourceMemory:                  resource.MustParse("5Gi"),
				corev1.ResourceHugePagesPrefix + "2Mi": resource.MustParse("12Gi"),
			},
		},
	}
	return c
}

func (q *LocalStoragePodManager) getManagerContainerInfo() corev1.Container {
	var MOUNTPROPAGATION = corev1.MountPropagationHostToContainer
	c := corev1.Container{
		Name:            "manager",
		Image:           q.cluster.ImageInfo.Images[define.ManagerImageName],
		ImagePullPolicy: "IfNotPresent",
		VolumeMounts: []corev1.VolumeMount{
			{
				Name:      "config-data",
				MountPath: "/data",
			},
			{
				Name:      "config-log",
				MountPath: "/log",
			},
			{
				Name:             "common-pg-hba-cfg", // + "-log",
				MountPath:        "/etc/postgres/",
				SubPath:          "",
				MountPropagation: &MOUNTPROPAGATION,
				ReadOnly:         true,
			},
			{
				Name:      "data",
				MountPath: "/disk1",
			},
		},
		Env: []corev1.EnvVar{
			{
				Name:  "logic_ins_id",
				Value: q.cluster.LogicInsId,
			},
		},
		Resources: corev1.ResourceRequirements{
			Requests: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
			Limits: corev1.ResourceList{
				corev1.ResourceCPU:    resource.MustParse("200m"),
				corev1.ResourceMemory: resource.MustParse("1Gi"),
			},
		},
	}
	return c
}

func (q *LocalStoragePodManager) setControllerReference(pod *corev1.Pod) error {
	kubeRes, err := getKubeResource(q.cluster.Name, q.cluster.Namespace)
	if err != nil {
		return err
	}
	if err := controllerutil.SetControllerReference(kubeRes, pod, mgr.GetManager().GetScheme()); err != nil {
		return err
	}
	return nil
}
