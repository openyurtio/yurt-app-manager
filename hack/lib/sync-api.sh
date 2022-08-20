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

## sync to code repo
destRepo=api
destModule=apps

git clone --single-branch --depth 1 git@github.com:openyurtio/${destRepo}.git ${destRepo}

if [ -d "${destRepo}/pkg/${destModule}/apis" ]
then
    echo "apps apis exists, remove it"
    rm -r ${destRepo}/pkg/${destModule}/apis/*
else
    mkdir -p ${destRepo}/pkg/${destModule}/apis
fi

if [ -d "${destRepo}/pkg/${destModule}/client" ]
then
    echo "client exists, remove it"
    rm -r ${destRepo}/pkg/${destModule}/client/*
else
    mkdir -p ${destRepo}/pkg/${destModule}/client
fi

echo "update ${destRepo} apis/"
cp -R yurt-app-manager/pkg/yurtappmanager/apis/* ${destRepo}/pkg/${destModule}/apis/
# remove controller depends functions
rm -r ${destRepo}/pkg/${destModule}/apis/apps/v1alpha1/defaults.go

echo "update ${destRepo} client/"
cp -R yurt-app-manager/pkg/yurtappmanager/client/* ${destRepo}/pkg/${destModule}/client/

echo "change import paths, and change them"
find ./${destRepo} -type f -name "*.go" -print0 | xargs -0 sed -i 's|github.com/openyurtio/yurt-app-manager/|github.com/openyurtio/api/|g'

#echo "test api"
#cd ${destRepo}
#go mod tidy
#make test

echo "push to ${destRepo}"
echo "version: $VERSION, commit: $COMMIT_ID, tag: $TAG"

if [ -z "$(git status --porcelain)" ]; then
  echo "nothing need to push, finished!"
else
  git add .
  git commit -m "align with yurt-app-manager commit $COMMIT_ID"
  git tag "$VERSION"
  git push origin main
fi

if [[ $TAG == v* ]] ;
then
    echo "push tag: TAG"
    git push origin "$TAG"
fi