#!/bin/bash

while true; do
    kubectl get node
    if [[ $? != 0 ]]; then
        echo "[$(basename $0)]" "waiting kube-apiserver ready ..."
        sleep 3
    else
        break
    fi
done

# remove master taint
for NODE_NAME in $(kubectl get node -l "node-role.kubernetes.io/master" | grep Ready | awk '{print$1}'); do
    echo "[$(basename $0)]" "Remove taint node-role.kubernetes.io/master from $NODE_NAME"
    kubectl taint nodes $NODE_NAME node-role.kubernetes.io/master-
done

# add node label
CLUSTER_NODE_LABEL="node.kubernetes.io/node"
for NODE_NAME in $(kubectl get node | grep Ready | awk '{print$1}'); do
    echo "[$(basename $0)]" "Add label \"$CLUSTER_NODE_LABEL\" to $NODE_NAME"
    kubectl label nodes $NODE_NAME "$CLUSTER_NODE_LABEL="
done

NODE_MASTER_LABEL="node.kubernetes.io/master"
echo "[$(basename $0)]" "Set label \"$NODE_MASTER_LABEL\""
for NODE_NAME in $(kubectl get node -l "node-role.kubernetes.io/master" --show-labels | grep -v $NODE_MASTER_LABEL | grep Ready | awk '{print$1}'); do
    echo "[$(basename $0)]" "Add label \"$NODE_MASTER_LABEL\" to $NODE_NAME"
    kubectl label nodes $NODE_NAME "$NODE_MASTER_LABEL="
done

ESDB_MASTER_LABEL="polarbox.esdb.master"
echo "[$(basename $0)]" "Set label \"$ESDB_MASTER_LABEL\""
LABEL_NODE_NUMBER=$(kubectl get node -l "$ESDB_MASTER_LABEL" | grep Ready | wc -l)
if [[ $LABEL_NODE_NUMBER -lt 3 ]]; then
    while [ $LABEL_NODE_NUMBER -le 3 ]; do
        NODE_NAME=$(kubectl get node --show-labels | grep -v "$ESDB_MASTER_LABEL" | grep Ready | head -1 | awk '{print$1}')
        echo "[$(basename $0)]" "Add esdb master label to $NODE_NAME"
        kubectl label nodes $NODE_NAME "$ESDB_MASTER_LABEL="
        LABEL_NODE_NUMBER=$(kubectl get node -l "$ESDB_MASTER_LABEL" | wc -l)
    done
fi

# Todo: 当有node NotReady时，重复执行这个脚本就会出现打两个标的情况
MINIO_NODE_LABEL="polarbox.minio.enabled"
echo "[$(basename $0)]" "Set label \"$MINIO_NODE_LABEL\""
MINIO_NODE_NUMBER=$(kubectl get node -l "$MINIO_NODE_LABEL" | grep Ready | wc -l)
if [[ $MINIO_NODE_NUMBER -lt 1 ]]; then
    NODE_NAME=$(kubectl get node --show-labels | grep -v "$MINIO_NODE_LABEL" | grep master | head -1 | awk '{print$1}')
    echo "[$(basename $0)]" "Add minio node label to $NODE_NAME"
    kubectl label nodes $NODE_NAME "${MINIO_NODE_LABEL}=true" --overwrite
fi

NFS_NODE_LABEL="polarbox.nfs.enabled"
echo "[$(basename $0)]" "Set label \"$NFS_NODE_LABEL\""
for NODE_NAME in $(kubectl get node --show-labels | grep -v "$NFS_NODE_LABEL" | grep master | awk '{print$1}'); do
    echo "[$(basename $0)]" "Add nfs node label to $NODE_NAME"
    kubectl label nodes $NODE_NAME "${NFS_NODE_LABEL}=true" --overwrite
done

NETWORK_ADMIN_LABEL="polarbox.network.admin.enable"
echo "[$(basename $0)]" "Set label \"$NETWORK_ADMIN_LABEL\""
FIRST_NODE_NAME=$(kubectl get node | grep master | head -1 | awk '{print$1}')
for NODE_NAME in $(kubectl get node --show-labels | grep -v "$FIRST_NODE_NAME" | grep -v "$NETWORK_ADMIN_LABEL" | grep Ready | awk '{print$1}'); do
    echo "[$(basename $0)]" "Add network admin node label to $NODE_NAME"
    kubectl label nodes $NODE_NAME "${NETWORK_ADMIN_LABEL}=true" --overwrite
done

SMS_NODE_LABEL="nodetype"
echo "[$(basename $0)]" "Set label \"$SMS_NODE_LABEL\""
for NODE_NAME in $(kubectl get node --show-labels | grep -v "$SMS_NODE_LABEL" |  grep Ready | awk '{print$1}'); do
    echo "[$(basename $0)]" "Add sms agent node label to $NODE_NAME"
    kubectl label nodes $NODE_NAME "${SMS_NODE_LABEL}=agent" --overwrite
done