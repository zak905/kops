#!/bin/bash

# Copyright 2019 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script will be run when building process starts to
# generate key-value information that represents the status of the
# workspace. The output should be like
#
# KEY1 VALUE1
# KEY2 VALUE2
#
# If the script exits with non-zero code, it's considered as a failure
# and the output will be discarded.

# The code below presents an implementation that works for git repository
git_rev=$(git rev-parse HEAD 2>/dev/null)
echo "BUILD_SCM_REVISION ${git_rev}"

# Check whether there are any uncommited changes
git diff-index --quiet HEAD -- 2>/dev/null
if [[ $? == 0 ]];
then
    tree_status="Clean"
else
    tree_status="Modified"
fi
echo "BUILD_SCM_STATUS ${tree_status}"

VERSION=`tools/get_version.sh | grep VERSION | awk '{print $2}'`
echo "STABLE_KOPS_VERSION ${VERSION}"

# + is valid in semver, but not in docker tags. Fixup CI versions.
# Note that this mirrors the logic in DefaultProtokubeImageName
PROTOKUBE_TAG=${VERSION/+/-}
echo "STABLE_PROTOKUBE_TAG ${PROTOKUBE_TAG}"



if [[ -z "${DOCKER_REGISTRY}" ]]; then
  DOCKER_REGISTRY="index.docker.io"
fi
if [[ -z "${DOCKER_IMAGE_PREFIX}" ]]; then
  DOCKER_IMAGE_PREFIX=`whoami`/
fi
echo "STABLE_DOCKER_REGISTRY ${DOCKER_REGISTRY}"
echo "STABLE_DOCKER_IMAGE_PREFIX ${DOCKER_IMAGE_PREFIX}"

if [[ -z "${KOPS_CONTROLLER_TAG}" ]]; then
  KOPS_CONTROLLER_TAG="${PROTOKUBE_TAG}"
fi
echo "STABLE_KOPS_CONTROLLER_TAG ${KOPS_CONTROLLER_TAG}"

if [[ -z "${DNS_CONTROLLER_TAG}" ]]; then
  DNS_CONTROLLER_TAG="${PROTOKUBE_TAG}"
fi
echo "STABLE_DNS_CONTROLLER_TAG ${DNS_CONTROLLER_TAG}"

if [[ -z "${KUBE_APISERVER_HEALTHCHECK_TAG}" ]]; then
  KUBE_APISERVER_HEALTHCHECK_TAG="${PROTOKUBE_TAG}"
fi
echo "STABLE_KUBE_APISERVER_HEALTHCHECK_TAG ${KUBE_APISERVER_HEALTHCHECK_TAG}"

