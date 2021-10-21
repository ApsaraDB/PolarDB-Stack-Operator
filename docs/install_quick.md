# 一键安装
### 前置要求
1. 已经安装好 docker 和 kubernetes
2. 您的三台机器互相之间已经配置好免密登录，可以通过 ssh root@hostname 直接登录
### 修改配置
默认配置为 env.yaml，您需要修改为您自己环境的信息，格式示例如下：
```
dbm_hosts:
  - ip: 10.0.0.77
    name: dbm-01
  - ip: 10.0.0.78
    name: dbm-02
  - ip: 10.0.0.79
    name: r03.dbm-03
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