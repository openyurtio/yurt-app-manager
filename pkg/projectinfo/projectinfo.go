/*
Copyright 2020 The OpenYurt Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package projectinfo

import (
	"fmt"
	"runtime"
)

var (
	projectPrefix = "yurt"
	labelPrefix   = "openyurt.io"
	gitVersion    = "v0.0.0"
	gitCommit     = "unknown"
	buildDate     = "1970-01-01T00:00:00Z"
)

// GetYurtAppManagerName returns name of tunnel
func GetYurtAppManagerName() string {
	return projectPrefix + "app-manager"
}

// normalizeGitCommit reserve 7 characters for gitCommit
func normalizeGitCommit(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}

	return commit
}

// Info contains version information.
type Info struct {
	GitVersion string `json:"gitVersion"`
	GitCommit  string `json:"gitCommit"`
	BuildDate  string `json:"buildDate"`
	GoVersion  string `json:"goVersion"`
	Compiler   string `json:"compiler"`
	Platform   string `json:"platform"`
}

// Get returns the overall codebase version.
func Get() Info {
	return Info{
		GitVersion: gitVersion,
		GitCommit:  normalizeGitCommit(gitCommit),
		BuildDate:  buildDate,
		GoVersion:  runtime.Version(),
		Compiler:   runtime.Compiler,
		Platform:   fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
	}
}
