package gsmmap

import "testing"

// TS 29.002 V19.1.0 §17.7.1: GroupId is filled with six TBCD fillers when LongGroupId is present.
func TestMutationSubscriberGroupFiller(t *testing.T) {
	for _, entry := range []VoiceBroadcastData{{GroupId: "123456"}, {LongGroupId: "12345678"}} {
		in := &InsertSubscriberDataArg{VbsSubscriptionData: VBSDataList{entry}}
		data, err := in.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		got, err := ParseInsertSubscriberData(data)
		if err != nil {
			t.Fatalf("Parse: %v", err)
		}
		wantEqual(t, "broadcast", in.VbsSubscriptionData, got.VbsSubscriptionData)
	}
}
