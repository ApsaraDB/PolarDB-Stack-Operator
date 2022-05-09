# 手工安装手册

安装前请先查阅[部署前置要求](prerequest.md)确认您的环境满足最低要求

### 预装步骤
给机器分配大页内存

分配的大页内存建议大约占机器总内存的 75%

```shell
#!/bin/bash
hugepage_mem_gb=256  # 256G
hugepages=`echo "${hugepage_mem_gb} * 1024 / 2" | bc`
hugepages=${hugepages%%.*}
echo $hugepages
grep 'vm.nr_hugepages=' /etc/sysctl.conf
if [ $? != 0 ]; then
    echo -e "\nvm.nr_hugepages=${hugepages}" >> /etc/sysctl.conf
else
    sed -i "/^vm.nr_hugepages/vm.nr_hugepages=${hugepages}" /etc/sysctl.conf
fi
sysctl -p /etc/sysctl.conf
grep 'none /dev/hugepages1G hugetlbfs pagesize=1G 0 0' /etc/fstab
if [ $? != 0 ]; then
    sed -i '$a\none /dev/hugepages1G hugetlbfs pagesize=1G 0 0' /etc/fstab
fi
grep 'none /dev/hugepages2M hugetlbfs pagesize=2M 0 0' /etc/fstab
if [ $? != 0 ]; then
    sed -i '$a\none /dev/hugepages2M hugetlbfs pagesize=2M 0 0' /etc/fstab
fi
sed -i 's/default_hugepagesz=1G hugepagesz=1G hugepages=20/default_hugepagesz=2M hugepagesz=1G hugepagesz=2M hugepages=5120/g' /etc/default/grub
grub2-mkconfig -o "$(sudo readlink -e /etc/grub2.cfg)"
```

修改完大页内存后，请手动重启服务器。

### 步骤一、安装 docker

