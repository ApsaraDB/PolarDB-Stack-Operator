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


package define

import (
	commondefine "github.com/ApsaraDB/PolarDB-Stack-Common/define"
	wfdefine "github.com/ApsaraDB/PolarDB-Stack-Workflow/define"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
)

const (
	WorkflowStateModifyClass         statemachine.State = "ModifyClass"
	WorkflowStateUpgradeMinorVersion statemachine.State = "UpgradeMinorVersion"
	WorkflowStateFlushParams         statemachine.State = "FlushParams"
	WorkflowStateRestartIns          statemachine.State = "RestartIns"
	WorkflowStateRestartCluster      statemachine.State = "RestartCluster"
	WorkflowStateSwitchRw            statemachine.State = "SwitchRw"
	WorkflowStateMigrateRw           statemachine.State = "MigrateRw"
	WorkflowStateMigrateRo           statemachine.State = "MigrateRo"
	WorkflowStateRebuildRo           statemachine.State = "RebuildRo"
	WorkflowStateAddRo               statemachine.State = "AddRo"
	WorkflowStateRemoveRo            statemachine.State = "RemoveRo"
	WorkflowStateExtendStorage       statemachine.State = "ExtendStorage"

	EnginePortRangesAnnotation = "polarbox.novip.engine.port.range"
	// Manager
	RwEnvServiceType            = "polardb_mpd_rw"
	RoEnvServiceType            = "polardb_mpd_ro"
	SharedStorageEnvStorageType = "fcsan"
	EngineImageName             = "engineImage"
	ManagerImageName            = "managerImage"
	PfsdImageName               = "pfsdImage"
	PfsdToolImageName           = "pfsdToolImage"
	ClusterManagerImageName     = "clusterManagerImage"

	AnnotationForceRebuild              = "forceRebuild"
	AnnotationMigrate                   = "migrate"
	AnnotationFlushParams               = "flushParams"
	AnnotationTempRoIds                 = "tempRoIds"
	AnnotationRestartIns                = "restartIns"
	AnnotationRestartCluster            = "restartCluster"
	AnnotationFlushParamsRestartCluster = "flushParamsRestartCluster"
	AnnotationSwitchRw                  = "switchRw"
	AnnotationRemoveIns                 = "removeIns"
	AnnotationExtendStorage             = "extendStorage"

	InsTypeRw string = "polardb_mpd_rw"
	InsTypeRo string = "polardb_mpd_ro"

	MpdClusterFinalizer = "mpd.finalizers.polardb.aliyun.com"
)

const (
	// 格式: "中断错误类型|导致中断的组件|中断原因"
	TempRoInsIdNotFound commondefine.InterruptErrorEnum = "TEMP_INS_ID_NOT_FOUND||temp ro instance id is not found."
	TempRoForRwMetaLost commondefine.InterruptErrorEnum = "TEMP_RO_FOR_RW_MEAT_LOST||"
	TempRoForRoMetaLost commondefine.InterruptErrorEnum = "TEMP_RO_FOR_RO_MEAT_LOST||"
	RoMetaInvalid       commondefine.InterruptErrorEnum = "RO_META_INVALID||ro metadata is invalid."
)

// 忽略以下模板参数的覆盖
var IgnoreTemplateParams []string = []string{"ssl", "ssl_cert_file", "ssl_key_file", "listen_addresses"}

// 忽略以下参数的更新
var IgnoreParams []string = []string{"log_line_prefix"}

var DefaultWfConf = map[wfdefine.WFConfKey]string{
	wfdefine.WorkFlowResourceName:      "_resourceName",
	wfdefine.WorkFlowResourceNameSpace: "_resourceNameSpace",
	wfdefine.WFInterruptReason:         "interrupt.reason",
	wfdefine.WFInterruptMessage:        "interrupt.message",
	wfdefine.WFInterruptPrevious:       "interrupt.previous.status",
	wfdefine.WFInterruptToRecover:      "interrupt.recover",
}
