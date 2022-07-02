#!/bin/bash -l
# Copyright 2020 The OpenYurt Authors.
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

if [[ -n "$SSH_DEPLOY_KEY" ]]
then
  mkdir -p ~/.ssh
  echo "$SSH_DEPLOY_KEY" > ~/.ssh/id_rsa
  chmod 600 ~/.ssh/id_rsa
fi

echo "git clone"
cd ..
git config --global user.email "openyurt-bot@openyurt.io"
git config --global user.name "openyurt-bot"
git clone --single-branch --depth 1 git@github.com:openyurtio/yurt-app-manager-api.git yurt-app-manager-api

echo "clear yurt-app-manager-api api/"
rm -r yurt-app-manager-api/pkg/yurtappmanager/apis/*

echo "clear yurt-app-manager-api client/"
rm -r yurt-app-manager-api/pkg/yurtappmanager/client/*

echo "update yurt-app-manager-api api/"
cp -R yurt-app-manager/pkg/yurtappmanager/apis/* yurt-app-manager-api/pkg/yurtappmanager/apis/

echo "update yurt-app-manager-api client/"
cp -R yurt-app-manager/pkg/yurtappmanager/client/* yurt-app-manager-api/pkg/yurtappmanager/client/

echo "change import path"
find ./yurt-app-manager-api -type f -name "*.go" -print0 | xargs -0 sed -i 's|github.com/openyurtio/yurt-app-manager/|github.com/openyurtio/yurt-app-manager-api/|g' ./yurt-app-manager-api/pkg/yurtappmanager/apis/addtoscheme_apps_v1alpha1.go

echo "test api"
cd yurt-app-manager-api
go mod tidy

echo "push to yurt-app-manager-api"
echo "version: $VERSION, commit: $COMMIT_ID, tag: $TAG"

if git diff --quiet
then
  echo "nothing need to push, finished!"
else
  git add .
  git commit -m "align with yurt-app-manager-$VERSION from commit $COMMIT_ID"
  git tag "$VERSION"
  git push origin main
fi

if [[ $TAG == v* ]] ;
then
    echo "push tag: TAG"
    git push origin "$TAG"
fi