# 使用手册

### 创建 PVC

创建的 PVC 是作为您数据库集群的存储使用。

**字段说明：**

| 字段      | 字段解释                                                     |
| --------- | ------------------------------------------------------------ |
| IP        | 存储管理 pod 所在的机器 IP，如 10.0.0.77                     |
| name      | 所要创建的 PVC 名称，如 pvc-32ze341nncwlczm47bsre            |
| volume_id | 共享存储的 wwid，该盘须支持 scsi 或 nvme 协议，如 32ze341nncwlczm47bsre |

***示例：***

```shell
curl -X POST "http://${IP}:2002/pvcs" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"name\": \"${name}\", \"namespace\": \"default\", \"need_format\": true, \"volume_id\": \"${volume_id}\", \"volume_type\": \"lun\"}"
```

返回如下所示即为创建成功

```shell
{"workflow_id":"7cc5191d-803f-4ea5-8eb1-92e5437a52e9"}
```

创建成功后您可以查看 k8s 找到您刚刚创建的 PVC

```shell
kubectl get pvc 
```

### 创建实例集群

通过创建实例集群，您可以通过访问该集群访问到您的数据库。

***步骤：***

1. 根据您的需求配置实例集群的 yaml，并且通过 kubectl 使之生效。

```shell
kubectl apply -f your-cluster.yaml
```

字段说明：

| 字段                    | 字段解释                                                     |
| ----------------------- | ------------------------------------------------------------ |
| operatorName            | 您使用的 operator 名称，默认为 polar-mpd，可以在您的 operator 启动参数中使用 --filter-operator-name 参数配置 |
| dbClusterType           | 表示存储类型为共享存储                                       |
| followerNum             | 需要创建几个 ro 节点                                         |
| classInfo.className     | 实例规格名称，与您配置的 instance_level_config 中 classKey 对应，如 polar.o.x4.medium，您可以搜索 all.yaml 找到全部配置 |
| versionCfg.versionName  | 对应镜像配置 configmap postgresql-1-0-minor-version-info-rwo-${versionCfg.versionName} 的后缀，image-open 为 all.yaml 中自带的示例配置 |
| netCfg.engineStartPort  | 创建数据库服务所使用的端口                                   |
| shareStore.sharePvcName | 您已经创建好的 PVC 名称                                      |
| shareStore.volumeId     | 您 PVC 对应共享盘的 wwid，该盘须支持 scsi 或 nvme 协议       |
| shareStore.diskQuota    | 您的数据库引擎能使用的磁盘大小，单位是 M，由于 pfs 每 10G 为一个单位，每次扩容只能扩 10G 的倍数 |

***示例：***

```yaml
apiVersion: mpd.polardb.aliyun.com/v1
kind: MPDCluster
metadata:
  name: mpdcluster-sample-2
  namespace: default
spec:
  operatorName: polar-mpd
  dbClusterType: share
  followerNum: 1
  classInfo:
    className: polar.o.x4.medium
  classInfoModifyTo:
    className: ""
  versionCfg:
    versionName: image-open
  netCfg:
    engineStartPort: 5780
  shareStore:
    drive: "pvc"
    sharePvcNamespace: "default"
    sharePvcName: "pvc-32ze341nncwlczm47bsre"
    diskQuota: "300000"
    volumeId: "32ze341nncwlczm47bsre"
    volumeType: "multipath"
```

验证结果：

2. clusterStatus: Running 表示创建成功

```shell
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: Running
```

### 变配

通过变配来修改您的实例的规格。

***步骤：***

1. 修改您已经创建好的 cluster

```shell
kubectl edit mpdcluster your-cluster-name
```

2. 修改 classInfoModifyTo.className 字段为您想要变更成的规格名称。

字段说明：

| 字段                        | 字段解释                                                     |
| --------------------------- | ------------------------------------------------------------ |
| classInfoModifyTo.className | 变配后的规格名称，与您配置的 instance_level_config 中 classKey 对应 |

***示例：***

```yaml
spec:   
  classInfoModifyTo:
    className: polar.o.x4.large
```

***状态验证：***

- 变配过程中 clusterStatus 会变为 ModifyClass

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com mpdcluster-sample-1 -o yaml | grep clusterStatus
  clusterStatus: ModifyClass
