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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type MPDLocalVolumeMode string
type MDPLocalVolumeFileSystem string
type MPDLocalVolumeConditionType string

const (
	LocalVolumeModeCustom    MPDLocalVolumeMode = "custom"    //用户提前预分配好lv, stack仅使用
	LocalVolumeModeAutomatic MPDLocalVolumeMode = "automatic" //用户不提前分配lv， 由stack创建
	LocalVolumeModeMixed     MPDLocalVolumeMode = "mixed"     //部分用户提前分配，部分由stack创建

	LocalVolumeFileSystemExt4  MDPLocalVolumeFileSystem = "ext4"
	LocalVolumeFileSystemPfs   MDPLocalVolumeFileSystem = "pfs"
	LocalVolumeFileSystemXfs   MDPLocalVolumeFileSystem = "xfs"
	LocalVolumeFileSystemEmpty MDPLocalVolumeFileSystem = ""

	VolConditionTypeLVReady     MPDLocalVolumeConditionType = "LVReady"
	VolConditionTypeLVSizeReady MPDLocalVolumeConditionType = "LVSizeReady"
	VolConditionTypeVGReady     MPDLocalVolumeConditionType = "VGReady"
	VolConditionTypePVCReady    MPDLocalVolumeConditionType = "PVCReady"
	VolConditionTypeFormatReady MPDLocalVolumeConditionType = "FormatReady"
)

// MPDHostLV holds the basic info for a lvm logical volume, lv_path and hostname
type MPDHostLV struct {
	LvID     string `json:"lvId,omitempty"`
	NodeName string `json:"nodeName,omitempty"`
	LvPath   string `json:"lvPath,omitempty"`
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// MPDLocalVolumeSpec defines the desired state of MPDLocalVolume
type MPDLocalVolumeSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Specify a driver to provision volume, default lvm
	// +optional
	Driver string `json:"driver,omitempty"`

	// The lvm logical volume used by MPDLocalVolume is provisioned by an administrator or dynamically provisioned
	// +optional
	VolumeMode MPDLocalVolumeMode `json:"mpdVolumeMode,omitempty"`

	// How many lvm logical volumes that are desired or claimed.
	// +optional
	LvNum uint64 `json:"lvNum,omitempty"`

	// LvResources contains LVM logical volume information, including basic metadata and location information
	// It takes effect only when volumeMode is custom or mixed
	// +optional
	LvResources []MPDHostLV `json:"lvResources,omitempty"`

	// +optional
	LvExpectedSizeMB uint64 `json:"lvExpectedSizeMB,omitempty"`

	// VolMode = maker 或 mixed时有效，即希望stack扩容至多大的
	LvExpectedExpandToSizeMB uint64 `json:"lvExpectedExpandToSizeMB,omitempty"`

	// TODO: create PVC or not. Reserved for future.
	//+optional
	CreatePVC bool `json:"createPVC,omitempty"`

	// File system, default ext4, empty means no need to format
	// +optional
	FormatFileSystem MDPLocalVolumeFileSystem `json:"formatFileSystem,omitempty"`

	// Description or comments
	// +optional
	Description string `json:"Description,omitempty"`
}

type LVErrorInfo struct {
	LastErrorTime metav1.Time `json:"lastErrorTime,omitempty"`
	LastErrorMsg  string      `json:"lastErrorMsg,omitempty"`
	ErrorCode     string      `json:"errorCode,omitempty"`
}

type MPDHostLVStatus struct {
	MPDHostLV `json:",inline"`
	//lv Name 需要由应用rename成与 MPDLocalVolume.Name 名称相关， 如： MPDLocalVol-01
	LvName string `json:"lvName,omitempty"`
	LvUuid string `json:"lvUUID,omitempty"`

	VgName   string `json:"vgName,omitempty"`
	VgUuid   string `json:"vgUUID,omitempty"`
	LvSizeMB uint64 `json:"lvSizeMB,omitempty"`

	//+optional
	FileSystem string       `json:"fileSystem,omitempty"`
	LvStatus   string       `json:"lvStatus,omitempty"`
	LvSectors  int32        `json:"lvSectors,omitempty"`
	ErrorInfo  *LVErrorInfo `json:"errorInfo,omitempty"`
}

