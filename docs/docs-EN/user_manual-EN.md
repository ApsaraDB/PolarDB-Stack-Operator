# User Guide

### Create a PVC

The created PVC is used as the storage for the database cluster.

**Field Description:**

| Field     | Description                                                  |
| --------- | ------------------------------------------------------------ |
| IP        | IP address of the machine where the storage management pod is located, e.g., 10.0.0.77. |
| name      | Name of the PVC to be created, e.g., pvc-32ze341nncwlczm47bsre |
| volume_id | wwid of the shared storage, e.g., 32ze341nncwlczm47bsre. This disk should support the SCSI or NVMe protocol. |

***Example:***

```shell
curl -X POST "http://${IP}:2002/pvcs" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"name\": \"${name}\", \"namespace\": \"default\", \"need_format\": true, \"volume_id\": \"${volume_id}\", \"volume_type\": \"lun\"}"
```

If the following information is returned, it indicates that the PVC is created successfully.

```shell
{"workflow_id":"7cc5191d-803f-4ea5-8eb1-92e5437a52e9"}
```

After the PVC is created, you can view the created PVC using the following command in Kubernetes:

```shell
kubectl get pvc 
```

### Create an Instance Cluster

Create an instance cluster. You can access your database by accessing this cluster.

***Steps:***

1. Configure the YAML file of the instance cluster according to your actual needs and execute the command `kubectl` to make it take effect.

```shell
kubectl apply -f your-cluster.yaml
```

Field Description:

| Field                   | Field Description                                            |
| ----------------------- | ------------------------------------------------------------ |
| operatorName            | The operator name you are using. By default it is polar-mpd and you can configure it using the parameter `--filter-operator-name` in the operator startup parameters. |
| dbClusterType           | The storage type is shared storage.                          |
| followerNum             | Number of ro nodes that need to be created.                  |
| classInfo.className     | Name of the instance's specification, which corresponds to the classKey in instance_level_config you have configured, e.g., polar.o.x4.medium. You can search all.yaml to find all configurations. |
| versionCfg.versionName  | The suffix of the corresponding image configuration configmap postgresql-1-0-minor-version-info-rwo-${versionCfg.versionName}. The image-open is the default configuration example in all.yaml. |
| netCfg.engineStartPort  | Port used to create database services.                       |
| shareStore.sharePvcName | Name of the created PVC.                                     |
| shareStore.volumeId     | The wwid of the corresponding shared disk of your PVC. This disk MUST support SCSI or NVMe protocol. |
| shareStore.diskQuota    | The disk size that can be used by your database engine, unit: MB. As PFS uses 10 GB as a unit, only a multiple of 10 GB is allowed for increasing the capacity each time. |

***Example:***

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

Verification result:

2. clusterStatus: Running indicates that the cluster is created successfully.

```shell
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: Running
```

### Change Specifications

Change the specifications of your instance.

***Steps:***

1. Modify the created cluster.

```shell
kubectl edit mpdcluster your-cluster-name
```

2. Modify the field classInfoModifyTo.className to the name of the specification you want to change the current one to.

Field Description:

| Field                       | Description                                                  |
| --------------------------- | ------------------------------------------------------------ |
| classInfoModifyTo.className | Name of the specification after change, which corresponds to classKey in instance_level_config you have configured. |

***Example:***

```yaml
spec:   
  classInfoModifyTo:
    className: polar.o.x4.large
```

***Verify the status:***

- During the process of changing specifications, the clusterStatus will be changed to ModifyClass.

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com mpdcluster-sample-1 -o yaml | grep clusterStatus
  clusterStatus: ModifyClass
```

- After changing specifications is completed, the clusterStatus will be changed to Running again.

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com mpdcluster-sample-1 -o yaml | grep clusterStatus
  clusterStatus: Running
```

### Switch between the RW Node and RO Nodes

Switch a ro node to the rw node in your database cluster.

***Steps:***

1. Modify the created cluster.

```shell
kubectl edit mpdcluster your-cluster-name
```

2. Add the following field whose description is shown in the table below.

| Field                         | Field Description                                            |
| ----------------------------- | ------------------------------------------------------------ |
| metadata.annotations.switchRw | The ID of the ro node to be switched to the rw node. The ID is listed in status.dbInstanceStatus. |

