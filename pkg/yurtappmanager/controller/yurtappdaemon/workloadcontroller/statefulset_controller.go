package workloadcontroller

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type StatefulSetControllor struct {
	client.Client

	scheme *runtime.Scheme
}

// var _ WorkloadControllor = &StatefulSetControllor{}