安装docker，请参见[Docker安装指南](https://docs.docker.com/engine/install/)

### 步骤二、安装 kubernetes

请安装kubernetes，版本要求1.14及以上版本。请参见[[K8S安装指南]](https://kubernetes.io/docs/setup/)。

### 步骤三、安装 mpd controller

1. ./build.sh 生成镜像，或直接使用 polardb/polar-mpd-controller:0.0.1-SNAPSHOT

2. 安装 mpdcluster crd

   a. 下载示例 yaml

```shell
wget https://github.com/ApsaraDB/PolarDB-Stack-Operator/blob/master/config/all.yaml
```

​      b. 修改 KUBERNETES_SERVICE_HOST 及 KUBERNETES_SERVICE_PORT 为您 k8s 集群 apiserver 的 IP 及端口

```shell
- name: KUBERNETES_SERVICE_HOST
  value: 10.0.0.77
- name: KUBERNETES_SERVICE_PORT
  value: "6443"
```

​      c. 修改镜像版本病应用apply 修改好的配置

```shell
kubectl apply -f all.yaml
```

3. 设置 node label

```shell
wget https://github.com/ApsaraDB/PolarDB-Stack-Operator/blob/master/script/set_labels.sh
./set_labels.sh
```

### 步骤四、安装存储管理

1. 安装 sms-agent

a. 安装 multipath
   
b. 安装并启动 supervisord ，并确认正常运行。

c. 在 sms 工程编译 agent，生成二进制包 bin/sms-agent。

```shell
make build-agent
```
d. 拷贝二进制包到主机的 /home/a/project/t-polardb-sms-agent/bin/polardb-sms-agent 目录上。

e. 配置 /etc/supervisord.d/polardb-sms-agent.ini

```python
AGENT_INI="/etc/supervisord.d/polardb-sms-agent.ini"

NODE_IP=$(ifconfig bond0 | grep netmask | awk '{print $2}')

cat <<EOF >$AGENT_INI
[program:polardb-sms-agent]
command=/home/a/project/t-polardb-sms-agent/bin/polardb-sms-agent --port=18888 --node-ip=$NODE_IP --node-id=%(host_node_name)s
process_name=%(program_name)s
startretries=1000
autorestart=unexpected
autostart=true
EOF
```

f. 配置 /etc/polardb-sms-agent.conf

```python
AGENT_CONF="/etc/polardb-sms-agent.conf"

cat <<EOF >$AGENT_CONF
blacklist {
    attachlist {
    }
    locallist {
    }
}
EOF
```

e. reload supervisor
```shell
supervisorctl reload
```

2. 编译 polardb-sms-manager 生成镜像

```python
./build/build-manager.sh
```

或使用已经打包好的镜像 
```shell
docker pull polardb/polardb-sms-manager:1.0.0
```

3. 创建存储管理元数据库

元数据库需要是关系型数据库，目前暂只支持 mysql，您需要自行创建元数据库并且保证该数据库可以连通。

创建元数据表结构，示例参考 PolarDB-Stack-Storage [scripts/db.sql](https://github.com/ApsaraDB/PolarDB-Stack-Storage/blob/master/scripts/db.sql)

4. 创建 sms-manager deployment
   
a. 下载示例 yaml

```shell
wget https://github.com/ApsaraDB/PolarDB-Stack-Storage/blob/master/deploy/all.yaml
```

b. 修改 KUBERNETES_SERVICE_HOST 及 KUBERNETES_SERVICE_PORT 为您 k8s 集群 apiserver 的 IP 及端口

```shell
- name: KUBERNETES_SERVICE_HOST
  value: 10.0.0.77
- name: KUBERNETES_SERVICE_PORT
  value: "6443"
```

c. 将元数据库的信息配置进 config map

```shell
apiVersion: v1
data:
  metabase.yml: |-
    metabase:
      host: 10.0.0.77
      port: 3306
      user: polar
      password: password
      type: mysql
      version: 8.0.26
kind: ConfigMap
metadata:
  name: metabase-config
  namespace: kube-system
```

d. apply 修改好的配置

```shell
kubectl apply -f all.yaml
```

5. 等待 sms 启动，如启动正常将能够看到 cloud-provider-wwid-usage- 开头的几个 cm 配置。

```shell
[root@dbm-01 ~]# kubectl -n kube-system get cm 
NAME                                 DATA   AGE
cloud-provider-wwid-usage-dbm-01     4      31s
cloud-provider-wwid-usage-dbm-02     4      33s
cloud-provider-wwid-usage-dbm-03     4      33s
```

注意：DATA 中有数据表示已经扫描到 polardb 所需的共享盘。

6. 在机器上安装 pfs，需要分别在您所有的机器上安装 pfs rpm 包。
[pfs 编译安装](https://github.com/ApsaraDB/PolarDB-FileSystem) 

```shell
wget https://github.com/ApsaraDB/PolarDB-FileSystem/releases/download/pfsd4pg-release-1.2.41-20211018/t-pfsd-opensource-1.2.41-1.el7.x86_64.rpm
rpm -ivh t-pfsd-opensource-1.2.41-1.el7.x86_64.rpm
```

7. 配置 pg_hba.conf

在每台机器上分别执行
```shell
mkdir -p /etc/postgres
cat <<EOF >"/etc/postgres/pg_hba.conf"
local all all           trust
host all postgres all reject
host all all all md5
local replication postgres           trust
host replication postgres all reject
host replication all all md5
EOF
```

### 安装网络管理模块

1. 将源码编译成 docker 镜像

```shell
wget https://github.com/ApsaraDB/PolarDB-Stack-Daemon/blob/master/build.sh
./build.sh
```

2. 下载示例 yaml

```shell
wget https://github.com/ApsaraDB/PolarDB-Stack-Daemon/blob/master/deploy/all.yaml
```

3. 使用 ifconfig 查看您的网口信息，并修改示例 yaml

```shell
NET_CARD_NAME: bond0

MiniLvs_BackendIf: bond0
```

4. 创建网络 DaemonSet

```shell
kubectl apply -f all.yaml
```

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