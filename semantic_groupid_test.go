// semantic_groupid_test.go
//
// The GroupId filler rule of VoiceBroadcastData and VoiceGroupCallData,
// checked through the public Parse and Marshal entry points.
package gsmmap

import (
	"errors"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// 3GPP TS 29.002 V19.1.0 §17.7.1 VoiceBroadcastData and VoiceGroupCallData:
// "groupId shall be filled with six TBCD fillers (1111) if the longGroupId
// is present".
func TestParseGroupIdWithLongGroupId(t *testing.T) {
	long := gsm_map.LongGroupId{0x21, 0x43, 0x65, 0x87}
	for _, tc := range []struct {
		name    string
		groupID gsm_map.GroupId
		want    error
	}{
		{"six fillers", gsm_map.GroupId{0xFF, 0xFF, 0xFF}, nil},
		{"two digits", gsm_map.GroupId{0x21, 0xFF, 0xFF}, ErrGroupIdFillerRequired},
		{"one digit", gsm_map.GroupId{0xF1, 0xFF, 0xFF}, ErrGroupIdFillerRequired},
		{"six digits", gsm_map.GroupId{0x21, 0x43, 0x65}, ErrGroupIdFillerRequired},
	} {
		lg := long
		vbs := gsm_map.VBSDataList{Values: []gsm_map.VoiceBroadcastData{{Groupid: tc.groupID, LongGroupId: &lg}}}
		vgcs := gsm_map.VGCSDataList{Values: []gsm_map.VoiceGroupCallData{{GroupId: tc.groupID, LongGroupId: &lg}}}
		for where, w := range map[string]*gsm_map.InsertSubscriberDataArg{
			"VBS":  {VbsSubscriptionData: &vbs},
			"VGCS": {VgcsSubscriptionData: &vgcs},
		} {
			t.Run(where+"/"+tc.name, func(t *testing.T) {
				data, err := w.MarshalBER()
				if err != nil {
					t.Fatalf("MarshalBER: %v", err)
				}
				got, err := ParseInsertSubscriberData(data)
				if !errors.Is(err, tc.want) {
					t.Fatalf("ParseInsertSubscriberData: err = %v, want %v", err, tc.want)
				}
				if err != nil {
					return
				}
				// The parsed value marshals back to the same octets.
				out, err := got.Marshal()
				if err != nil {
					t.Fatalf("Marshal: %v", err)
				}
				if string(out) != string(data) {
					t.Errorf("Marshal = %x, want %x", out, data)
				}
			})
		}
	}
}

// Without a LongGroupId the GroupId carries its digits.
func TestParseGroupIdWithoutLongGroupId(t *testing.T) {
	vbs := gsm_map.VBSDataList{Values: []gsm_map.VoiceBroadcastData{{Groupid: gsm_map.GroupId{0x21, 0xFF, 0xFF}}}}
	data, err := (&gsm_map.InsertSubscriberDataArg{VbsSubscriptionData: &vbs}).MarshalBER()
	if err != nil {
		t.Fatalf("MarshalBER: %v", err)
	}
	got, err := ParseInsertSubscriberData(data)
	if err != nil {
		t.Fatalf("ParseInsertSubscriberData: %v", err)
	}
	if g := got.VbsSubscriptionData[0].GroupId; g != "12" {
		t.Errorf("GroupId = %q, want %q", g, "12")
	}
}
