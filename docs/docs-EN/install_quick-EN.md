# Install Automatically

Before installation, refer to [Deployment Prerequisites](prerequest-EN.md) to make sure that your environment meets the minimum requirements.

### Environment Requirements
1. Only the automatic installation can only be performed on the CentOS operating system.
2. Make sure you have installed Docker and Kubernetes.
3. Make sure you have installed the metadatabase and can connect to it using the IP address and port number. Currently, the metadatabase can only be MySQL.
4. Make sure you have configured password-less SSH access for all machines to access each other.
5. Make sure you have installed Helm.

### Modify Configurations
The default configuration is in the file env.yaml. You need to modify it according to your own environment information. The format example is as follows:
```yaml
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
metabase:
  host: 10.0.0.77
  port: 3306
  user: polar
  password: password
  type: mysql
  version: 8.0.26
```
Field Description:

| Field | Description | Requirement |
| --- | --- | --- |
| dbm_hosts.ip | IP address that can access three hosts | - |
| dbm_hosts.name | Host names of three machines | - |
| network.interface | NIC name | You can use the command ifconfig to get it. |
| k8s.host | IP address of the Kubernetes apiserver | - |
| k8s.port | Port number of the Kubernetes apiserver | - |
| metabase.host | IP address of the metadatabase | - |
| metabase.port | Port number of the metadatabase | - |
| metabase.user | User name of the metadatabase | - |
| metabase.password | Password for logging in to the metadatabase | - |
| metabase.type | Metadatabase type | Currently, it can only be MySQL. |
| metabase.version| Version number of the metadatabase | - |

### Install
Clone the project.
```shell
git clone https://github.com/ApsaraDB/PolarDB-Stack-Operator.git
```
Modify the configuration following the instructions above:
```shell
vim env.yaml
```
Run the installation script.
```bash
./install.sh
```
If no error is returned, it indicates that installing succeeded.

### Create a DB Cluster

1. Call the corresponding API to create a PVC as the example below. Note that the IP address should be replaced with that of your host.

```bash
curl -X POST "http://10.0.0.77:2002/pvcs" -H "accept: application/json" -H "Content-Type: application/json" -d "{ \"name\": \"pvc-32ze341nncwlczm47bsre\", \"namespace\": \"default\", \"need_format\": true, \"volume_id\": \"32ze341nncwlczm47bsre\", \"volume_type\": \"lun\"}"
```

2. Check whether the PVC is created successfully.

```bash
kubectl get pvc 
```

3. Create a cluster instance and fill in the cluster configuration with the PVC name created earlier.

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
