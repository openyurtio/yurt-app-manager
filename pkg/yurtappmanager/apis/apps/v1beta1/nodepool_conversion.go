/*
Copyright 2022 The OpenYurt Authors.

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
// +kubebuilder:docs-gen:collapse=Apache License

package v1beta1

/*
For imports, we'll need the controller-runtime
[`conversion`](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/conversion?tab=doc)
package, plus the API version for our hub type (v1), and finally some of the
standard packages.
*/
import (
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/conversion"

	"github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
)

// +kubebuilder:docs-gen:collapse=Imports

func (src *NodePool) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.NodePool)

	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.Type = v1alpha1.NodePoolType(src.Spec.Type)
	dst.Spec.Selector = src.Spec.Selector
	dst.Spec.Annotations = src.Spec.Annotations
	dst.Spec.Taints = src.Spec.Taints

	dst.Status.ReadyNodeNum = src.Status.ReadyNodeNum
	dst.Status.UnreadyNodeNum = src.Status.UnreadyNodeNum
	dst.Status.Nodes = src.Status.Nodes

	klog.Infof("convert from v1beta1 to v1alpha1 for %s", dst.Name)

	return nil
}

func (dst *NodePool) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha1.NodePool)

	dst.ObjectMeta = src.ObjectMeta

	dst.Spec.Type = NodePoolType(src.Spec.Type)
	dst.Spec.Selector = src.Spec.Selector
	dst.Spec.Annotations = src.Spec.Annotations
	dst.Spec.Taints = src.Spec.Taints

	dst.Status.ReadyNodeNum = src.Status.ReadyNodeNum
	dst.Status.UnreadyNodeNum = src.Status.UnreadyNodeNum
	dst.Status.Nodes = src.Status.Nodes

	klog.Infof("convert from v1alpha1 to v1beta1 for %s", dst.Name)
	return nil
}
