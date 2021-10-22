#!/usr/bin/env bash

ENV_CONFIG=env.yaml

main() {
  parse_args
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
  install
}

install() {
  echo "Start installing PolarDB Stack..."
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
  cmd="helm install --dry-run --debug --generate-name ./polardb_stack_chart $network_interface_cmd $k8s_host_cmd $k8s_port_cmd $metabase_host_cmd $metabase_port_cmd $metabase_user_cmd $metabase_password_cmd $metabase_type_cmd $metabase_version_cmd"
  echo $cmd
  $cmd
}

main "$@"