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
	"gitlab.alibaba-inc.com/polar-as/polar-mpd-controller/pkg/wfimpl"
	"os"
)

func isTesting() bool {
	if len(os.Args) > 1 && os.Args[1][:5] == "-test" {
		return true
	}
	return false
}

func init() {
	if !isTesting() {
		Register()
	}
}

func Register() {
	// 注册工作流步骤
	wfManager := wfimpl.GetSharedStorageClusterWfManager()
	wfManager.RegisterSteps(
		&InitMeta{},
		&PrepareStorage{},
		&CreateRwPod{},
		&CreateRoPods{},
		&CreateNetwork{},
		&AddToClusterManager{},
		&CreateClusterManager{},
		&UpdateRunningStatus{},
		&GenerateTempRoIds{},
		&InitTempRoMeta{},
		&CreateTempRoForRw{},
		&ConvertTempRoToRo{},
		&SwitchNewRoToRw{},
		&DeleteOldRw{},
		&UpdateModifyClassMeta{},
		&FlushParamsIfNecessary{},
		&DisableHA{},
		&EnableHA{},
		&EnsureNewRoUpToDate{},
		&CleanModifyClassTempMeta{},
		&EnsureCmRwAffinity{},
		&SaveParamsLastUpdateTime{},
		&RestartIns{},
		&RestartCluster{},
		&FlushParams{},
		&RestartClusterIfNeed{},
		&SwitchRw{},
		&GenerateMigrateTempId{},
		&EnsureRoMigrate{},
		&InitUpgradeImages{},
		&CleanUpgradeTempMeta{},
		&UpgradeCmVersion{},
		&CleanOldTempMeta{},
		&SetRebuildTag{},
		&CleanAllTempMeta{},
		&RemoveClusterManager{},
		&RemoveAllInsPod{},
		&CleanTempRoMeta{},
		&GenerateRebuildRoTempId{},
		&StopOldRo{},
		&GenerateAddRoTempId{},
		&RemoveRo{},
		&ExtendStorageFs{},
	)
}