```

- 变配完成后 clusterStatus 重新变为 Running

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com mpdcluster-sample-1 -o yaml | grep clusterStatus
  clusterStatus: Running
```

### 读写切换

将您数据库集群的 ro 节点切换为 rw 节点。

***步骤：***

1. 修改您已经创建好的 cluster。

```shell
kubectl edit mpdcluster your-cluster-name
```

2. 添加如下字段，字段说明：

| 字段                          | 字段解释                                                     |
| ----------------------------- | ------------------------------------------------------------ |
| metadata.annotations.switchRw | 需要切换为 rw 节点的 ro 节点 ID，ID 在 status.dbInstanceStatus 中列出 |

***示例：***

```yaml
metadata:
  annotations:
    switchRw: "7614"
```

***状态验证：***

正在切换中

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: SwitchRw
```

### 修改引擎参数（刷参）

修改您创建实例的引擎参数。

***步骤：***

1. 查看参数，参数保存在名称为 ${your-cluster-name}-user-params 的 configmap 中。

```plain
kubectl edit cm your-cluster-name-user-params
```

2. 编辑您想要修改的参数，参数取值范围见附录，如：

```plain
checkpoint_timeout: '{"name":"checkpoint_timeout","value":"300","updateTime":"2021-09-29T07:28:59Z","inoperative_value":""}'
```

3. 您可以修改 value 字段为您想要的值，同时更新 updateTime 为当前时间，由于参数更新会校验 updateTime，因此请务必更新 updateTime 字段。

4. 修改完成上述的 configmap，编辑您的 cluster 并设置刷参操作。

```shell
kubectl edit mpdcluster your-cluster-name
```

添加如下字段，字段说明：

| 字段                             | 字段说明                 |
| -------------------------------- | ------------------------ |
| metadata.annotations.flushParams | 设置为 true 表示需要刷参 |

***示例：***

```yaml
metadata:
  annotations:
    flushParams: "true"
```

### 实例重启

将您数据库集群中的某一个节点重启。

***步骤：***

1. 修改您已经创建好的 cluster 。

```shell
kubectl edit mpdcluster your-cluster-name
```

2. 添加如下字段，字段说明：

| 字段                            | 字段解释                                                     |
| ------------------------------- | ------------------------------------------------------------ |
| metadata.annotations.restartIns | 您想要重启的数据库实例 ID，ID 在 status.dbInstanceStatus 中列出，为一个数字 |

***示例：***

```yaml
metadata:
  annotations:
    restartIns: "7503"
```

### 集群重启

将您数据库集群中的某一个节点重启。

***步骤：***

1. 修改您已经创建好的 cluster

```shell
kubectl edit mpdcluster your-cluster-name
```

2. 添加如下字段，字段说明：

| 字段                                | 字段解释                     |
| ----------------------------------- | ---------------------------- |
| metadata.annotations.restartCluster | 设置为 true 表示需要重启实例 |

示例：

```yaml
metadata:
  annotations:
    restartCluster: "true"
```

### 小版本升级

小版本升级将您的数据库集群所使用的镜像版本升级。

***步骤：***

1. 创建一个新的镜像配置，内容修改为您需要升级的镜像版本，使用 kubectl 使之生效。

字段说明：

| 字段           | 字段解释       |
| -------------- | -------------- |
| pgEngineImage  | DB 引擎镜像    |
| pgManagerImage | manager 镜像   |
| pfsdImage      | pfsd 镜像      |
| pfsdToolImage  | pfsd tool 镜像 |

```yaml
apiVersion: v1
data:
  name: image-open
  pfsdImage: polardb/pfsd:1.2.41-20211018
  pfsdToolImage: polardb/pfsd_tool:1.2.41-20211018
  pgClusterManagerImage: polardb/polardb-cluster-manager:latest
  pgEngineImage: polardb/polardb_pg_engine_release:11beta2.20210910.d558886c.20211018195123
  pgManagerImage: polardb/polardb_pg_manager:20211018195123.9ae43314
kind: ConfigMap
metadata:
  labels:
    configtype: minor_version_info
    dbClusterMode: WriteReadMore
    dbType: PostgreSQL
    dbVersion: "1.0"
  name: postgresql-1-0-minor-version-info-rwo-image-open
  namespace: kube-system
