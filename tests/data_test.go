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
	"github.com/ApsaraDB/PolarDB-Stack-Common/business/domain"
	mpdv1 "github.com/ApsaraDB/PolarDB-Stack-Operator/apis/mpd/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"testing"
)

func getTestMPDClusterName() types.NamespacedName {

	clusterName := types.NamespacedName{
		Namespace: "default",
		Name:      "mpdcluster-open-test",
	}

	return clusterName
}

func getTestMPDCluster() *mpdv1.MPDCluster {

	clusterName := getTestMPDClusterName()

	cluster := &mpdv1.MPDCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:                       clusterName.Name,
			Namespace:                  clusterName.Namespace,
		},
		Spec: mpdv1.MPDClusterSpec{
			OperatorName:       "test",
			DBClusterType:      mpdv1.MPDClusterSharedVol,
			FollowerNum:        1,
			ClassInfo:          mpdv1.InstanceClassInfo{
				ClassName: "polar.o.x4.medium",
				Cpu: "2000m",
				Memory: "17Gi",
			},
			ClassInfoModifyTo:  mpdv1.InstanceClassInfo{},
			ShareStore:         &mpdv1.ShareStoreConfig{
				Drive:             "pvc",
				SharePvcName:      "pvc-test",
				SharePvcNameSpace: "default",
				VolumeId:          "test",
				VolumeType:        "multipath",
				DiskQuota:         "520000",
			},
			NetCfg:             mpdv1.DBNetConfig{
				EngineStartPort: 5400,
			},
			VersionCfg:         mpdv1.VersionInfo{
				VersionName: "image-open",
			},
			VersionCfgModifyTo: mpdv1.VersionInfo{},
			ResourceAdditional: map[string]mpdv1.AdditionalResourceCfg{
				"pfsd": mpdv1.AdditionalResourceCfg{
					CPUCores:    resource.MustParse("500m"),
					LimitMemory: resource.MustParse("1Gi"),
					Config:      "-w 8 -s 20 -i 8192 -f",
				},
			},
		},
		Status: mpdv1.MPDClusterStatus{
			DBInstanceStatus: map[string]*mpdv1.MPDClusterInstanceStatus{
				"1": &mpdv1.MPDClusterInstanceStatus{
					InsId: "1",
					InsType: "rw",
					PodName: "mpdcluster-open-test-0-1",
					PodNameSpace: "default",
					PhysicalInsId: "0",
				},
			},
		},
	}

	return cluster
}

func getIns() *domain.DbIns {
	ins := &domain.DbIns{
		DbInsId:           domain.DbInsId{
			PhysicalInsId: "4",
			InsId: "5",
		},
		ResourceName:      "mpdcluster-open-test-4-5",
		ResourceNamespace: "default",
	}
	return ins
}

func createInstanceSystemReousrces(t *testing.T, ctx context.Context, cc client.Client) {

	instanceSystemResourcesName := types.NamespacedName{
		Namespace: "kube-system",
		Name:      "instance-system-resources",
	}

	instanceSystemResources := &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: instanceSystemResourcesName.Name,
			Namespace: instanceSystemResourcesName.Namespace,
		},
		Data: map[string]string{
			"enableResourceShare": "true",
			"shared": "single:\n  manager:\n    limits:\n      cpu: 500m\n      memory: 256Mi\n    requests:\n      cpu: 50m\n      memory: 64Mi\nreadWriteMany:\n  pfsdTool:\n    limits:\n      cpu: 500m\n      memory: 256Mi\n    requests:\n      cpu: 500m\n      memory: 256Mi\n  pfsd:\n    limits:\n      cpu: 500m\n      memory: 256Mi\n    requests:\n      cpu: 500m\n      memory: 256Mi\n  manager:\n    limits:\n      cpu: 500m\n      memory: 128Mi\n    requests:\n      cpu: 500m\n      memory: 128Mi\nmaxscale:\n  operator:\n    limits:\n      cpu: 1000m\n      memory: 1Gi\n    requests:\n      cpu: 1000m\n      memory: 1Gi\n",
		},
		BinaryData: nil,
	}

	createResource(t, ctx, cc, instanceSystemResources, instanceSystemResourcesName)
}

func createControllerConfig(t *testing.T, ctx context.Context, cc client.Client) {

	controllerConfigName := types.NamespacedName{
		Namespace: "kube-system",
		Name:      "controller-config",
	}

	controllerConfig := &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: controllerConfigName.Name,
			Namespace: controllerConfigName.Namespace,
		},
		Data: map[string]string{
			"degradeDaemonSet": "",
			"degradeDeployment": "",
			"disabledWfEvents": "",
			"sshUser": "root",
			"sshPassword": "",
			"controllerNodeLabel": "node-role.kubernetes.io/master",
		},
		BinaryData: nil,
	}

	createResource(t, ctx, cc, controllerConfig, controllerConfigName)
}

