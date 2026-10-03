package gsmmap

import (
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// SMS-CSI's sms-CAMEL-TDP-DataList is OPTIONAL in the ASN.1 and need be
// present only in the first segment of a segmented SMS-CSI (TS 29.002
// clause 17.7.1), so a wire SMS-CSI without it decodes, with a nil list.
func TestParseSMSCSIWithoutTDPDataList(t *testing.T) {
	cch := gsm_map.CamelCapabilityHandling(3)
	imsi := gsm_map.IMSI{0x21, 0x43, 0x65}
	csi := &gsm_map.SMSCSI{CamelCapabilityHandling: &cch}
	for name, vlr := range map[string]*gsm_map.VlrCamelSubscriptionInfo{
		"mo-sms-CSI": {MoSmsCSI: csi},
		"mt-sms-CSI": {MtSmsCSI: csi},
	} {
		t.Run(name, func(t *testing.T) {
			arg := gsm_map.InsertSubscriberDataArg{
				Imsi:                     &imsi,
				VlrCamelSubscriptionInfo: vlr,
			}
			data, err := arg.MarshalBER()
			if err != nil {
				t.Fatalf("MarshalBER: %v", err)
			}
			got, err := ParseInsertSubscriberData(data)
			if err != nil {
				t.Fatalf("ParseInsertSubscriberData: %v", err)
			}
			v := got.VlrCamelSubscriptionInfo
			parsed := v.MoSmsCSI
			if name == "mt-sms-CSI" {
				parsed = v.MtSmsCSI
			}
			if parsed == nil || parsed.SmsCAMELTDPDataList != nil || parsed.CamelCapabilityHandling == nil || *parsed.CamelCapabilityHandling != 3 {
				t.Errorf("%s = %+v, want no list and camelCapabilityHandling 3", name, parsed)
			}
		})
	}
}
