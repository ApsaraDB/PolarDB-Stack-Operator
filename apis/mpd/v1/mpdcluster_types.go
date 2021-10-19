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


/*


Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type MPDClusterType string

const (
	MPDClusterSharedVol MPDClusterType = "share"
	MPDClusterLocalVol  MPDClusterType = "local"
)

type FloatingIPMode string

const (
	FloatingIPModeVirtualIF FloatingIPMode = "VirtualIF" //虚拟网卡
	FloatingIPModeMultiIP   FloatingIPMode = "MultiIP"   //单网卡多IP
)

type MaxScaleResourceType string

const (
	MaxScaleResourceTypeStand           MaxScaleResourceType = "stand"
	MaxScaleResourceTypeHighPerformance MaxScaleResourceType = "high"
	MaxScaleResourceTypeLowPerformance  MaxScaleResourceType = "low"
)

type MPDClusterInstanceRole string

const (
	MPDClusterInstanceRoleLeader   MPDClusterInstanceRole = "leader"
	MPDClusterInstanceRoleFollower MPDClusterInstanceRole = "follower"
	MPDClusterInstanceRoleLogger   MPDClusterInstanceRole = "logger"
)

type MPDClusterInstanceType string

const (
	MPDClusterInstanceTypeRW     MPDClusterInstanceType = "rw"
	MPDClusterInstanceTypeRO     MPDClusterInstanceType = "ro"
	MPDClusterInstanceTypeTempRO MPDClusterInstanceType = "tempro"
)

type MPDClusterConditionType string

const (
	ClusterConditionLocalVolReady MPDClusterConditionType = "localVolReady"
	ClusterConditionLeaderIPReady MPDClusterConditionType = "LeaderIPReady"
	ClusterConditionFollowerReady MPDClusterConditionType = "FollowerReady"
	ClusterConditionProxyReady    MPDClusterConditionType = "ProxyReady"
)

type InstanceStatusType string

//专用于额外定义资源相关参数配置
type AdditionalResourceCfg struct {
	CPUCores    resource.Quantity `json:"cpu_cores"`
	LimitMemory resource.Quantity `json:"limit_memory"`
	Config      string            `json:"config"`
}

type InstanceClassInfo struct {
	ClassName string `json:"className"`
	Cpu       string `json:"cpu,omitempty"`
	Memory    string `json:"memory,omitempty"`
	Iops      string `json:"iops,omitempty"`
}

type ShareStoreConfig struct {
	Drive             string `json:"drive"`                  // 存储driver类型， 取值：wwid 或 pvc。对于有存储控制权的情况，应使用pvc模式，其逻辑是：用户先在存储管理中，把lun, lv ,pvc创建好，然后再在此处关联pvc名称
	SharePvcName      string `json:"sharePvcName,omitempty"` // drive=pvc时，创建时必填
	SharePvcNameSpace string `json:"sharePvcNamespace,omitempty"`
	VolumeId          string `json:"volumeId,omitempty"`
	VolumeType        string `json:"volumeType,omitempty"`
	DiskQuota         string `json:"diskQuota,omitempty"`

	ShareVolumeWWID string `json:"shareVolumeWWID,omitempty"` // drive=wwid时有效，必填.由MPDClusterController依据wwid以及DB集群名称，创建pvc，然后再更新回SharePvcName,SharePvcNameSpace信息
	IsVolumeFormat  bool   `json:"isVolumeFormat,omitempty"`  // drive=wwid时有效
	VolumeFormat    string `json:"volumeFormat,omitempty"`    // IsVolumeFormat = true时有效，默认为pfs
}

type MaxScaleInfo struct {
	Enabled          bool                 `json:"enabled"`          // 是否开启，三节点中，默认为true
	Name             string               `json:"name"`             //maxScale 名称
	ResourceType     MaxScaleResourceType `json:"resourceType"`     //资源类型
	ConsistencyLevel int                  `json:"consistencyLevel"` //会话一致性级别
}

type DBNetConfig struct {
	NetType             string          `json:"netType,omitempty"`             //网络类型，目前仅支持host
	EngineStartPort     int             `json:"engineStartPort,omitempty"`     //引擎服务端口开始端口
	EngineAddress       string          `json:"engineAddress,omitempty"`       //引擎服务网卡或者IP，默认为主机的客户网卡，无需配置
	ProxyStartPort      int             `json:"proxyStartPort,omitempty"`      //代理服务端口开始端口
	ProxyNetIF          string          `json:"proxyNetIF,omitempty"`          //代理服务网卡或者IP，默认为主机的客户网卡，无需配置
	PortStep            int             `json:"portStep,omitempty"`            //如有引擎或代理工作在同一主机上，在前一主机可用的情况下，端口增加步长，默认为5，小于0表示使用默认值，等于0相当于不允许在同一主机上运行,最大值100，超出100使用默认值5
	EnableEngineAdminIP bool            `json:"enableEngineAdminIP,omitempty"` //开放引擎管理网访问，默认为false，不开放
	EnableProxyAdminIP  bool            `json:"enableProxyAdminIP,omitempty"`  //开放代理管理网访问，默认为false，不开放
	LeaderFloatingIP    *FloatingIPAddr `json:"leaderFloatingIP,omitempty"`    // 默认为无，不为leader创建浮动IP
}

type FloatingIPAddr struct {
	IPAddressCfgType string         `json:"ipAddressCfgType,omitempty"` //取值两种：auto: 自动分配IP, manual:手动指定IP
	IPAddress        string         `json:"ipAddress,omitempty"`
	Mask             int            `json:"mask,omitempty"`
	GateWay          string         `json:"gateWay,omitempty"`    //+optional
	BasedNetIf       string         `json:"basedNetIf,omitempty"` //IP创建基准物理网卡
	IPMode           FloatingIPMode `json:"ipMode,omitempty"`     // IP创建模式：默认虚拟网卡模式
}

type MPDClusterSpec struct {
	OperatorName       string                           `json:"operatorName"`
	DBClusterType      MPDClusterType                   `json:"dbClusterType"`                //集群类型，默认: local
	DBType             string                           `json:"dbType,omitempty"`             //数据库类型，默认polar-o
	Description        string                           `json:"description,omitempty"`        // 数据库描述，用于存放名称，描述等
	FollowerNum        int                              `json:"followerNum"`                  // 从节点数量, share模式下，表示ro数量，local模式下，表示follower数量。local模式下，取值可以2、4,如为其它值，取默认值2
	ClassInfo          InstanceClassInfo                `json:"classInfo"`                    // 实例规格名称
	ClassInfoModifyTo  InstanceClassInfo                `json:"classInfoModifyTo"`            //实例需要变更的规格名称，不为空，且与DBInsClassName值不同时，将触发变配
	ResourceAdditional map[string]AdditionalResourceCfg `json:"resourceAdditional,omitempty"` // 资源额外补充信息（尽量少用，临进扩展用信息），映射关系：容器或pod名称-->配置内容，便于扩展
	LocalVolName       string                           `json:"localVolName,omitempty"`       // ClusterType = local时有效，有效时不可为空, 与 MPDLocalVol.Name对应,如存储具备管理权限，那么，api提交时，也应先创建好MPDLocalVol对象，并且设置DBCluster Owner Ref为MPDLocalVol对象
	ShareStore         *ShareStoreConfig                `json:"shareStore,omitempty"`         // ClusterType = share时有效，有效时不可为空
	DBProxyInfo        MaxScaleInfo                     `json:"dbProxyInfo,omitempty"`        //DB代理信息
	NetCfg             DBNetConfig                      `json:"netCfg,omitempty"`             //DB的网络配置
	VersionCfg         VersionInfo                      `json:"versionCfg"`                   // 版本信息
	VersionCfgModifyTo VersionInfo                      `json:"versionCfgModifyTo,omitempty"` // 版本升级信息
}

type VersionInfo struct {
	VersionName         string            `json:"versionName,omitempty"` //指代版本定义的cm的名称
	EngineImage         string            `json:"engineImage,omitempty"`
	ManagerImage        string            `json:"managerImage,omitempty"`
	ClusterManagerImage string            `json:"clusterManagerImage,omitempty"`
	PfsdImage           string            `json:"pfsdImage,omitempty"`     //可选，仅pfsd文件系统时有效
	PfsdToolImage       string            `json:"pfsdToolImage,omitempty"` //可选，仅pfsd文件系统时有效
	OtherImages         map[string]string `json:"otherImages,omitempty"`   //扩展空间，如有新的定义，可通过此属性扩展，映射关系： 容器名称->容器Image
}

type DBInstanceNetInfo struct {
	NetType              string `json:"netType,omitempty"` //网络类型，目前仅支持host
	WorkingPort          int    `json:"workingPort,omitempty"`
	WorkingHostIP        string `json:"workingHostIP,omitempty"`
	EnableWorkingAdminIP bool   `json:"enableWorkingAdminIP,omitempty"` //开放引擎管理网访问，默认为false，不开放
	WorkingAdminIP       string `json:"workingAdminIP,omitempty"`
}

type MPDClusterInstanceState struct {
	Reason    string       `json:"reason,omitempty"`
	State     string       `json:"state,omitempty"`
	ErrorInfo string       `json:"errorInfo,omitempty"`
	StartedAt *metav1.Time `json:"startAt,omitempty"`
	FinishAt  *metav1.Time `json:"finishAt,omitempty"`
}

type MPDClusterInstanceStatus struct {
	PhysicalInsId string                  `json:"physicalInsId,omitempty"` // 物理ID
	InsId         string                  `json:"insId,omitempty"`         // 实例ID
	InsName       string                  `json:"insName,omitempty"`       // 等于InsId的初始值，后续重建实例该值不会变
	PodName       string                  `json:"podName,omitempty"`
	PodNameSpace  string                  `json:"podNameSpace,omitempty"`
	NodeName      string                  `json:"nodeName,omitempty"`
	HostClientIP  string                  `json:"hostClientIP,omitempty"`
	PolarFsHostId string                  `json:"polarFsHostId,omitempty"`
	Installed     bool                    `json:"installed,omitempty"`
	Role          MPDClusterInstanceRole  `json:"role,omitempty"`
	InsType       MPDClusterInstanceType  `json:"insType,omitempty"`
	Status        InstanceStatusType      `json:"status,omitempty"`
	VersionInfo   VersionInfo             `json:"versionInfo,omitempty"`
	InsClassInfo  InstanceClassInfo       `json:"insClassInfo,omitempty"`
	NetInfo       DBInstanceNetInfo       `json:"netInfo,omitempty"`
	CurrentState  MPDClusterInstanceState `json:"currentState,omitempty"`
	LastState     MPDClusterInstanceState `json:"lastState,omitempty"`
}

type MPDClusterManagerStatus struct {
	WorkingPort int    `json:"workingPort,omitempty"`
	DeployName  string `json:"deployName,omitempty"`
}

type DBMaxScaleStatus struct {
	WorkingPort       int                 `json:"workingPort,omitempty"`
	MaxScaleName      string              `json:"maxScaleName,omitempty"`
	MaxScaleNameSpace string              `json:"maxScaleNameSpace,omitempty"`
	WorkingAddr       []DBInstanceNetInfo `json:"workingAddr,omitempty"`
}

type MPDClusterCondition struct {
	Type MPDClusterConditionType `json:"type" protobuf:"bytes,1,opt,name=type,casttype=MPDClusterConditionType"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status" protobuf:"bytes,2,opt,name=status,casttype=ConditionStatus"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty" protobuf:"bytes,3,opt,name=lastProbeTime"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty" protobuf:"bytes,4,opt,name=lastTransitionTime"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty" protobuf:"bytes,5,opt,name=reason"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty" protobuf:"bytes,6,opt,name=message"`
}

type MPDClusterStatus struct {
	ClusterStatus        statemachine.State                   `json:"clusterStatus,omitempty"`
	FollowerNum          int                                  `json:"followerNum,omitempty"`
	InsClassInfo         InstanceClassInfo                    `json:"insClassInfo,omitempty"` //全局要求规格定义
	FloatingIP           *FloatingIPAddr                      `json:"floatingIP,omitempty"`
	LocalVolName         string                               `json:"localVolName,omitempty"`
	LeaderInstanceId     string                               `json:"leaderInstanceId,omitempty"`
	LeaderInstanceHost   string                               `json:"leaderInstanceHost,omitempty"`
	LogicInsId           string                               `json:"logicId,omitempty"`
	DBInstanceStatus     map[string]*MPDClusterInstanceStatus `json:"dbInstanceStatus,omitempty"` //每实例状态
	ClusterManagerStatus MPDClusterManagerStatus              `json:"clusterManagerStatus,omitempty"`
	ProxyStatus          DBMaxScaleStatus                     `json:"proxyStatus,omitempty"`
	Conditions           []MPDClusterCondition                `json:"conditions,omitempty"`
}

// +genclient
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MPDCluster is the Schema for the mpdclusters API
type MPDCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MPDClusterSpec   `json:"spec,omitempty"`
	Status MPDClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MPDClusterList contains a list of MPDCluster
type MPDClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MPDCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MPDCluster{}, &MPDClusterList{})
}