```

2. 修改您已经创建好的 cluster

```shell
kubectl edit mpdcluster your-cluster-name
```

3. 修改 classInfoModifyTo.className 字段为您想要变更成的规格名称，添加如下字段，字段说明：

| 字段                           | 字段解释                                                     |
| ------------------------------ | ------------------------------------------------------------ |
| versionCfgModifyTo.versionName | 变配后的规格名称，与您配置的 instance_level_config 中 classKey 对应 |

***示例：***

```yaml
spec: 
  versionCfgModifyTo:
    versionName: your-new-config-name
```

### 完整重建

重建您的数据库集群。

***步骤：***

1. 修改您已经创建好的 cluster

```shell
kubectl edit mpdcluster your-cluster-name
```

2. 添加如下字段，字段说明：

| 字段                              | 字段说明                     |
| --------------------------------- | ---------------------------- |
| metadata.annotations.forceRebuild | 设置为 true 表示需要重建集群 |

***示例：***

```yaml
metadata:
  annotations:
    forceRebuild: "true"
```

***状态检查：***

正在重建中

```shell
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: Rebuild
```

重建完成

```shell
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: Running
```

### 添加 RO

增加一个 ro 节点。

**注意：**需要保证您有空余的机器，否则会端口冲突。

***步骤：***

1. 修改您已经创建好的 cluster

```shell
kubectl edit mpdcluster your-cluster-name
```

2. 添加字段，字段说明：

| 字段        | 字段说明             |
| ----------- | -------------------- |
| followerNum | 添加后总共的 ro 数量 |

***示例：***

```yaml
spec:
  followerNum: 2
```

### RW 迁移

将您的 rw 节点迁移到另外一台机器上。

***注意：***节点迁移只能迁移到没有集群实例的节点，否则会发生端口冲突。

***步骤：***

1. 修改您已经创建好的 cluster

```shell
kubectl edit mpdcluster your-cluster-name
```

2. 添加如下字段，字段说明：

| 字段                         | 字段解释                                                     |
| ---------------------------- | ------------------------------------------------------------ |
| metadata.annotations.migrate | 形式为 ${insId}\|${targetNode}，如 7601\|dbm-02，insId 表示您当前 rw 节点 ID，targetNode 表示您想要迁移到的机器名，可以通过 kubectl get node 获得 |

***示例：***

```yaml
metadata:
  annotations:
    migrate: 7601|dbm-02
```

***状态验证：***

迁移中

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: MigrateRw
```

迁移完成

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: Running
```

### RO 迁移

将您的 ro 节点迁移到另外一台机器上。

**注意：**节点迁移只能迁移到没有集群实例的节点，否则会发生端口冲突。

***步骤：***

1. 修改您已经创建好的 cluster

```shell
kubectl edit mpdcluster your-cluster-name
```

2. 添加如下字段，字段说明：

| 字段                         | 字段说明                                                     |
| ---------------------------- | ------------------------------------------------------------ |
| metadata.annotations.migrate | 形式为 ${insId}\|${targetNode}，如 7601\|dbm-02，insId 表示您迁移的 ro 节点 ID，targetNode 表示您想要迁移到的机器名，可以通过 kubectl get node 获得 |

***示例：***

```yaml
metadata:
  annotations:
    migrate: 7601|dbm-02
```

***状态验证：***

迁移中

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: MigrateRo
```

迁移完成

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: Running
```

### 存储扩容

扩容您数据库使用的磁盘。

执行前，您需要先将您的磁盘扩容。

修改您已经创建好的 cluster

```shell
kubectl edit mpdcluster your-cluster-name
```

添加如下字段，字段说明：

| 字段                               | 字段说明                 |
| ---------------------------------- | ------------------------ |
| metadata.annotations.extendStorage | 设置为 true 表示扩展存储 |

修改如下字段为您扩容后想要的大小，字段说明：

| 字段                      | 字段说明                                                     |
| ------------------------- | ------------------------------------------------------------ |
| spec.shareStore.diskQuota | 您的数据库引擎能使用的磁盘大小，单位是 M，由于 pfs 每 10G 为一个单位，每次扩容只能扩 10G 的倍数 |

示例：

```yaml
metadata:
  annotations:
    extendStorage: "true"
