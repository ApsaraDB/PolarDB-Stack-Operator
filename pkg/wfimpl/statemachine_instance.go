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
	"sync"

	"github.com/ApsaraDB/PolarDB-Stack-Workflow/statemachine"
)

var (
	sharedStorageClusterSmOnce       sync.Once
	sharedStorageClusterStateMachine *statemachine.StateMachine
	localStorageClusterSmOnce        sync.Once
	localStorageClusterStateMachine  *statemachine.StateMachine
)

func GetSharedStorageClusterStateMachine() *statemachine.StateMachine {
	sharedStorageClusterSmOnce.Do(func() {
		if sharedStorageClusterStateMachine == nil {
			sharedStorageClusterStateMachine = statemachine.CreateStateMachineInstance(ResourceType)
			sharedStorageClusterStateMachine.RegisterStableState(statemachine.StateRunning, statemachine.StateInterrupt, statemachine.StateInit)
		}
	})
	return sharedStorageClusterStateMachine
}

func GetLocalStorageClusterStateMachine() *statemachine.StateMachine {
	localStorageClusterSmOnce.Do(func() {
		if localStorageClusterStateMachine == nil {
			localStorageClusterStateMachine = statemachine.CreateStateMachineInstance(localResourceType)
			localStorageClusterStateMachine.RegisterStableState(statemachine.StateRunning, statemachine.StateInterrupt, statemachine.StateInit)
		}
	})
	return localStorageClusterStateMachine
}