***Example:***

```yaml
metadata:
  annotations:
    switchRw: "7614"
```

***Verify the status:***

Switching

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: SwitchRw
```

### Modify Engine Parameters

Modify the engine parameters of the created instance.

***Steps:***

1. View the parameters which are saved in the configmap named ${your-cluster-name}-user-params.

```plain
kubectl edit cm your-cluster-name-user-params
```

2. Edit the parameters you want to modify. For the value range of each parameter, refer to the Appendix. For example: 

```plain
checkpoint_timeout: '{"name":"checkpoint_timeout","value":"300","updateTime":"2021-09-29T07:28:59Z","inoperative_value":""}'
```

3. You can modify the value of the field `value` to the one you need and change the value of the field `updateTime` to the current time. As updating parameters requires verifying `updateTime`, remember you MUST update the value of the field `updateTime`.

4. After modifying the above configmap, edit your cluster and set the parameter modification operation.

```shell
kubectl edit mpdcluster your-cluster-name
```

Add the following field whose description is shown in the table below.

| Field                            | Field Description                                            |
| -------------------------------- | ------------------------------------------------------------ |
| metadata.annotations.flushParams | If this field is set to true, it indicates that modifying parameters is required. |

***Example:***

```yaml
metadata:
  annotations:
    flushParams: "true"
```

### Restart an Instance

You can restart a specific node in the database cluster.

***Steps:***

1. Modify the created cluster.

```shell
kubectl edit mpdcluster your-cluster-name
```

2. Add the following field whose description is shown in the table below.

| Field                           | Field Description                                            |
| ------------------------------- | ------------------------------------------------------------ |
| metadata.annotations.restartIns | The ID of the database instance you want to restart. The ID is listed in status.dbInstanceStatus as a number. |

***Example:***

```yaml
metadata:
  annotations:
    restartIns: "7503"
```

### Restart a Cluster

Restart the entire database cluster.

***Steps:***

1. Modify the created cluster.

```shell
kubectl edit mpdcluster your-cluster-name
```

2. Add the following field whose description is shown in the table below.

| Field                               | Field Description                                            |
| ----------------------------------- | ------------------------------------------------------------ |
| metadata.annotations.restartCluster | If this field is set to true, it indicates that restarting the cluster is required. |

**Example:**

```yaml
metadata:
  annotations:
    restartCluster: "true"
```

### Update the Image Version

Modify configmap postgresql-1-0-minor-version-info-rwo-image-open.

```bash
kubectl -n kube-system edit configmap postgresql-1-0-minor-version-info-rwo-image-open
```

Update the image version that you need to modify as the example below:

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

Field Description:

| Field          | Field Description             |
| -------------- | ----------------------------- |
| pgEngineImage  | Image of the database engine. |
| pgManagerImage | Image of the manager          |
| pfsdImage      | Image of the pfsd             |
| pfsdToolImage  | Image of the pfsd tool        |

### Upgrade the Minor Version

Upgrade the image version used by your database cluster.

***Steps:***

1. Create a new configuration for the image, modify the configuration parameters to those for the image version you need to upgrade to, and use the command `kubectl` to make the configuration take effect.

Field Description:

| Field          | Field Description            |
| -------------- | ---------------------------- |
| pgEngineImage  | Image of the database engine |
| pgManagerImage | Image of the manager         |
| pfsdImage      | Image of the pfsd            |
| pfsdToolImage  | Image of the pfsd tool       |

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

2. Modify the created cluster.

```shell
kubectl edit mpdcluster your-cluster-name
```

3. Modify the field `versionCfgModifyTo.versionName` to the name of the specification you want to change the current one to. Add the following field whose description is shown in the table below.

| Field                          | Field Description                                            |
| ------------------------------ | ------------------------------------------------------------ |
| versionCfgModifyTo.versionName | The name of the specification you want to change the current one to, e.g., "image-open". |

***Example:***

```yaml
spec: 
  versionCfgModifyTo:
    versionName: your-new-config-name
