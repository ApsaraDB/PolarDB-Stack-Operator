# Install Manually

Before installation, refer to [Deployment Prerequisites](prerequest-EN.md) to make sure that your environment meets the minimum requirements.

## Prepare the Machine

Allocate huge pages for the machine. It is suggested that the allocated huge page size should account for approximately 75% of the machine's total memory size.

```bash
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

After modifying the huge pages, please restart the server manually.

### Step 1. Install Docker

Install Docker. For details, refer to [Install Docker Engine](https://docs.docker.com/engine/install/).

### Step2. Install Kubernetes

Install Kubernetes version 1.14 and above. For details, refer to [Kubernetes Documentation](https://kubernetes.io/docs/setup/).

### Step 3. Install the MPD Controller

1. Use the script ./build.sh to build an image or directly use the image polardb/polar-mpd-controller:0.0.1-SNAPSHOT.

2. Install the mpdcluster crd.

   a. Download the sample YAML file.

```shell
wget https://github.com/ApsaraDB/PolarDB-Stack-Operator/blob/master/config/all.yaml
```

​      b. Set KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT to the IP address and port number of the apiserver of your Kubernetes cluster, respectively.

```shell
- name: KUBERNETES_SERVICE_HOST
  value: 10.0.0.77
- name: KUBERNETES_SERVICE_PORT
  value: "6443"
```

​      c. Modify the image version and apply the modified configuration.

```shell
kubectl apply -f all.yaml
```

3. Set the node label.

```shell
wget https://github.com/ApsaraDB/PolarDB-Stack-Operator/blob/master/script/set_labels.sh
./set_labels.sh
```

### Step 4. Install the Storage Management Module

1. Install the sms-agent.

​		a. Install the multipath.

​		b. Install and start supervisord, and make sure that it is running normally.

​		c. In the sms project, compile agent and generate the binary package bin/sms-agent.

```shell
make build-agent
```
​		d. Copy the binary package to the directory /home/a/project/t-polardb-sms-agent/bin/polardb-sms-agent on the host.

​		e. Configure the /etc/supervisord.d/polardb-sms-agent.ini.

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

​		f. Configure the /etc/polardb-sms-agent.conf.

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

​		e. Reload supervisor.

```shell
supervisorctl reload
```

2. Compile polardb-sms-manager to build an image.

```python
./build/build-manager.sh
```

Or you can directly use the packaged image.
```shell
docker pull polardb/polardb-sms-manager:1.0.0
```

3. Create the storage management metadatabase

The metadatabase should be a relational database. Currently, it can only be MySQL. You need to create the metadatabase by yourself and make sure that the metadatabase can be connected.

Create the metadata table structure. You can refer to the example PolarDB-Stack-Storage [scripts/db.sql](https://github.com/ApsaraDB/PolarDB-Stack-Storage/blob/master/scripts/db.sql).

4. Create sms-manager deployment
   

​		a. Download the sample YAML file.

```shell
wget https://github.com/ApsaraDB/PolarDB-Stack-Storage/blob/master/deploy/all.yaml
```

​		b. Set KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT to the IP address and port number of the apiserver of your Kubernetes cluster, respectively.

```shell
- name: KUBERNETES_SERVICE_HOST
  value: 10.0.0.77
- name: KUBERNETES_SERVICE_PORT
  value: "6443"
```

​		c. Add the configuration information of the metadatabase to the configmap.

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

​		d. Apply the modified configuration.

```shell
kubectl apply -f all.yaml
```

5. Wait for sms to start up. If it can start up normally, you will see the cm configurations starting with cloud-provider-wwid-usage-.

```shell
[root@dbm-01 ~]# kubectl -n kube-system get cm 
NAME                                 DATA   AGE
cloud-provider-wwid-usage-dbm-01     4      31s
cloud-provider-wwid-usage-dbm-02     4      33s
cloud-provider-wwid-usage-dbm-03     4      33s
```

> **Note**: If there is data in the column of DATA, it indicates the shared disk required by PolarDB has been scanned.

6. Install pfs rpm packages on all machines. For details, refer to [Compile and Install PFS](https://github.com/ApsaraDB/PolarDB-FileSystem).

```shell
wget https://github.com/ApsaraDB/PolarDB-FileSystem/releases/download/pfsd4pg-release-1.2.41-20211018/t-pfsd-opensource-1.2.41-1.el7.x86_64.rpm
rpm -ivh t-pfsd-opensource-1.2.41-1.el7.x86_64.rpm
```

7. Configure pg_hba.conf.

​		Execute the following commands on each machine.

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

### Install the Network Management Module

1. Compile the source code into a Docker image.

```shell
wget https://github.com/ApsaraDB/PolarDB-Stack-Daemon/blob/master/build.sh
./build.sh
```

2. Download the sample YAML file.

```shell
wget https://github.com/ApsaraDB/PolarDB-Stack-Daemon/blob/master/deploy/all.yaml
```

3. View the NIC information using the command `ifconfig` and modify the sample YAML file.

```shell
NET_CARD_NAME: bond0

MiniLvs_BackendIf: bond0
```

4. Create the network DaemonSet

```shell
kubectl apply -f all.yaml
```

### Create a DB Cluster

1. Call the corresponding API to create a PVC as the example below. Note that the IP address should be replaced with that of your host.

```shell
curl -X POST "http://10.0.0.77:2002/pvcs" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"name\": \"pvc-32ze341nncwlczm47bsre\", \"namespace\": \"default\", \"need_format\": true, \"volume_id\": \"32ze341nncwlczm47bsre\", \"volume_type\": \"lun\"}"
```

2. Check whether the PVC is created successfully.

```plain
kubectl get pvc 
```

3. Create a cluster instance and fill in the cluster configuration with the PVC name created earlier.

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