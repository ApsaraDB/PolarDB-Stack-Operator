#!/usr/bin/env bash

main() {
  echo "Start installing PolarDB Stack..."
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
  if [[ $# -eq 0 ]]; then
    echo "Invalid usage!!!"
    show_usage
    exit 1
  fi
}

main "$@"