type PVCInfo struct {
	Namespace string       `json:"namespace,omitempty"`
	Name      string       `json:"name,omitempty"`
	ErrorInfo *LVErrorInfo `json:"errorInfo,omitempty"`
}

type LVPVCMap map[string]PVCInfo

type MPDLocalVolumeCondition struct {
	Type MPDLocalVolumeConditionType `json:"type"`
	// Status is the status of the condition.
	// Can be True, False, Unknown.
	Status corev1.ConditionStatus `json:"status"`
	// Last time we probed the condition.
	// +optional
	LastProbeTime metav1.Time `json:"lastProbeTime,omitempty"`
	// Last time the condition transitioned from one status to another.
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
	// Unique, one-word, CamelCase reason for the condition's last transition.
	// +optional
	Reason string `json:"reason,omitempty"`
	// Human-readable message indicating details about last transition.
	// +optional
	Message string `json:"message,omitempty"`
}

// MPDLocalVolumeStatus defines the observed state of MPDLocalVolume
type MPDLocalVolumeStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	//取LvStatusMap中size最小的值
	LvSizeMB uint64 `json:"LvSizeMB,omitempty"`

	// How many lvm logical volumes that are observed.
	// +optional
	LVNum int `json:"lvNum,omitempty"`

	// Number of valid and available LVM logical volumes
	// +optional
	ValidLVNum int `json:"validLVNum,omitempty"`

	// PVC number for each LVM logical volume
	// +optional
	PVCNum int `json:"pvcNumber,omitempty"`

	// Avaliable PVC number in current mode
	// +optional
	PVCReadyNum bool `json:"pvcReadyNum,omitempty"`

	// Whether PVCs are ready
	// +optional
	PVCReady bool `json:"pvcReady,omitempty"`

	//映射关系：lvID->MPDHostLVStatus
	// +optional
	LvStatus map[string]MPDHostLVStatus `json:"lvStatus,omitempty"`

	// Observed file system format
	// +optional
	FormatFileSystem MDPLocalVolumeFileSystem `json:"formatFileSystem,omitempty"`

	// PVC and LVM logical colume map, lvID->LVPVCInfo
	// +optional
	PVCInfo LVPVCMap `json:"pvcInfo,omitempty"`

	// Current Condition of MDPLocalVolume. If underlying persistent volume is being
	// resized then the Condition will be set to 'ResizeStarted'.
	// +optional
	// +patchMergeKey=type
	// +patchStrategy=merge
	Conditions []MPDLocalVolumeCondition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name="lvnum",type="integer",JSONPath=".spec.lvNum",description="lvm logical volumes number claimed"
// +kubebuilder:printcolumn:name="current lvnum",type="integer",JSONPath=".status.validLVNum",description="current lvm logical volumes available"
// +kubebuilder:printcolumn:name="pvcready",type=boolean,JSONPath=.status.pvcReady,description="LVM related PVCs are ready"
// +kubebuilder:printcolumn:name="age",type="date",JSONPath=".metadata.creationTimestamp",description="the age of this resource"
// +kubebuilder:resource:scope=Namespaced,categories={all,mpd},shortName=mv,singular=mpdlocalvolume,path=mpdlocalvolumes
// +kubebuilder:storageversion
// +kubebuilder:subresource:status
// +groupName:mdp

// MPDLocalVolume is the Schema for the mpdlocalvolumes API
type MPDLocalVolume struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MPDLocalVolumeSpec   `json:"spec,omitempty"`
	Status MPDLocalVolumeStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MPDLocalVolumeList contains a list of MPDLocalVolume
type MPDLocalVolumeList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MPDLocalVolume `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MPDLocalVolume{}, &MPDLocalVolumeList{})
}
