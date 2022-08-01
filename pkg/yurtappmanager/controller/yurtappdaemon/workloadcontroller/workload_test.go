package workloadcontroller

const (
	failed  = "\u2717"
	succeed = "\u2713"
)

//func TestGetRevision(t *testing.T) {
//	tests := []struct {
//		name   string
//		w      *Workload
//		expect string
//	}{
//		{
//			"normal",
//			&Workload{
//				Spec: WorkloadSpec{},
//			},
//			"false",
//		},
//	}
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			t.Parallel()
//			t.Logf("\tTestCase: %s", tt.name)
//			t.Logf("expect: %v", tt.expect)
//			{
//				get := tt.w.Spec.Ref.GetLabels()[unitv1alpha1.ControllerRevisionHashLabelKey]
//				t.Logf("get: %v, expect: %v", get, tt.expect)
//				if !reflect.DeepEqual(get, tt.expect) {
//					//t.Fatalf("\t%s\texpect %v, but get %v", failed, expect, get)
//				}
//				t.Logf("\t%s\texpect %v, get %v", succeed, tt.expect, get)
//
//			}
//		})
//	}
//}
