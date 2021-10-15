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


package domain

import (
	"errors"
	"strconv"
	"sync"

	"gitlab.alibaba-inc.com/polar-as/polar-common-domain/utils"
)

var fsIdConst = make([]string, 28)

var fsIdGeneratorMutex sync.Mutex

func init() {
	fsIdNumber := 5
	for i := 0; i < 21; i++ {
		fsIdConst[i] = strconv.Itoa(fsIdNumber)
		fsIdNumber += 1
	}
}

func FsIdGenerator(cluster *SharedStorageCluster) (string, error) {
	fsIdGeneratorMutex.Lock()
	defer fsIdGeneratorMutex.Unlock()

	dispatchedIds := getNowDispatchedFsId(cluster)

	for _, id := range fsIdConst {
		if !utils.ContainsString(dispatchedIds, id, nil) {
			return id, nil
		}
	}
	return "", errors.New("find all fsId dispatched, please check")
}

func getNowDispatchedFsId(cluster *SharedStorageCluster) []string {
	var dispatchedIds []string
	if cluster.RwIns != nil &&
		cluster.RwIns.StorageHostId != "" {
		dispatchedIds = append(dispatchedIds, cluster.RwIns.StorageHostId)
	}
	for _, ins := range cluster.RoInses {
		if ins.StorageHostId != "" {
			dispatchedIds = append(dispatchedIds, ins.StorageHostId)
		}
	}
	for _, ins := range cluster.TempRoInses {
		if ins.StorageHostId != "" {
			dispatchedIds = append(dispatchedIds, ins.StorageHostId)
		}
	}
	return dispatchedIds
}