```

### Recreate Completely

Recreate the database cluster.

***Steps:***

1. Modify the created cluster.

```shell
kubectl edit mpdcluster your-cluster-name
```

2. Add the following field whose description is shown in the table below.

| Field                             | Field Description                                            |
| --------------------------------- | ------------------------------------------------------------ |
| metadata.annotations.forceRebuild | If this field is set to true, it indicates that recreating the database cluster is required. |

***Example:***

```yaml
metadata:
  annotations:
    forceRebuild: "true"
```

***Check the status:***

Recreating:

```shell
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: Rebuild
```

Recreated:

```shell
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: Running
```

### Add a RO Node

Add a ro (read-only) node.

**Note**: Make sure you have spare machines. Otherwise, a port conflict will occur.

***Steps:***

1. Modify the created cluster.

```shell
kubectl edit mpdcluster your-cluster-name
```

2. Add the following field whose description is shown in the table below.

| Field       | Field Description                      |
| ----------- | -------------------------------------- |
| followerNum | Total number of ro nodes after adding. |

***Example:***

```yaml
spec:
  followerNum: 2
```

### Migrate the RW Node

Migrate the rw (read-write) node to another machine.

**Note**: The node can only be migrated to a node without any cluster instance. Otherwise, the port conflicts will occur.

***Steps:***

1. Modify the created cluster.

```shell
kubectl edit mpdcluster your-cluster-name
```

2. Add the following field whose description is shown in the table below.

| Field                        | Field Description                                            |
| ---------------------------- | ------------------------------------------------------------ |
| metadata.annotations.migrate | The format is `${insId}\|${targetNode}` (e.g., 7601\|dbm-02), in which the `insid` is the ID of the current rw node and the `targetNode` is the name of the destination machine. The targetNode can be obtained using the command `kubectl get node`. |

***Example:***

```yaml
metadata:
  annotations:
    migrate: 7601|dbm-02
```

***Verify the status***

Migrating:

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: MigrateRw
```

Migrated:

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: Running
```

### Migrate a RO Node

Migrate a ro (read-only) node to another machine.

**Note**: The node can only be migrated to a node without any cluster instance. Otherwise, the port conflicts will occur.

***Steps:***

1. Modify the created cluster.

```shell
kubectl edit mpdcluster your-cluster-name
```

2. Add the following field whose description is shown in the table below.

| Field                        | Field Description                                            |
| ---------------------------- | ------------------------------------------------------------ |
| metadata.annotations.migrate | The format is `${insId}\|${targetNode}` (e.g., 7601\|dbm-02), in which the `insid` is the ID of the ro node to be migrated and the `targetNode` is the name of the destination machine. The targetNode can be obtained using the command `kubectl get node`. |

***Example:***

```yaml
metadata:
  annotations:
    migrate: 7601|dbm-02
```

***Verify the status:***

Migrating:

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: MigrateRo
```

Migrated:

```yaml
[root@dbm-01 ~]# kubectl get mpdclusters.mpd.polardb.aliyun.com your-cluster-name -o yaml | grep clusterStatus
  clusterStatus: Running
```

### Increase Storage Capacity

Increase the storage capacity used by your database. 

Before that, you need to increase the disk capacity.

1. Modify the created cluster.

```shell
kubectl edit mpdcluster your-cluster-name
```

2. Add the following field whose description is shown in the table below.

| Field                              | Field Description                                            |
| ---------------------------------- | ------------------------------------------------------------ |
| metadata.annotations.extendStorage | If this field is set to true, it indicates that increasing the storage capacity is required. |

Modify the value of the following field to the disk size you need after increasing. The field is described as follows:

| Field                     | Field Description                                            |
| ------------------------- | ------------------------------------------------------------ |
| spec.shareStore.diskQuota | The disk size that can be used by your database engine, unit: MB. As PFS uses 10 GB as a unit, only a multiple of 10 GB is allowed for increasing the capacity each time. |

Example:

```yaml
metadata:
  annotations:
    extendStorage: "true"
```

### Retry after Failure

This operation is suitable for the situation in which the operation failed and the status is changed to Interrupt.

Change the value of interrupt.recover from F to T.

```yaml
metadata:
  annotations:
    interrupt.recover: F
```

### Delete the Instance Cluster

Perform this operation with caution.

```yaml
kubectl delete mpdcluster your-cluster-name
```

### Appendix A

