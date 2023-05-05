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

if [[ $(uname) == 'Darwin' ]]; then
  READLINK_BIN="greadlink"
else
  READLINK_BIN="readlink"
fi

if [[ -z "${SOURCE_PATH}" ]]; then
  export SOURCE_PATH="$(${READLINK_BIN} -f $(dirname ${0})/..)"
else
  export SOURCE_PATH="$(${READLINK_BIN} -f "${SOURCE_PATH}")"
fi

if [[ -z "${BINARY_PATH}" ]]; then
  export BINARY_PATH="${SOURCE_PATH}/bin"
else
  export BINARY_PATH="$(${READLINK_BIN} -f "${BINARY_PATH}")"
  export PATH="${BINARY_PATH}:${PATH}"
fi

if [[ "${SOURCE_PATH}" != *"src/gitlab.com/warrenio/library/terraform-provider-warren" ]]; then
  SOURCE_SYMLINK_PATH="${SOURCE_PATH}/tmp/src/gitlab.com/warrenio/library/terraform-provider-warren"

  if [[ -d "${SOURCE_PATH}/tmp" && $TEST_CLEANUP == true ]]; then
    rm -rf "${SOURCE_PATH}/tmp"
  fi

  if [[ ! -d "${SOURCE_PATH}/tmp" ]]; then
    mkdir -p "${SOURCE_PATH}/tmp/src/gitlab.com/warrenio/library"
    ln -s "${SOURCE_PATH}" "${SOURCE_SYMLINK_PATH}"
  fi

  cd "${SOURCE_SYMLINK_PATH}"

  export GOPATH="${SOURCE_PATH}/tmp"
  export GOBIN="${SOURCE_PATH}/tmp/bin"
  export PATH="${GOBIN}:${PATH}"
fi
