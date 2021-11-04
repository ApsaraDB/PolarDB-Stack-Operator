# 一键安装

安装前请先查阅[部署前置要求](prerequest.md)确认您的环境满足最低要求

### 脚本一键安装环境要求
1. 一键安装仅支持 centos 系统
2. 已经安装好 docker 和 kubernetes
3. 已经安装好元数据库，并且可以通过 IP + port 的形式连通，数据库类型目前仅支持 mysql 
4. 每台机器互相之间已经配置好免密 ssh 

### 修改配置
默认配置为 env.yaml，您需要修改为您自己环境的信息，格式示例如下：
```
dbm_hosts:
  - ip: 10.0.0.77
    name: dbm-01
  - ip: 10.0.0.78
    name: dbm-02
  - ip: 10.0.0.79
    name: dbm-03
network:
  interface: eth0
k8s:
  host: 10.0.0.77
  port: 6443
```
字段说明：

| 字段 | 含义 | 要求 |
| --- | --- | --- |
| dbm_hosts.ip | 可以访问到您三台主机的 IP 地址 | - |
| dbm_hosts.name | 您三台机器的主机名 | - |
| network.interface | 您的网口名称 | 通过 ifconfig 查询得到 |
| k8s.host | 您的 k8s apiserver IP 地址 | - |
| k8s.port | 您的 k8s apiserver port | - |
| metabase.host | 您的元数据库 IP 地址 | - |
| metabase.port | 您的元数据库 port， | - |
| metabase.user | 您的元数据库用户名 | - |
| metabase.password | 您的元数据库登录密码 | - | 
| metabase.type | 您的元数据库类型 | 目前仅支持 mysql |
| metabase.version| 您的元数据库版本号 | - |

### 安装
克隆工程
```shell
git clone https://github.com/ApsaraDB/PolarDB-Stack-Operator.git
```
按上述说明修改配置 
```shell
vim env.yaml
```
运行安装脚本
```shell
./install.sh
```
如无错误则表示安装成功

### 创建 DB 集群

1. 创建 PVC ，调用接口，示例如下，IP 需要换成您的主机 IP 。

```shell
curl -X POST "http://10.0.0.77:2002/pvcs" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"name\": \"pvc-32ze341nncwlczm47bsre\", \"namespace\": \"default\", \"need_format\": true, \"volume_id\": \"32ze341nncwlczm47bsre\", \"volume_type\": \"lun\"}"
```

2. 查看 PVC 是否创建成功。

```plain
kubectl get pvc 
```

3. 创建实例集群，将前面创建的 PVC name 填入您的集群配置。

```shell
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
