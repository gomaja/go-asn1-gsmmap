package gsmmap

import (
	"fmt"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// Ext-SS-Info has one selected alternative: TS 29.002 V19.1.0 §17.7.1.
func TestMutationExtSSInfoAlternatives(t *testing.T) {
	cases := []struct {
		name  string
		value ExtSSInfo
	}{
		{"forwarding", ExtSSInfo{ForwardingInfo: &ExtForwInfo{SsCode: 0x21, ForwardingFeatureList: []ExtForwFeature{{SsStatus: HexBytes{1}}}}}},
		{"barring", ExtSSInfo{CallBarringInfo: &ExtCallBarInfo{SsCode: 0x92, CallBarringFeatureList: []ExtCallBarringFeature{{SsStatus: HexBytes{1}}}}}},
		{"cug", ExtSSInfo{CugInfo: &CUGInfo{}}},
		{"data", ExtSSInfo{SsData: &ExtSSData{SsCode: 0x11, SsStatus: HexBytes{1}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := &InsertSubscriberDataArg{ProvisionedSS: []ExtSSInfo{tc.value}}
			data, err := in.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			out, err := ParseInsertSubscriberData(data)
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "selected SS alternative", in, out)
		})
	}
}

// Ext-NoRepCondTime reserved ranges map to 5 and 30: TS 29.002 V19.1.0 §17.7.1.
func TestMutationExtNoReplyTimeMapping(t *testing.T) {
	cases := []struct{ wire, want int64 }{{1, 5}, {4, 5}, {5, 5}, {6, 6}, {30, 30}, {31, 30}, {100, 30}}
	for _, tc := range cases {
		t.Run(fmt.Sprint(tc.wire), func(t *testing.T) {
			value := tc.wire
			info := gsm_map.NewExtSSInfoForwardingInfo(gsm_map.ExtForwInfo{SsCode: gsm_map.SSCode{0x2a},
				ForwardingFeatureList: &gsm_map.ExtForwFeatureList{Values: []gsm_map.ExtForwFeature{{SsStatus: gsm_map.ExtSSStatus{1}, NoReplyConditionTime: &value}}}})
			w := &gsm_map.InsertSubscriberDataArg{ProvisionedSS: &gsm_map.ExtSSInfoList{Values: []gsm_map.ExtSSInfo{info}}}
			out, err := ParseInsertSubscriberData(strictBER(t, w))
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "mapped no reply time", int(tc.want), *out.ProvisionedSS[0].ForwardingInfo.ForwardingFeatureList[0].NoReplyConditionTime)
		})
	}
}
