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


package workflow_shared

import (
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/define"
	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/wfimpl"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
)

func init() {
	smIns := wfimpl.GetSharedStorageClusterStateMachine()

	// 注册稳定态到非稳定态的转换检测及非稳定态的入口
	smIns.RegisterStateTranslateMainEnter(statemachine.StateInit, checkInstall, statemachine.StateCreating, installMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateInterrupt, checkRebuild, statemachine.StateRebuild, rebuildMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkRebuild, statemachine.StateRebuild, rebuildMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkModifyClass, define.WorkflowStateModifyClass, modifyClassMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkRestartCluster, define.WorkflowStateRestartCluster, restartClusterMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkRestartIns, define.WorkflowStateRestartIns, restartInsMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkFlushParams, define.WorkflowStateFlushParams, flushParamsMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkSwitchRw, define.WorkflowStateSwitchRw, switchRwMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkMigrateRo, define.WorkflowStateMigrateRo, migrateRoMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkMigrateRw, define.WorkflowStateMigrateRw, migrateRwMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkUpgradeMinorVersion, define.WorkflowStateUpgradeMinorVersion, upgradeMinorVersionMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkRebuildRo, define.WorkflowStateRebuildRo, rebuildRoMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkAddRo, define.WorkflowStateAddRo, addRoMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkRemoveRo, define.WorkflowStateRemoveRo, removeRoMainEnter)
	smIns.RegisterStateTranslateMainEnter(statemachine.StateRunning, checkExtendStorage, define.WorkflowStateExtendStorage, extendStorageMainEnter)

}
