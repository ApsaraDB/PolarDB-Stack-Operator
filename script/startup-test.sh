#!/bin/sh

PACKAGES=`go list ./... | grep -v apis | grep -v client`

echo "" > coverage.out
echo "mode: set" > coverage-all.out
TEST_FAILED=0;
export KUBECONFIG=`pwd`/test/data/k8s/config
for pkg in ${PACKAGES}; do
  TEST_MODE=true go test -gcflags='all=-N -l' -coverpkg=./... -coverprofile=coverage.out -covermode=set $pkg || TEST_FAILED=1;
  tail -n +2 coverage.out >> coverage-all.out;
done;
if [ "$TEST_FAILED" -eq "0" ] ; then
  go tool cover -func=coverage-all.out |grep total
else
  exit 1
fi
