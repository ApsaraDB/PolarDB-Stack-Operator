#!/usr/bin/env bash

main() {
  echo "Start installing PolarDB Stack..."
}

show_usage() {
  cat <<-EOF
Usage: $0 [OPTION]
Install PolarDB Stack.
    -h, --help                     Display help and exit
EOF
}

main "$@"