The engine parameters are described as follows.

| Parameter Name                                   | Default Value | Restarting Required or Not | Value Range                                                  |
| ------------------------------------------------ | ------------- | -------------------------- | ------------------------------------------------------------ |
| archive_mode                                     | on            | Y                          | [always\|on\|off]                                            |
| autovacuum_vacuum_cost_delay                     | 0             | N                          | [-1-100]                                                     |
| autovacuum_vacuum_cost_limit                     | 10000         | N                          | [-1-10000]                                                   |
| auto_explain.log_analyze                         | off           | N                          | [on\|off]                                                    |
| auto_explain.log_buffers                         | off           | N                          | [on\|off]                                                    |
| auto_explain.log_format                          | text          | N                          | [text\|xml\|json\|yaml]                                      |
| auto_explain.log_min_duration                    | 5000          | N                          | [-1-2147483647]                                              |
| auto_explain.log_nested_statements               | off           | N                          | [on\|off]                                                    |
| auto_explain.log_timing                          | on            | N                          | [on\|off]                                                    |
| auto_explain.log_triggers                        | off           | N                          | [on\|off]                                                    |
| auto_explain.log_verbose                         | off           | N                          | [on\|off]                                                    |
| auto_explain.sample_rate                         | 1             | N                          | [0-1]                                                        |
| checkpoint_completion_target                     | 0.9           | N                          | [0-1]                                                        |
| checkpoint_timeout                               | 300           | N                          | [1-86400]                                                    |
| default_transaction_deferrable                   | off           | N                          | [on\|off]                                                    |
| default_with_oids                                | off           | N                          | [on\|off]                                                    |
| default_with_rowids                              | off           | N                          | [on\|off]                                                    |
| enable_partitionwise_aggregate                   | on            | N                          | [on\|off]                                                    |
| enable_partitionwise_join                        | on            | N                          | [on\|off]                                                    |
| extra_float_digits                               | 0             | N                          | [-15-3]                                                      |
| idle_in_transaction_session_timeout              | 3600000       | N                          | ^(0\|[1-9]\d{3,8}\|1\d{9}\|2000000000)$                      |
| jit                                              | off           | N                          | [on\|off]                                                    |
| lock_timeout                                     | 0             | N                          | ^(0\|[1-9]\d{3,8}\|1\d{9}\|2000000000)$                      |
| log_min_duration_statement                       | -1            | N                          | [-1-2147483647]                                              |
| log_statement                                    | all           | N                          | [none\|ddl\|mod\|all]                                        |
| log_temp_files                                   | 100000        | N                          | [-1-2147483647]                                              |
| log_timezone                                     | 'PRC'         | N                          | ^'(((UTC)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|((GMT)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|CST6CDT\|Poland\|Kwajalein\|MST\|NZ\|Universal\|Libya\|Turkey\|EST5EDT\|Greenwich\|NZ-CHAT\|MET\|Portugal\|GMT-0\|CET\|Eire\|PST8PDT\|Jamaica\|GMT\|Zulu\|Japan\|ROC\|GB-Eire\|ROK\|Navajo\|Singapore\|posixrules\|GB\|EST\|GMT0\|Hongkong\|PRC\|Iran\|MST7MDT\|WET\|W-SU\|UCT\|Cuba\|Egypt\|EET\|Israel\|UTC\|HST\|Iceland)'$ |
| max_parallel_maintenance_workers                 | 2             | N                          | [0-1024]                                                     |
| max_parallel_workers                             | 2             | N                          | [0-512]                                                      |
| max_parallel_workers_per_gather                  | 2             | N                          | [0-512]                                                      |
| min_parallel_index_scan_size                     | 64            | N                          | [0-715827882]                                                |
| min_parallel_table_scan_size                     | 1024          | N                          | [0-715827882]                                                |
| old_snapshot_threshold                           | -1            | Y                          | [-1-86400]                                                   |
| polar_comp_dynatune                              | 0             | Y                          | [0-100]                                                      |
| polar_comp_dynatune_profile                      | oltp          | Y                          | [oltp\|reporting\|mixed]                                     |
| polar_comp_enable_pruning                        | on            | N                          | [on\|off]                                                    |
| polar_comp_redwood_date                          | on            | N                          | [on\|off]                                                    |
| polar_comp_redwood_greatest_least                | on            | N                          | [on\|off]                                                    |
| polar_comp_redwood_raw_names                     | on            | N                          | [on\|off]                                                    |
| polar_comp_redwood_strings                       | on            | N                          | [on\|off]                                                    |
| polar_comp_stmt_level_tx                         | on            | N                          | [on\|off]                                                    |
| dbms_job.database_name                           | 'postgres'    | Y                          | ^'\w+'$                                                      |
| statement_timeout                                | 0             | N                          | ^(0\|[1-9]\d{3,8}\|1\d{9}\|2000000000)$                      |
| temp_file_limit                                  | 524288000     | N                          | [-1-1048576000]                                              |
| timezone                                         | 'PRC'         | N                          | ^'(((UTC)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|((GMT)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|CST6CDT\|Poland\|Kwajalein\|MST\|NZ\|Universal\|Libya\|Turkey\|EST5EDT\|Greenwich\|NZ-CHAT\|MET\|Portugal\|GMT-0\|CET\|Eire\|PST8PDT\|Jamaica\|GMT\|Zulu\|Japan\|ROC\|GB-Eire\|ROK\|Navajo\|Singapore\|posixrules\|GB\|EST\|GMT0\|Hongkong\|PRC\|Iran\|MST7MDT\|WET\|W-SU\|UCT\|Cuba\|Egypt\|EET\|Israel\|UTC\|HST\|Iceland)'$ |
| track_commit_timestamp                           | off           | Y                          | [on\|off]                                                    |
| vacuum_defer_cleanup_age                         | 0             | N                          | [0-1000000]                                                  |
| wal_level                                        | logical       | Y                          | [replica\|logical]                                           |
| work_mem                                         | 4096          | N                          | [4096-524288]                                                |
| polar_empty_string_is_null_enable                | on            | N                          | [on\|off]                                                    |
| polar_enable_varchar2_length_with_byte           | on            | N                          | [on\|off]                                                    |
| polar_enable_base64_decode                       | on            | N                          | [on\|off]                                                    |
| polar_enable_nls_date_format                     | on            | N                          | [on\|off]                                                    |
| polar_enable_rowid                               | on            | N                          | [on\|off]                                                    |
| postgres_fdw.polar_connection_check              | off           | N                          | [on\|off]                                                    |
| polar_comp_custom_plan_tries                     | 5             | N                          | [-1-100]                                                     |
| dblink.polar_auto_port_mapping                   | off           | N                          | [on\|off]                                                    |
| dblink.polar_connection_check                    | off           | N                          | [on\|off]                                                    |
| pg_stat_statements.enable_superuser_track        | on            | N                          | [on\|off]                                                    |
| cron.polar_allow_superuser_task                  | on            | Y                          | [on\|off]                                                    |
| polar_stat_sql.enable_qps_monitor                | on            | N                          | [on\|off]                                                    |
| polar_enable_audit_log_bind_sql_parameter        | off           | N                          | [on\|off]                                                    |
| polar_enable_audit_log_bind_sql_parameter_new    | off           | N                          | [on\|off]                                                    |
| polar_enable_replica_use_smgr_cache              | on            | Y                          | [off\|on]                                                    |
| idle_session_timeout                             | 0             | N                          | [0-2147483647]                                               |
| polar_resource_group.total_mem_limit_rate        | 95            | N                          | [50-100]                                                     |
| polar_resource_group.total_mem_limit_remain_size | 524288        | N                          | [131072-2097151]                                             |

