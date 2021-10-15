#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# corresponding to go mod init <module>
MODULE=gitlab.alibaba-inc.com/polar-as/polar-mpd-controller
# api package
APIS_PKG=apis
# generated output package
OUTPUT_PKG=
# group-version such as foo:v1alpha1
GROUP_VERSION=mpd:v1

SCRIPT_ROOT=$(dirname "${BASH_SOURCE[0]}")/..
CODEGEN_PKG=${CODEGEN_PKG:-$(cd "${SCRIPT_ROOT}"; ls -d -1 ~/go/pkg/mod/k8s.io/code-generator@v0.0.0-20190612205613-18da4a14b22b 2>/dev/null || echo ../code-generator)}

# generate the code with:
# --output-base    because this script should also be able to run inside the vendor dir of
#                  k8s.io/kubernetes. The output-base is needed for the generators to output into the vendor dir
#                  instead of the $GOPATH directly. For normal projects this can be dropped.
bash "${CODEGEN_PKG}"/generate-groups.sh "client" \
${MODULE}/${OUTPUT_PKG} ${MODULE}/${APIS_PKG} \
${GROUP_VERSION} \
--go-header-file "${SCRIPT_ROOT}"/hack/boilerplate.go.txt \
--output-base "${SCRIPT_ROOT}/../../.." \