```

### 失败重试

适用于操作失败状态变为 Interrupt

将 interrupt.recover 的值由 F 改为 T

```yaml
metadata:
  annotations:
    interrupt.recover: F
```

### 删除实例集群

请谨慎操作

```yaml
kubectl delete mpdcluster your-cluster-name
```

### 附录

引擎参数说明

| 参数名称                                         | 默认值     | 重启 | 修改范围                                                     |
| ------------------------------------------------ | ---------- | ---- | ------------------------------------------------------------ |
| archive_mode                                     | on         | 是   | [always\|on\|off]                                            |
| autovacuum_vacuum_cost_delay                     | 0          | 否   | [-1-100]                                                     |
| autovacuum_vacuum_cost_limit                     | 10000      | 否   | [-1-10000]                                                   |
| auto_explain.log_analyze                         | off        | 否   | [on\|off]                                                    |
| auto_explain.log_buffers                         | off        | 否   | [on\|off]                                                    |
| auto_explain.log_format                          | text       | 否   | [text\|xml\|json\|yaml]                                      |
| auto_explain.log_min_duration                    | 5000       | 否   | [-1-2147483647]                                              |
| auto_explain.log_nested_statements               | off        | 否   | [on\|off]                                                    |
| auto_explain.log_timing                          | on         | 否   | [on\|off]                                                    |
| auto_explain.log_triggers                        | off        | 否   | [on\|off]                                                    |
| auto_explain.log_verbose                         | off        | 否   | [on\|off]                                                    |
| auto_explain.sample_rate                         | 1          | 否   | [0-1]                                                        |
| checkpoint_completion_target                     | 0.9        | 否   | [0-1]                                                        |
| checkpoint_timeout                               | 300        | 否   | [1-86400]                                                    |
| default_transaction_deferrable                   | off        | 否   | [on\|off]                                                    |
| default_with_oids                                | off        | 否   | [on\|off]                                                    |
| default_with_rowids                              | off        | 否   | [on\|off]                                                    |
| enable_partitionwise_aggregate                   | on         | 否   | [on\|off]                                                    |
| enable_partitionwise_join                        | on         | 否   | [on\|off]                                                    |
| extra_float_digits                               | 0          | 否   | [-15-3]                                                      |
| idle_in_transaction_session_timeout              | 3600000    | 否   | ^(0\|[1-9]\d{3,8}\|1\d{9}\|2000000000)$                      |
| jit                                              | off        | 否   | [on\|off]                                                    |
| lock_timeout                                     | 0          | 否   | ^(0\|[1-9]\d{3,8}\|1\d{9}\|2000000000)$                      |
| log_min_duration_statement                       | -1         | 否   | [-1-2147483647]                                              |
| log_statement                                    | all        | 否   | [none\|ddl\|mod\|all]                                        |
| log_temp_files                                   | 100000     | 否   | [-1-2147483647]                                              |
| log_timezone                                     | 'PRC'      | 否   | ^'(((UTC)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|((GMT)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|CST6CDT\|Poland\|Kwajalein\|MST\|NZ\|Universal\|Libya\|Turkey\|EST5EDT\|Greenwich\|NZ-CHAT\|MET\|Portugal\|GMT-0\|CET\|Eire\|PST8PDT\|Jamaica\|GMT\|Zulu\|Japan\|ROC\|GB-Eire\|ROK\|Navajo\|Singapore\|posixrules\|GB\|EST\|GMT0\|Hongkong\|PRC\|Iran\|MST7MDT\|WET\|W-SU\|UCT\|Cuba\|Egypt\|EET\|Israel\|UTC\|HST\|Iceland)'$ |
| max_parallel_maintenance_workers                 | 2          | 否   | [0-1024]                                                     |
| max_parallel_workers                             | 2          | 否   | [0-512]                                                      |
| max_parallel_workers_per_gather                  | 2          | 否   | [0-512]                                                      |
| min_parallel_index_scan_size                     | 64         | 否   | [0-715827882]                                                |
| min_parallel_table_scan_size                     | 1024       | 否   | [0-715827882]                                                |
| old_snapshot_threshold                           | -1         | 是   | [-1-86400]                                                   |
| polar_comp_dynatune                              | 0          | 是   | [0-100]                                                      |
| polar_comp_dynatune_profile                      | oltp       | 是   | [oltp\|reporting\|mixed]                                     |
| polar_comp_enable_pruning                        | on         | 否   | [on\|off]                                                    |
| polar_comp_redwood_date                          | on         | 否   | [on\|off]                                                    |
| polar_comp_redwood_greatest_least                | on         | 否   | [on\|off]                                                    |
| polar_comp_redwood_raw_names                     | on         | 否   | [on\|off]                                                    |
| polar_comp_redwood_strings                       | on         | 否   | [on\|off]                                                    |
| polar_comp_stmt_level_tx                         | on         | 否   | [on\|off]                                                    |
| dbms_job.database_name                           | 'postgres' | 是   | ^'\w+'$                                                      |
| statement_timeout                                | 0          | 否   | ^(0\|[1-9]\d{3,8}\|1\d{9}\|2000000000)$                      |
| temp_file_limit                                  | 524288000  | 否   | [-1-1048576000]                                              |
| timezone                                         | 'PRC'      | 否   | ^'(((UTC)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|((GMT)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|CST6CDT\|Poland\|Kwajalein\|MST\|NZ\|Universal\|Libya\|Turkey\|EST5EDT\|Greenwich\|NZ-CHAT\|MET\|Portugal\|GMT-0\|CET\|Eire\|PST8PDT\|Jamaica\|GMT\|Zulu\|Japan\|ROC\|GB-Eire\|ROK\|Navajo\|Singapore\|posixrules\|GB\|EST\|GMT0\|Hongkong\|PRC\|Iran\|MST7MDT\|WET\|W-SU\|UCT\|Cuba\|Egypt\|EET\|Israel\|UTC\|HST\|Iceland)'$ |
| track_commit_timestamp                           | off        | 是   | [on\|off]                                                    |
| vacuum_defer_cleanup_age                         | 0          | 否   | [0-1000000]                                                  |
| wal_level                                        | logical    | 是   | [replica\|logical]                                           |
| work_mem                                         | 4096       | 否   | [4096-524288]                                                |
| polar_empty_string_is_null_enable                | on         | 否   | [on\|off]                                                    |
| polar_enable_varchar2_length_with_byte           | on         | 否   | [on\|off]                                                    |
| polar_enable_base64_decode                       | on         | 否   | [on\|off]                                                    |
| polar_enable_nls_date_format                     | on         | 否   | [on\|off]                                                    |
| polar_enable_rowid                               | on         | 否   | [on\|off]                                                    |
| postgres_fdw.polar_connection_check              | off        | 否   | [on\|off]                                                    |
| polar_comp_custom_plan_tries                     | 5          | 否   | [-1-100]                                                     |
| dblink.polar_auto_port_mapping                   | off        | 否   | [on\|off]                                                    |
| dblink.polar_connection_check                    | off        | 否   | [on\|off]                                                    |
| pg_stat_statements.enable_superuser_track        | on         | 否   | [on\|off]                                                    |
| cron.polar_allow_superuser_task                  | on         | 是   | [on\|off]                                                    |
| polar_stat_sql.enable_qps_monitor                | on         | 否   | [on\|off]                                                    |
| polar_enable_audit_log_bind_sql_parameter        | off        | 否   | [on\|off]                                                    |
| polar_enable_audit_log_bind_sql_parameter_new    | off        | 否   | [on\|off]                                                    |
| polar_enable_replica_use_smgr_cache              | on         | 是   | [off\|on]                                                    |
| idle_session_timeout                             | 0          | 否   | [0-2147483647]                                               |
| polar_resource_group.total_mem_limit_rate        | 95         | 否   | [50-100]                                                     |
| polar_resource_group.total_mem_limit_remain_size | 524288     | 否   | [131072-2097151]                                             |



### 附录

引擎参数说明

| 参数名称                                         | 默认值     | 重启 | 修改范围                                                     |
| ------------------------------------------------ | ---------- | ---- | ------------------------------------------------------------ |
| archive_mode                                     | on         | 是   | [always\|on\|off]                                            |
| autovacuum_vacuum_cost_delay                     | 0          | 否   | [-1-100]                                                     |
| autovacuum_vacuum_cost_limit                     | 10000      | 否   | [-1-10000]                                                   |
| auto_explain.log_analyze                         | off        | 否   | [on\|off]                                                    |
| auto_explain.log_buffers                         | off        | 否   | [on\|off]                                                    |
| auto_explain.log_format                          | text       | 否   | [text\|xml\|json\|yaml]                                      |
| auto_explain.log_min_duration                    | 5000       | 否   | [-1-2147483647]                                              |
| auto_explain.log_nested_statements               | off        | 否   | [on\|off]                                                    |
| auto_explain.log_timing                          | on         | 否   | [on\|off]                                                    |
| auto_explain.log_triggers                        | off        | 否   | [on\|off]                                                    |
| auto_explain.log_verbose                         | off        | 否   | [on\|off]                                                    |
| auto_explain.sample_rate                         | 1          | 否   | [0-1]                                                        |
| checkpoint_completion_target                     | 0.9        | 否   | [0-1]                                                        |
| checkpoint_timeout                               | 300        | 否   | [1-86400]                                                    |
| default_transaction_deferrable                   | off        | 否   | [on\|off]                                                    |
| default_with_oids                                | off        | 否   | [on\|off]                                                    |
| default_with_rowids                              | off        | 否   | [on\|off]                                                    |
| enable_partitionwise_aggregate                   | on         | 否   | [on\|off]                                                    |
| enable_partitionwise_join                        | on         | 否   | [on\|off]                                                    |
| extra_float_digits                               | 0          | 否   | [-15-3]                                                      |
| idle_in_transaction_session_timeout              | 3600000    | 否   | ^(0\|[1-9]\d{3,8}\|1\d{9}\|2000000000)$                      |
| jit                                              | off        | 否   | [on\|off]                                                    |
| lock_timeout                                     | 0          | 否   | ^(0\|[1-9]\d{3,8}\|1\d{9}\|2000000000)$                      |
| log_min_duration_statement                       | -1         | 否   | [-1-2147483647]                                              |
| log_statement                                    | all        | 否   | [none\|ddl\|mod\|all]                                        |
| log_temp_files                                   | 100000     | 否   | [-1-2147483647]                                              |
| log_timezone                                     | 'PRC'      | 否   | ^'(((UTC)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|((GMT)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|CST6CDT\|Poland\|Kwajalein\|MST\|NZ\|Universal\|Libya\|Turkey\|EST5EDT\|Greenwich\|NZ-CHAT\|MET\|Portugal\|GMT-0\|CET\|Eire\|PST8PDT\|Jamaica\|GMT\|Zulu\|Japan\|ROC\|GB-Eire\|ROK\|Navajo\|Singapore\|posixrules\|GB\|EST\|GMT0\|Hongkong\|PRC\|Iran\|MST7MDT\|WET\|W-SU\|UCT\|Cuba\|Egypt\|EET\|Israel\|UTC\|HST\|Iceland)'$ |
| max_parallel_maintenance_workers                 | 2          | 否   | [0-1024]                                                     |
| max_parallel_workers                             | 2          | 否   | [0-512]                                                      |
| max_parallel_workers_per_gather                  | 2          | 否   | [0-512]                                                      |
| min_parallel_index_scan_size                     | 64         | 否   | [0-715827882]                                                |
| min_parallel_table_scan_size                     | 1024       | 否   | [0-715827882]                                                |
| old_snapshot_threshold                           | -1         | 是   | [-1-86400]                                                   |
| polar_comp_dynatune                              | 0          | 是   | [0-100]                                                      |
| polar_comp_dynatune_profile                      | oltp       | 是   | [oltp\|reporting\|mixed]                                     |
| polar_comp_enable_pruning                        | on         | 否   | [on\|off]                                                    |
| polar_comp_redwood_date                          | on         | 否   | [on\|off]                                                    |
| polar_comp_redwood_greatest_least                | on         | 否   | [on\|off]                                                    |
| polar_comp_redwood_raw_names                     | on         | 否   | [on\|off]                                                    |
| polar_comp_redwood_strings                       | on         | 否   | [on\|off]                                                    |
| polar_comp_stmt_level_tx                         | on         | 否   | [on\|off]                                                    |
| dbms_job.database_name                           | 'postgres' | 是   | ^'\w+'$                                                      |
| statement_timeout                                | 0          | 否   | ^(0\|[1-9]\d{3,8}\|1\d{9}\|2000000000)$                      |
| temp_file_limit                                  | 524288000  | 否   | [-1-1048576000]                                              |
| timezone                                         | 'PRC'      | 否   | ^'(((UTC)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|((GMT)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|CST6CDT\|Poland\|Kwajalein\|MST\|NZ\|Universal\|Libya\|Turkey\|EST5EDT\|Greenwich\|NZ-CHAT\|MET\|Portugal\|GMT-0\|CET\|Eire\|PST8PDT\|Jamaica\|GMT\|Zulu\|Japan\|ROC\|GB-Eire\|ROK\|Navajo\|Singapore\|posixrules\|GB\|EST\|GMT0\|Hongkong\|PRC\|Iran\|MST7MDT\|WET\|W-SU\|UCT\|Cuba\|Egypt\|EET\|Israel\|UTC\|HST\|Iceland)'$ |
| track_commit_timestamp                           | off        | 是   | [on\|off]                                                    |
| vacuum_defer_cleanup_age                         | 0          | 否   | [0-1000000]                                                  |
| wal_level                                        | logical    | 是   | [replica\|logical]                                           |
| work_mem                                         | 4096       | 否   | [4096-524288]                                                |
| polar_empty_string_is_null_enable                | on         | 否   | [on\|off]                                                    |
| polar_enable_varchar2_length_with_byte           | on         | 否   | [on\|off]                                                    |
| polar_enable_base64_decode                       | on         | 否   | [on\|off]                                                    |
| polar_enable_nls_date_format                     | on         | 否   | [on\|off]                                                    |
| polar_enable_rowid                               | on         | 否   | [on\|off]                                                    |
| postgres_fdw.polar_connection_check              | off        | 否   | [on\|off]                                                    |
| polar_comp_custom_plan_tries                     | 5          | 否   | [-1-100]                                                     |
| dblink.polar_auto_port_mapping                   | off        | 否   | [on\|off]                                                    |
| dblink.polar_connection_check                    | off        | 否   | [on\|off]                                                    |
| pg_stat_statements.enable_superuser_track        | on         | 否   | [on\|off]                                                    |
| cron.polar_allow_superuser_task                  | on         | 是   | [on\|off]                                                    |
| polar_stat_sql.enable_qps_monitor                | on         | 否   | [on\|off]                                                    |
| polar_enable_audit_log_bind_sql_parameter        | off        | 否   | [on\|off]                                                    |
| polar_enable_audit_log_bind_sql_parameter_new    | off        | 否   | [on\|off]                                                    |
| polar_enable_replica_use_smgr_cache              | on         | 是   | [off\|on]                                                    |
| idle_session_timeout                             | 0          | 否   | [0-2147483647]                                               |
| polar_resource_group.total_mem_limit_rate        | 95         | 否   | [50-100]                                                     |
| polar_resource_group.total_mem_limit_remain_size | 524288     | 否   | [131072-2097151]                                             |



以下参数为高危参数，请确认影响后谨慎修改

| 参数名称                                     | 默认值                                                       | 重启 | 修改范围                            |
| -------------------------------------------- | ------------------------------------------------------------ | ---- | ----------------------------------- |
| autovacuum_max_workers                       | 5                                                            | 是   | [1-262143]                          |
| autovacuum_work_mem                          | 200MB                                                        | 否   | [-1-2147483647]                     |
| enable_hashagg                               | off                                                          | 否   | [on\|off]                           |
| log_connections                              | off                                                          | 否   | [on\|off]                           |
| log_disconnections                           | off                                                          | 否   | [on\|off]                           |
| max_standby_streaming_delay                  | 30000                                                        | 否   | [-1-2147483647]                     |
| polar_bgwriter_batch_size_flushlist          | 100                                                          | 是   | [1-10000]                           |
| polar_bgwriter_max_batch_size                | 5000                                                         | 否   | [0-1073741823]                      |
| polar_bgwriter_sleep_lsn_lag                 | 100                                                          | 否   | [0-1000]                            |
| polar_buffer_copy_lsn_lag_with_cons_lsn      | 100                                                          | 否   | [1-1000]                            |
| polar_buffer_copy_min_modified_count         | 5                                                            | 否   | [0-100]                             |
| polar_bulk_extend_size                       | 512                                                          | 否   | [0-1073741823]                      |
| polar_check_checkpoint_legal_interval        | 10                                                           | 否   | [1-3600]                            |
| polar_clog_slot_size                         | 512                                                          | 是   | [128-8192]                          |
| polar_comp_early_lock_release                | off                                                          | 否   | [on\|off]                           |
| polar_copy_buffers                           | 16384                                                        | 是   | [128-1073741823]                    |
| polar_enable_connectby_multiexpr             | on                                                           | 否   | [on\|off]                           |
| polar_enable_physical_repl_non_super_wal_snd | off                                                          | 否   | [on\|off]                           |
| polar_enable_show_polar_comp_version         | on                                                           | 否   | [on\|off]                           |
| polar_enable_syslog_pipe_buffer              | off                                                          | 否   | [on\|off]                           |
| polar_global_temp_table_debug                | off                                                          | 否   | [on\|off]                           |
| polar_max_log_files                          | 20                                                           | 否   | [-1-2147483647]                     |
| polar_max_super_conns                        | 100                                                          | 否   | [-1-262143]                         |
| polar_num_active_global_temp_table           | 1000                                                         | 是   | [0-1000000]                         |
| polar_parallel_bgwriter_check_interval       | 10                                                           | 否   | [1-600]                             |
| polar_parallel_bgwriter_delay                | 10                                                           | 否   | [1-10000]                           |
| polar_parallel_bgwriter_workers              | 5                                                            | 否   | [0-8]                               |
| polar_parallel_new_bgwriter_threshold_time   | 10                                                           | 否   | [1-3600]                            |
| polar_read_ahead_xlog_num                    | 200                                                          | 是   | [0-200]                             |
| polar_redo_hashtable_size                    | 131072                                                       | 是   | [16-1073741823]                     |
| polar_ring_buffer_vacuum                     | 128                                                          | 否   | [10-1000]                           |
| polar_spl_savepoint_enable                   | on                                                           | 否   | [on\|off]                           |
| polar_temp_relation_file_in_shared_storage   | off                                                          | 是   | [on\|off]                           |
| polar_vfs.libmm_num_partition                | 32                                                           | 是   | [4-128]                             |
| polar_vfs.libmm_size                         | 67108864                                                     | 是   | [67108864-1073741824]               |
| polar_enable_fullpage_snapshot               | off                                                          | 否   | [on\|off]                           |
| polar_enable_default_polar_comp_guc          | on                                                           | 否   | [on\|off]                           |
| polar_stat_stale_cost                        | 0.0001                                                       | 否   | [0-2147483647]                      |
| polar_auditlog_max_query_length              | 4096                                                         | 否   | [512-49152]                         |
| polar_use_statistical_relpages               | on                                                           | 否   | [on\|off]                           |
| shared_preload_libraries                     | 'polar_vfs,polar_worker,dbms_pipe,polar_gen,pg_stat_statements,auth_delay,auto_explain,pg_cron,dbms_job,polar_stat_sql' | 是   | ^'(polar_vfs,polar_worker)(,\w+)*'$ |
| wal_keep_segments                            | 32                                                           | 否   | [0-100000]                          |
| temp_tablespaces                             | 'polar_tmp'                                                  | 否   |                                     |
| polar_csn_enable                             | on                                                           | 是   | [on\|off]                           |
| polar_max_non_super_conns                    | 400                                                          | 否   | [-1-262143]                         |
| polar_enable_master_xlog_read_ahead          | on                                                           | 是   | [on\|off]                           |
| polar_enable_early_launch_checkpointer       | on                                                           | 是   | [on\|off]                           |
| polar_enable_early_launch_parallel_bgwriter  | on                                                           | 是   | [on\|off]                           |
| polar_enable_wal_prefetch                    | off                                                          | 是   | [on\|off]                           |
| polar_enable_persisted_buffer_pool           | on                                                           | 是   | [on\|off]                           |
| polar_cast_decode_list                       | ''                                                           | 否   |                                     |