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
 * @Description: LocalStorageCluster元数据存储
 */
type ILocalStorageClusterRepository interface {

	/**
	 * @Description: 获得所有LocalStorageCluster元数据
	 * @return []domain.LocalStorageCluster
	 * @return error
	 */
	GetAll() ([]*LocalStorageCluster, error)

	/**
	 * @Description: 根据名称获取LocalStorageCluster元数据
	 * @param name
	 * @param namespace
	 * @return *domain.LocalStorageCluster
	 */
	GetByName(name, namespace string) (*LocalStorageCluster, error)

	/**
	 * @Description: 根据视图模型获取LocalStorageCluster元数据
	 * @param data
	 * @return *domain.LocalStorageCluster
	 */
	GetByData(data interface{}, useModifyClass bool, useUpgradeVersion bool) *LocalStorageCluster

	/**
	 * @Description: 创建LocalStorageCluster
	 * @param *domain.LocalStorageCluster
	 * @return error
	 */
	Create(*LocalStorageCluster) error

	/**
	 * @Description: 更新LocalStorageCluster
	 * @param *domain.LocalStorageCluster
	 * @return error
	 */
	Update(*LocalStorageCluster) error

	/**
	 * @Description: 将LocalStorageCluster状态设置为Running
	 * @param name
	 * @param namespace
	 * @return error
	 */
	UpdateRunningStatus(name, namespace string) error
}
