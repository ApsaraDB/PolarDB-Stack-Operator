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

/**
 * @Description: SharedStorageCluster元数据存储
 */
type ISharedStorageClusterRepository interface {

	/**
	 * @Description: 获得所有SharedStorageCluster元数据
	 * @return []domain.SharedStorageCluster
	 * @return error
	 */
	GetAll() ([]*SharedStorageCluster, error)

	/**
	 * @Description: 根据名称获取SharedStorageCluster元数据
	 * @param name
	 * @param namespace
	 * @return *domain.SharedStorageCluster
	 */
	GetByName(name, namespace string) (*SharedStorageCluster, error)

	/**
	 * @Description: 根据视图模型获取SharedStorageCluster元数据
	 * @param data
	 * @return *domain.SharedStorageCluster
	 */
	GetByData(data interface{}, useModifyClass bool, useUpgradeVersion bool) *SharedStorageCluster

	/**
	 * @Description: 创建SharedStorageCluster
	 * @param *domain.SharedStorageCluster
	 * @return error
	 */
	Create(*SharedStorageCluster) error

	/**
	 * @Description: 更新SharedStorageCluster
	 * @param *domain.SharedStorageCluster
	 * @return error
	 */
	Update(*SharedStorageCluster) error

	/**
	 * @Description: 将SharedStorageCluster状态设置为Running
	 * @param name
	 * @param namespace
	 * @return error
	 */
	UpdateRunningStatus(name, namespace string) error

	/**
	 * @Description: 更新引擎状态
	 * @param name
	 * @param namespace
	 * @return error
	 */
	UpdateInsStatus(*SharedStorageCluster) error
}