### Appendix B

The engine parameters are described as follows.

| Parameter Name                                   | Default Value | Restarting Required or Not | Value Range                                                  |
| ------------------------------------------------ | ------------- | -------------------------- | ------------------------------------------------------------ |
| archive_mode                                     | on            | Y                          | [always\|on\|off]                                            |
| autovacuum_vacuum_cost_delay                     | 0             | N                          | [-1-100]                                                     |
| autovacuum_vacuum_cost_limit                     | 10000         | N                          | [-1-10000]                                                   |
| auto_explain.log_analyze                         | off           | N                          | [on\|off]                                                    |
| auto_explain.log_buffers                         | off           | N                          | [on\|off]                                                    |
| auto_explain.log_format                          | text          | N                          | [text\|xml\|json\|yaml]                                      |
| auto_explain.log_min_duration                    | 5000          | N                          | [-1-2147483647]                                              |
| auto_explain.log_nested_statements               | off           | N                          | [on\|off]                                                    |
| auto_explain.log_timing                          | on            | N                          | [on\|off]                                                    |
| auto_explain.log_triggers                        | off           | N                          | [on\|off]                                                    |
| auto_explain.log_verbose                         | off           | N                          | [on\|off]                                                    |
| auto_explain.sample_rate                         | 1             | N                          | [0-1]                                                        |
| checkpoint_completion_target                     | 0.9           | N                          | [0-1]                                                        |
| checkpoint_timeout                               | 300           | N                          | [1-86400]                                                    |
| default_transaction_deferrable                   | off           | N                          | [on\|off]                                                    |
| default_with_oids                                | off           | N                          | [on\|off]                                                    |
| default_with_rowids                              | off           | N                          | [on\|off]                                                    |
| enable_partitionwise_aggregate                   | on            | N                          | [on\|off]                                                    |
| enable_partitionwise_join                        | on            | N                          | [on\|off]                                                    |
| extra_float_digits                               | 0             | N                          | [-15-3]                                                      |
| idle_in_transaction_session_timeout              | 3600000       | N                          | ^(0\|[1-9]\d{3,8}\|1\d{9}\|2000000000)$                      |
| jit                                              | off           | N                          | [on\|off]                                                    |
| lock_timeout                                     | 0             | N                          | ^(0\|[1-9]\d{3,8}\|1\d{9}\|2000000000)$                      |
| log_min_duration_statement                       | -1            | N                          | [-1-2147483647]                                              |
| log_statement                                    | all           | N                          | [none\|ddl\|mod\|all]                                        |
| log_temp_files                                   | 100000        | N                          | [-1-2147483647]                                              |
| log_timezone                                     | 'PRC'         | N                          | ^'(((UTC)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|((GMT)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|CST6CDT\|Poland\|Kwajalein\|MST\|NZ\|Universal\|Libya\|Turkey\|EST5EDT\|Greenwich\|NZ-CHAT\|MET\|Portugal\|GMT-0\|CET\|Eire\|PST8PDT\|Jamaica\|GMT\|Zulu\|Japan\|ROC\|GB-Eire\|ROK\|Navajo\|Singapore\|posixrules\|GB\|EST\|GMT0\|Hongkong\|PRC\|Iran\|MST7MDT\|WET\|W-SU\|UCT\|Cuba\|Egypt\|EET\|Israel\|UTC\|HST\|Iceland)'$ |
| max_parallel_maintenance_workers                 | 2             | N                          | [0-1024]                                                     |
| max_parallel_workers                             | 2             | N                          | [0-512]                                                      |
| max_parallel_workers_per_gather                  | 2             | N                          | [0-512]                                                      |
| min_parallel_index_scan_size                     | 64            | N                          | [0-715827882]                                                |
| min_parallel_table_scan_size                     | 1024          | N                          | [0-715827882]                                                |
| old_snapshot_threshold                           | -1            | Y                          | [-1-86400]                                                   |
| polar_comp_dynatune                              | 0             | Y                          | [0-100]                                                      |
| polar_comp_dynatune_profile                      | oltp          | Y                          | [oltp\|reporting\|mixed]                                     |
| polar_comp_enable_pruning                        | on            | N                          | [on\|off]                                                    |
| polar_comp_redwood_date                          | on            | N                          | [on\|off]                                                    |
| polar_comp_redwood_greatest_least                | on            | N                          | [on\|off]                                                    |
| polar_comp_redwood_raw_names                     | on            | N                          | [on\|off]                                                    |
| polar_comp_redwood_strings                       | on            | N                          | [on\|off]                                                    |
| polar_comp_stmt_level_tx                         | on            | N                          | [on\|off]                                                    |
| dbms_job.database_name                           | 'postgres'    | Y                          | ^'\w+'$                                                      |
| statement_timeout                                | 0             | N                          | ^(0\|[1-9]\d{3,8}\|1\d{9}\|2000000000)$                      |
| temp_file_limit                                  | 524288000     | N                          | [-1-1048576000]                                              |
| timezone                                         | 'PRC'         | N                          | ^'(((UTC)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|((GMT)(-){0,1}(\d\|[1-9]\d\|1([0-5]\d\|6[0-7])))\|CST6CDT\|Poland\|Kwajalein\|MST\|NZ\|Universal\|Libya\|Turkey\|EST5EDT\|Greenwich\|NZ-CHAT\|MET\|Portugal\|GMT-0\|CET\|Eire\|PST8PDT\|Jamaica\|GMT\|Zulu\|Japan\|ROC\|GB-Eire\|ROK\|Navajo\|Singapore\|posixrules\|GB\|EST\|GMT0\|Hongkong\|PRC\|Iran\|MST7MDT\|WET\|W-SU\|UCT\|Cuba\|Egypt\|EET\|Israel\|UTC\|HST\|Iceland)'$ |
| track_commit_timestamp                           | off           | Y                          | [on\|off]                                                    |
| vacuum_defer_cleanup_age                         | 0             | N                          | [0-1000000]                                                  |
| wal_level                                        | logical       | Y                          | [replica\|logical]                                           |
| work_mem                                         | 4096          | N                          | [4096-524288]                                                |
| polar_empty_string_is_null_enable                | on            | N                          | [on\|off]                                                    |
| polar_enable_varchar2_length_with_byte           | on            | N                          | [on\|off]                                                    |
| polar_enable_base64_decode                       | on            | N                          | [on\|off]                                                    |
| polar_enable_nls_date_format                     | on            | N                          | [on\|off]                                                    |
| polar_enable_rowid                               | on            | N                          | [on\|off]                                                    |
| postgres_fdw.polar_connection_check              | off           | N                          | [on\|off]                                                    |
| polar_comp_custom_plan_tries                     | 5             | N                          | [-1-100]                                                     |
| dblink.polar_auto_port_mapping                   | off           | N                          | [on\|off]                                                    |
| dblink.polar_connection_check                    | off           | N                          | [on\|off]                                                    |
| pg_stat_statements.enable_superuser_track        | on            | N                          | [on\|off]                                                    |
| cron.polar_allow_superuser_task                  | on            | Y                          | [on\|off]                                                    |
| polar_stat_sql.enable_qps_monitor                | on            | N                          | [on\|off]                                                    |
| polar_enable_audit_log_bind_sql_parameter        | off           | N                          | [on\|off]                                                    |
| polar_enable_audit_log_bind_sql_parameter_new    | off           | N                          | [on\|off]                                                    |
| polar_enable_replica_use_smgr_cache              | on            | Y                          | [off\|on]                                                    |
| idle_session_timeout                             | 0             | N                          | [0-2147483647]                                               |
| polar_resource_group.total_mem_limit_rate        | 95            | N                          | [50-100]                                                     |
| polar_resource_group.total_mem_limit_remain_size | 524288        | N                          | [131072-2097151]                                             |



