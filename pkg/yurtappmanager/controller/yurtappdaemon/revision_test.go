package yurtappdaemon

import (
	appsalphav1 "github.com/openyurtio/yurt-app-manager/pkg/yurtappmanager/apis/apps/v1alpha1"
	apps "k8s.io/api/apps/v1"
	"reflect"
	"testing"
)

//const (
//	failed  = "\u2717"
//	succeed = "\u2713"
//)

func TestNextRevision(t *testing.T) {
	tests := []struct {
		name      string
		revisions []*apps.ControllerRevision
		expect    int64
	}{
		{
			"zero",
			[]*apps.ControllerRevision{},
			1,
		},
		{
			"normal",
			[]*apps.ControllerRevision{
				{
					Revision: 1,
				},
			},
			2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)
			{
				get := nextRevision(tt.revisions)

				if !reflect.DeepEqual(get, tt.expect) {
					t.Fatalf("\t%s\texpect %v, but get %v", failed, tt.expect, get)
				}
				t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)

			}
		})
	}
}

func TestGetYurtAppDaemonPatch(t *testing.T) {

	tests := []struct {
		name string
		ud   *appsalphav1.YurtAppDaemon
	}{
		{
			"normal",
			&appsalphav1.YurtAppDaemon{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			t.Logf("\tTestCase: %s", tt.name)
			{
				get, _ := getYurtAppDaemonPatch(tt.ud)

				//if !reflect.DeepEqual(get, expect) {
				//	t.Fatalf("\t%s\texpect %v, but get %v", failed, expect, get)
				//}
				t.Logf("\t%s\tget %v", succeed, get)

			}
		})
	}
}
