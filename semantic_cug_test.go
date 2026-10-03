// semantic_cug_test.go
//
// The empty CUG-SubscriptionList, checked through the public Marshal and
// Parse entry points.
package gsmmap

import "testing"

// cug-SubscriptionList is mandatory in CUG-Info but may be empty (3GPP TS
// 29.002 V19.1.0 §17.7.1), so a nil list marshals as an empty present list.
func TestCUGInfoEmptySubscriptionList(t *testing.T) {
	for name, list := range map[string][]CUGSubscription{"nil": nil, "empty": {}} {
		t.Run(name, func(t *testing.T) {
			in := &InsertSubscriberDataArg{ProvisionedSS: []ExtSSInfo{{CugInfo: &CUGInfo{CugSubscriptionList: list}}}}
			data, err := in.Marshal()
			if err != nil {
				t.Fatalf("Marshal: %v", err)
			}
			got, err := ParseInsertSubscriberData(data)
			if err != nil {
				t.Fatalf("ParseInsertSubscriberData: %v", err)
			}
			semWantEqual(t, "ProvisionedSS", []ExtSSInfo{{CugInfo: &CUGInfo{}}}, got.ProvisionedSS)
		})
	}
}
