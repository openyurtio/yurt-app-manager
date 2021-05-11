package uniteddaemonset

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/validation"
)

func getWorkloadPrefix(controllerName, nodepoolName string) string {
	prefix := fmt.Sprintf("%s-%s-", controllerName, nodepoolName)
	if len(validation.NameIsDNSSubdomain(prefix, true)) != 0 {
		prefix = fmt.Sprintf("%s-", controllerName)
	}
	return prefix
}
