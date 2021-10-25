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
	"fmt"
	"sync"

	"github.com/ApsaraDB/PolarDB-Stack-Operator/pkg/configuration"
	wfengineimpl "github.com/ApsaraDB/PolarDB-Stack-Workflow/implement/wfengine"
	"github.com/ApsaraDB/PolarDB-Stack-Workflow/wfengine"
)

var ResourceType = "shared"
var localResourceType = "local"

var (
	sharedStorageClusterWfOnce    sync.Once
	sharedStorageClusterWfManager *wfengine.WfManager
	localStorageClusterWfOnce     sync.Once
	localStorageClusterWfManager  *wfengine.WfManager
)

func GetSharedStorageClusterWfManager() *wfengine.WfManager {

	sharedStorageClusterWfOnce.Do(func() {
		if sharedStorageClusterWfManager == nil {
			var err error
			sharedStorageClusterWfManager, err = createWfManager(ResourceType, configuration.GetConfig().WorkFlowMetaDir)
			sharedStorageClusterWfManager.RegisterRecover(wfengineimpl.CreateDefaultRecover())
			if err != nil {
				panic(fmt.Sprintf("create %s wf manager failed: %v", ResourceType, err))
			}
		}
	})
	return sharedStorageClusterWfManager
}

func createWfManager(resourceType, workFlowMetaDir string) (wfManager *wfengine.WfManager, err error) {
	wfManager, err = wfengine.CreateWfManager(
		resourceType,
		workFlowMetaDir,
		wfengineimpl.CreateDefaultWfMetaLoader,
		wfengineimpl.CreateDefaultWorkflowHook,
		wfengineimpl.GetDefaultMementoStorageFactory(resourceType, false),
	)
	return
}

func GetLocalStorageClusterWfManager() *wfengine.WfManager {

	localStorageClusterWfOnce.Do(func() {
		if localStorageClusterWfManager == nil {
			var err error
			localStorageClusterWfManager, err = createWfManager(localResourceType, configuration.GetConfig().WorkFlowMetaDir)
			localStorageClusterWfManager.RegisterRecover(wfengineimpl.CreateDefaultRecover())
			if err != nil {
				panic(fmt.Sprintf("create %s wf manager failed: %v", ResourceType, err))
			}
		}
	})
	return localStorageClusterWfManager
}
