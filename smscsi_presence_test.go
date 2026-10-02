package gsmmap

import (
	"errors"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// SMS-CSI's sms-CAMEL-TDP-DataList is OPTIONAL in the ASN.1, but
// TS 29.002 clause 8.8.1 requires it in every SMS-CSI, so a wire SMS-CSI
// without it is rejected rather than decoded to a CSI that arms nothing.
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
			if _, err := ParseInsertSubscriberData(data); !errors.Is(err, ErrCamelSMSCSIMissingTDPData) {
				t.Errorf("ParseInsertSubscriberData: err = %v, want ErrCamelSMSCSIMissingTDPData", err)
			}
		})
	}
}
