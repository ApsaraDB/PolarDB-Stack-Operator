#!/usr/bin/env bash

ENV_CONFIG=env.yaml

main() {
  parse_args
  echo "Start installing PolarDB Stack..."
  update_config
  set_node_label
  install_multipath
  install_supervisor
  install_agent
  agent_ini
  agent_conf
  install_pfs
}

show_usage() {
  cat <<-EOF
Usage: $0 [OPTION]
Install PolarDB Stack.
    -h, --help                     Display help and exit
    -c, --config                   The config for PolarDB Stack, default is env.yaml
EOF
}

parse_args() {
  if [[ $# -ne 0 ]]; then
    local args=$(getopt \
      -o "h,c" \
      --long "help,config" \
      -- "$@")
    if [[ $? -ne 0 ]]; then
      err "Invalid options"
    fi
    eval set -- "${args}"
    while true; do
      case "$1" in
      -h | --help)
        show_usage
        exit 0
        ;;
      esac
    done
  fi
}

update_config() {
  if [ ! -f "$ENV_CONFIG" ]; then
    echo "Config file $ENV_CONFIG not exist, exit."
    exit 1
  fi
  network_interface=$(grep "interface" $ENV_CONFIG | awk '{print $2}')
  if [ ! -z "$network_interface" ]; then
    network_interface_cmd="--set network.interface=$network_interface "
    echo "set config $network_interface_cmd"
  fi
  k8s_host=$(grep "k8s" $ENV_CONFIG -A2 | grep "host" | awk '{print $2}')
  if [ ! -z "$k8s_host" ]; then
    k8s_host_cmd="--set KUBERNETES_SERVICE_HOST=$k8s_host "
    echo "set config $k8s_host_cmd"
  fi
  k8s_port=$(grep "k8s" $ENV_CONFIG -A2 | grep "port" | awk '{print $2}')
  if [ ! -z "$k8s_port" ]; then
    k8s_port_cmd="--set KUBERNETES_SERVICE_PORT=$k8s_port "
    echo "set config $k8s_port_cmd"
  fi
  metabase_host=$(grep "metabase" $ENV_CONFIG -A6 | grep "host" | awk '{print $2}')
  if [ ! -z "$metabase_host" ]; then
    metabase_host_cmd="--set metabase.host=$metabase_host "
    echo "set config $metabase_host_cmd"
  fi
  metabase_port=$(grep "metabase" $ENV_CONFIG -A6 | grep "port" | awk '{print $2}')
  if [ ! -z "$metabase_port" ]; then
    metabase_port_cmd="--set metabase.port=$metabase_port "
    echo "set config $metabase_port_cmd"
  fi
  metabase_user=$(grep "metabase" $ENV_CONFIG -A6 | grep "user" | awk '{print $2}')
  if [ ! -z "$metabase_user" ]; then
    metabase_user_cmd="--set metabase.user=$metabase_user "
    echo "set config $metabase_user_cmd"
  fi
  metabase_password=$(grep "metabase" $ENV_CONFIG -A6 | grep "password" | awk '{print $2}')
  if [ ! -z "$metabase_password" ]; then
    metabase_password_cmd="--set metabase.password=$metabase_password "
    echo "set config $metabase_password_cmd"
  fi
  metabase_type=$(grep "metabase" $ENV_CONFIG -A6 | grep "type" | awk '{print $2}')
  if [ ! -z "$metabase_type" ]; then
    metabase_type_cmd="--set metabase.type=$metabase_type "
    echo "set config $metabase_type_cmd"
  fi
  metabase_version=$(grep "metabase" $ENV_CONFIG -A6 | grep "version" | awk '{print $2}')
  if [ ! -z "$metabase_version" ]; then
    metabase_version_cmd="--set metabase.version=$metabase_version "
    echo "set config $metabase_version_cmd"
  fi
  cmd="helm install --dry-run --debug --generate-name ./polardb-stack-chart $network_interface_cmd $k8s_host_cmd $k8s_port_cmd $metabase_host_cmd $metabase_port_cmd $metabase_user_cmd $metabase_password_cmd $metabase_type_cmd $metabase_version_cmd"
  echo $cmd
  $cmd
}

set_node_label() {
  ./script/set_labels.sh
}

install_multipath() {
  network_interface=$(grep "interface" $ENV_CONFIG | awk '{print $2}')
  if [ -z "$network_interface" ]; then
    echo "No network interface configured, exit."
    exit 1
  fi

  ips=( `sed -n '/dbm_hosts/,/network/p' $ENV_CONFIG | grep 'ip' | awk '{print $3}'` )
  cnt=${#ips[@]}

  for ((i=0;i<$cnt;i++));
  do
    base_cmd="ssh root@${ips[$i]}"
    $base_cmd yum install -y device-mapper-multipath
    $base_cmd "cat <<EOF >/etc/multipath.conf
defaults {
	user_friendly_names no
	find_multipaths no
}
blacklist {
devnode "^sda$"
}
EOF"
    $base_cmd systemctl enable multipathd.service
    $base_cmd systemctl start multipathd.service
  done
}

install_supervisor() {
  network_interface=$(grep "interface" $ENV_CONFIG | awk '{print $2}')
  if [ -z "$network_interface" ]; then
    echo "No network interface configured, exit."
    exit 1
  fi

  ips=( `sed -n '/dbm_hosts/,/network/p' $ENV_CONFIG | grep 'ip' | awk '{print $3}'` )
  cnt=${#ips[@]}

  for ((i=0;i<$cnt;i++));
  do
    base_cmd="ssh root@${ips[$i]}"
    $base_cmd yum install -y supervisor
    $base_cmd systemctl enable supervisord
    $base_cmd systemctl start supervisord
  done
}

install_agent() {
  wget https://github.com/ApsaraDB/PolarDB-Stack-Storage/releases/download/v1.0.0/sms-agent
  mkdir -p /home/a/project/t-polardb-sms-agent/bin/
  cp sms-agent /home/a/project/t-polardb-sms-agent/bin/polardb-sms-agent
  chmod u+x /home/a/project/t-polardb-sms-agent/bin/polardb-sms-agent

  network_interface=$(grep "interface" $ENV_CONFIG | awk '{print $2}')
  if [ -z "$network_interface" ]; then
    echo "No network interface configured, exit."
    exit 1
  fi

  ips=( `sed -n '/dbm_hosts/,/network/p' $ENV_CONFIG | grep 'ip' | awk '{print $3}'` )
  cnt=${#ips[@]}

  for ((i=0;i<$cnt;i++));
  do
    ssh root@${ips[$i]} mkdir -p /home/a/project/t-polardb-sms-agent/bin
    scp /home/a/project/t-polardb-sms-agent/bin/polardb-sms-agent root@${ips[$i]}:/home/a/project/t-polardb-sms-agent/bin/polardb-sms-agent
  done
}

agent_ini() {

  AGENT_INI="/etc/supervisord.d/polardb-sms-agent.ini"

  network_interface=$(grep "interface" $ENV_CONFIG | awk '{print $2}')
  if [ -z "$network_interface" ]; then
    echo "No network interface configured, exit."
    exit 1
  fi

  ips=( `sed -n '/dbm_hosts/,/network/p' $ENV_CONFIG | grep 'ip' | awk '{print $3}'` )
  cnt=${#ips[@]}

  for ((i=0;i<$cnt;i++));
  do
    base_cmd="ssh root@${ips[$i]}"

    NODE_IP=$($base_cmd ifconfig $network_interface | grep netmask | awk '{print $2}')

    echo $base_cmd
    $base_cmd touch $AGENT_INI
    $base_cmd "cat <<EOF >$AGENT_INI
[program:polardb-sms-agent]
command=/home/a/project/t-polardb-sms-agent/bin/polardb-sms-agent --port=18888 --node-ip=$NODE_IP --node-id=%(host_node_name)s
process_name=%(program_name)s
startretries=1000
autorestart=unexpected
autostart=true
EOF"
    $base_cmd cat $AGENT_INI

  done
}

agent_conf() {

  AGENT_CONF="/etc/polardb-sms-agent.conf"

  network_interface=$(grep "interface" $ENV_CONFIG | awk '{print $2}')
  if [ -z "$network_interface" ]; then
    echo "No network interface configured, exit."
    exit 1
  fi

  ips=( `sed -n '/dbm_hosts/,/network/p' $ENV_CONFIG | grep 'ip' | awk '{print $3}'` )
  cnt=${#ips[@]}

  for ((i=0;i<$cnt;i++));
  do
    base_cmd="ssh root@${ips[$i]}"

    $base_cmd "cat <<EOF >$AGENT_CONF
blacklist {
    attachlist {
    }
    locallist {
    }
}
EOF"
  done
}

install_pfs() {
  pfsd=t-pfsd-opensource-1.2.41-1.el7.x86_64.rpm
  wget https://github.com/ApsaraDB/polardb-file-system/releases/download/pfsd4pg-release-1.2.41-20211018/$pfsd
  rpm -ivh $pfsd

  network_interface=$(grep "interface" $ENV_CONFIG | awk '{print $2}')
  if [ -z "$network_interface" ]; then
    echo "No network interface configured, exit."
    exit 1
  fi

  ips=( `sed -n '/dbm_hosts/,/network/p' $ENV_CONFIG | grep 'ip' | awk '{print $3}'` )
  cnt=${#ips[@]}

  for ((i=0;i<$cnt;i++));
  do
    scp $pfsd root@${ips[$i]}:/tmp/$pfsd
    ssh root@${ips[$i]} rpm -ivh /tmp/$pfsd
  done
}

main "$@"