func createInstanceLevelResource(t *testing.T, ctx context.Context, cc client.Client) {

	instanceLevelResourceName := types.NamespacedName{
		Namespace: "kube-system",
		Name:      "postgresql-1-0-level-polar-o-x4-medium-resource-rwo",
	}

	instanceLevelResource := &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: instanceLevelResourceName.Name,
			Namespace: instanceLevelResourceName.Namespace,
			Labels: map[string]string{
				"classKey": "polar.o.x4.medium",
				"configtype": "instance_level",
				"dbClusterMode": "WriteReadMore",
				"dbType": "PostgreSQL",
				"dbVersion": "1.0",
				"leveltype": "instance_level_resource",
			},
		},
		Data: map[string]string{
		},
		BinaryData: nil,
	}

	createResource(t, ctx, cc, instanceLevelResource, instanceLevelResourceName)
}

func createInstanceLevelConfig(t *testing.T, ctx context.Context, cc client.Client) {

	name := types.NamespacedName{
		Namespace: "kube-system",
		Name:      "postgresql-1-0-level-polar-o-x4-medium-config-rwo",
	}

	obj := &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name.Name,
			Namespace: name.Namespace,
			Labels: map[string]string{
				"classKey": "polar.o.x4.medium",
				"configtype": "instance_level",
				"dbClusterMode": "WriteReadMore",
				"dbType": "PostgreSQL",
				"dbVersion": "1.0",
				"leveltype": "instance_level_config",
			},
		},
		Data: map[string]string{
		},
		BinaryData: nil,
	}

	createResource(t, ctx, cc, obj, name)
}

func createMycnfTemplate(t *testing.T, ctx context.Context, cc client.Client) {
	name := types.NamespacedName{
		Namespace: "kube-system",
		Name:      "postgresql-1-0-mycnf-template-rwo",
	}
	obj := &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name.Name,
			Namespace: name.Namespace,
			Labels: map[string]string{
			},
		},
		Data: map[string]string{
		},
		BinaryData: nil,
	}
	createResource(t, ctx, cc, obj, name)
}

func prepareConfigMap(t *testing.T, ctx context.Context, cc client.Client) {
	createInstanceSystemReousrces(t, ctx, cc)
	createControllerConfig(t, ctx, cc)
	createInstanceLevelResource(t, ctx, cc)
	createInstanceLevelConfig(t, ctx, cc)
	createMycnfTemplate(t, ctx, cc)
}

func createUserParams(t *testing.T, ctx context.Context, cc client.Client, mpdClusterName string) {
	name := types.NamespacedName{
		Namespace: "default",
		Name:      mpdClusterName + "-user-params",
	}
	obj := &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name.Name,
			Namespace: name.Namespace,
			Labels: map[string]string{
			},
		},
		Data: map[string]string{
		},
		BinaryData: nil,
	}
	createResource(t, ctx, cc, obj, name)
}

func createRunningParams(t *testing.T, ctx context.Context, cc client.Client, mpdClusterName string) {
	name := types.NamespacedName{
		Namespace: "default",
		Name:      mpdClusterName + "-running-params",
	}
	obj := &v1.ConfigMap{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name.Name,
			Namespace: name.Namespace,
			Labels: map[string]string{
			},
		},
		Data: map[string]string{
		},
		BinaryData: nil,
	}
	createResource(t, ctx, cc, obj, name)
}

func createAccountAuroraSecret(t *testing.T, ctx context.Context, cc client.Client) {
	name := types.NamespacedName{
		Namespace: "default",
		Name:      "mpdcluster-open-test-1-aurora",
	}
	obj := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name.Name,
			Namespace: name.Namespace,
			Labels: map[string]string{
				"mpdcluster_name": "mpdcluster-open-test",
			},
		},
		StringData: map[string]string{
			"Account": "aurora",
			"Password": "aurora-test",
			"Priviledge_type": "7",
		},
	}
	createResource(t, ctx, cc, obj, name)
}

func createAccountReplicatorSecret(t *testing.T, ctx context.Context, cc client.Client) {
	name := types.NamespacedName{
		Namespace: "default",
		Name:      "mpdcluster-open-test-1-replicator",
	}
	obj := &v1.Secret{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name: name.Name,
			Namespace: name.Namespace,
			Labels: map[string]string{
				"mpdcluster_name": "mpdcluster-open-test",
			},
		},
		StringData: map[string]string{
			"Account": "replicator",
			"Password": "replicator-test",
			"Priviledge_type": "7",
		},
	}
	createResource(t, ctx, cc, obj, name)
}