The following table lists high-risk parameters. Confirm the impact first and modify them with caution.

| Parameter Name                               | Default Value                                                | Restarting Required or Not | Value Range                         |
| -------------------------------------------- | ------------------------------------------------------------ | -------------------------- | ----------------------------------- |
| autovacuum_max_workers                       | 5                                                            | Y                          | [1-262143]                          |
| autovacuum_work_mem                          | 200MB                                                        | N                          | [-1-2147483647]                     |
| enable_hashagg                               | off                                                          | N                          | [on\|off]                           |
| log_connections                              | off                                                          | N                          | [on\|off]                           |
| log_disconnections                           | off                                                          | N                          | [on\|off]                           |
| max_standby_streaming_delay                  | 30000                                                        | N                          | [-1-2147483647]                     |
| polar_bgwriter_batch_size_flushlist          | 100                                                          | Y                          | [1-10000]                           |
| polar_bgwriter_max_batch_size                | 5000                                                         | N                          | [0-1073741823]                      |
| polar_bgwriter_sleep_lsn_lag                 | 100                                                          | N                          | [0-1000]                            |
| polar_buffer_copy_lsn_lag_with_cons_lsn      | 100                                                          | N                          | [1-1000]                            |
| polar_buffer_copy_min_modified_count         | 5                                                            | N                          | [0-100]                             |
| polar_bulk_extend_size                       | 512                                                          | N                          | [0-1073741823]                      |
| polar_check_checkpoint_legal_interval        | 10                                                           | N                          | [1-3600]                            |
| polar_clog_slot_size                         | 512                                                          | Y                          | [128-8192]                          |
| polar_comp_early_lock_release                | off                                                          | N                          | [on\|off]                           |
| polar_copy_buffers                           | 16384                                                        | Y                          | [128-1073741823]                    |
| polar_enable_connectby_multiexpr             | on                                                           | N                          | [on\|off]                           |
| polar_enable_physical_repl_non_super_wal_snd | off                                                          | N                          | [on\|off]                           |
| polar_enable_show_polar_comp_version         | on                                                           | N                          | [on\|off]                           |
| polar_enable_syslog_pipe_buffer              | off                                                          | N                          | [on\|off]                           |
| polar_global_temp_table_debug                | off                                                          | N                          | [on\|off]                           |
| polar_max_log_files                          | 20                                                           | N                          | [-1-2147483647]                     |
| polar_max_super_conns                        | 100                                                          | N                          | [-1-262143]                         |
| polar_num_active_global_temp_table           | 1000                                                         | Y                          | [0-1000000]                         |
| polar_parallel_bgwriter_check_interval       | 10                                                           | N                          | [1-600]                             |
| polar_parallel_bgwriter_delay                | 10                                                           | N                          | [1-10000]                           |
| polar_parallel_bgwriter_workers              | 5                                                            | N                          | [0-8]                               |
| polar_parallel_new_bgwriter_threshold_time   | 10                                                           | N                          | [1-3600]                            |
| polar_read_ahead_xlog_num                    | 200                                                          | Y                          | [0-200]                             |
| polar_redo_hashtable_size                    | 131072                                                       | Y                          | [16-1073741823]                     |
| polar_ring_buffer_vacuum                     | 128                                                          | N                          | [10-1000]                           |
| polar_spl_savepoint_enable                   | on                                                           | N                          | [on\|off]                           |
| polar_temp_relation_file_in_shared_storage   | off                                                          | Y                          | [on\|off]                           |
| polar_vfs.libmm_num_partition                | 32                                                           | Y                          | [4-128]                             |
| polar_vfs.libmm_size                         | 67108864                                                     | Y                          | [67108864-1073741824]               |
| polar_enable_fullpage_snapshot               | off                                                          | N                          | [on\|off]                           |
| polar_enable_default_polar_comp_guc          | on                                                           | N                          | [on\|off]                           |
| polar_stat_stale_cost                        | 0.0001                                                       | N                          | [0-2147483647]                      |
| polar_auditlog_max_query_length              | 4096                                                         | N                          | [512-49152]                         |
| polar_use_statistical_relpages               | on                                                           | N                          | [on\|off]                           |
| shared_preload_libraries                     | 'polar_vfs,polar_worker,dbms_pipe,polar_gen,pg_stat_statements,auth_delay,auto_explain,pg_cron,dbms_job,polar_stat_sql' | Y                          | ^'(polar_vfs,polar_worker)(,\w+)*'$ |
| wal_keep_segments                            | 32                                                           | N                          | [0-100000]                          |
| temp_tablespaces                             | 'polar_tmp'                                                  | N                          |                                     |
| polar_csn_enable                             | on                                                           | Y                          | [on\|off]                           |
| polar_max_non_super_conns                    | 400                                                          | N                          | [-1-262143]                         |
| polar_enable_master_xlog_read_ahead          | on                                                           | Y                          | [on\|off]                           |
| polar_enable_early_launch_checkpointer       | on                                                           | Y                          | [on\|off]                           |
| polar_enable_early_launch_parallel_bgwriter  | on                                                           | Y                          | [on\|off]                           |
| polar_enable_wal_prefetch                    | off                                                          | Y                          | [on\|off]                           |
| polar_enable_persisted_buffer_pool           | on                                                           | Y                          | [on\|off]                           |
| polar_cast_decode_list                       | ''                                                           | N                          |                                     |