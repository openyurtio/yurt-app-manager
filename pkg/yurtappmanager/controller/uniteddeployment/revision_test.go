/*
Copyright 2020 The OpenYurt Authors.
Copyright 2019 The Kruise Authors.
Copyright 2017 The Kubernetes Authors.

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

package uniteddeployment

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	apps "k8s.io/api/apps/v1"
)

var _ = Describe("Revision", func() {

	It("revisions length is 0, expect get 1", func() {
		revisions := []*apps.ControllerRevision{}
		revisionNum := nextRevision(revisions)
		Expect(revisionNum).To(Equal(int64(1)))

	})

	It("last revision number is 1, expect get 2", func() {
		revisions := []*apps.ControllerRevision{{Revision: 1}}
		revisionNum := nextRevision(revisions)
		Expect(revisionNum).To(Equal(int64(2)))
	})

	It("last revision number is 2, expect get 3", func() {
		revisions := []*apps.ControllerRevision{{Revision: 2}}
		revisionNum := nextRevision(revisions)
		Expect(revisionNum).To(Equal(int64(3)))
	})

	It("last revision number is 10, expect get 11", func() {
		revisions := []*apps.ControllerRevision{{Revision: 10}}
		revisionNum := nextRevision(revisions)
		Expect(revisionNum).To(Equal(int64(11)))
	})

})
