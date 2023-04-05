#!/usr/bin/env bash

# Copyright (c) 2019 SAP SE or an SAP affiliate company. All rights reserved.
# Copyright 2022 OYE Network OÜ. All rights reserved.
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

set -e

. $(dirname "$0")/prepare-source-path.sh

LD_FLAGS="${LD_FLAGS:-$(${SOURCE_PATH}/hack/get-build-ld-flags.sh)}"

if [[ -z "$LOCAL_BUILD" ]]; then
  CGO_ENABLED=0 GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build \
    -a \
    -v \
    -ldflags "$LD_FLAGS" \
    -o ${BINARY_PATH}/terraform-provider-warren \
    ./main.go
else
  GOOS=$(go env GOOS) GOARCH=$(go env GOARCH) go build \
    -v \
    -ldflags "$LD_FLAGS" \
    -o ${BINARY_PATH}/terraform-provider-warren \
    ./main.go
fi

echo "Build script finished"
