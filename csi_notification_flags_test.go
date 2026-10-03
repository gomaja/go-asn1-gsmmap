package gsmmap

import (
	"reflect"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// 3GPP TS 29.002 V19.1.0 §17.7.1 has the same rule for every CSI, e.g.
// "notificationToCSE and csi-Active shall not be present when D-CSI is sent
// to VLR/GMSC. They may only be included in ATSI/ATM ack/NSDC message." The
// package implements none of AnyTimeSubscriptionInterrogation,
// AnyTimeModification or NoteSubscriberDataModified: its CSIs travel in
// InsertSubscriberData (to the VLR and SGSN) and SendRoutingInfo-Res (to the
// GMSC). The public CSI types have no field for the two flags, so Marshal
// never sends them, and Parse drops them.

func TestCSITypesHaveNoNotificationFlags(t *testing.T) {
	for _, v := range []any{OCSI{}, TCSI{}, DCSI{}, SSCSI{}, MCSI{}, SMSCSI{}, GPRSCSI{}, MGCSI{}} {
		typ := reflect.TypeOf(v)
		for _, f := range []string{"NotificationToCSE", "CsiActive"} {
			if _, ok := typ.FieldByName(f); ok {
				t.Errorf("%s has field %s", typ.Name(), f)
			}
		}
	}
}

func setFlags(n, a **struct{}) {
	*n, *a = &struct{}{}, &struct{}{}
}

// allCSIISD is a camelISD carrying every CSI of the VLR and SGSN CAMEL
// subscription info.
func allCSIISD() *InsertSubscriberDataArg {
	a := camelISD()
	a.VlrCamelSubscriptionInfo.SsCSI = &SSCSI{SsEventList: []SsCode{SsCodeECT}, GsmSCFAddress: testGsmSCF}
	a.VlrCamelSubscriptionInfo.MCSI = &MCSI{MobilityTriggers: []MMCode{MMCodeIMSIAttach}, ServiceKey: 8, GsmSCFAddress: testGsmSCF}
	a.SgsnCAMELSubscriptionInfo.MgCsi = &MGCSI{MobilityTriggers: []MMCode{MMCodeGPRSAttach}, ServiceKey: 9, GsmSCFAddress: testGsmSCF}
	return a
}

func TestParseDropsCSINotificationFlags(t *testing.T) {
	t.Run("InsertSubscriberData", func(t *testing.T) {
		w := isdWire(t, allCSIISD())
		v, s := w.VlrCamelSubscriptionInfo, w.SgsnCAMELSubscriptionInfo
		setFlags(&v.OCSI.NotificationToCSE, &v.OCSI.CsiActive)
		setFlags(&v.SsCSI.NotificationToCSE, &v.SsCSI.CsiActive)
		setFlags(&v.MCSI.NotificationToCSE, &v.MCSI.CsiActive)
		setFlags(&v.DCSI.NotificationToCSE, &v.DCSI.CsiActive)
		setFlags(&v.VtCSI.NotificationToCSE, &v.VtCSI.CsiActive)
		setFlags(&v.MoSmsCSI.NotificationToCSE, &v.MoSmsCSI.CsiActive)
		setFlags(&v.MtSmsCSI.NotificationToCSE, &v.MtSmsCSI.CsiActive)
		setFlags(&s.GprsCSI.NotificationToCSE, &s.GprsCSI.CsiActive)
		setFlags(&s.MoSmsCSI.NotificationToCSE, &s.MoSmsCSI.CsiActive)
		setFlags(&s.MtSmsCSI.NotificationToCSE, &s.MtSmsCSI.CsiActive)
		setFlags(&s.MgCsi.NotificationToCSE, &s.MgCsi.CsiActive)
		got := parseISD(t, w)
		wantEqual(t, "InsertSubscriberDataArg", allCSIISD(), got)

		data, err := got.Marshal()
		if err != nil {
			t.Fatalf("Marshal: %v", err)
		}
		var again gsm_map.InsertSubscriberDataArg
		if err := again.UnmarshalBER(data); err != nil {
			t.Fatalf("UnmarshalBER: %v", err)
		}
		v, s = again.VlrCamelSubscriptionInfo, again.SgsnCAMELSubscriptionInfo
		for name, p := range map[string]*struct{}{
			"O-CSI notificationToCSE": v.OCSI.NotificationToCSE, "O-CSI csiActive": v.OCSI.CsiActive,
			"SS-CSI notificationToCSE": v.SsCSI.NotificationToCSE, "SS-CSI csi-Active": v.SsCSI.CsiActive,
			"M-CSI notificationToCSE": v.MCSI.NotificationToCSE, "M-CSI csi-Active": v.MCSI.CsiActive,
			"D-CSI notificationToCSE": v.DCSI.NotificationToCSE, "D-CSI csi-Active": v.DCSI.CsiActive,
			"VT-CSI notificationToCSE": v.VtCSI.NotificationToCSE, "VT-CSI csi-Active": v.VtCSI.CsiActive,
			"MO-SMS-CSI notificationToCSE": v.MoSmsCSI.NotificationToCSE, "MO-SMS-CSI csi-Active": v.MoSmsCSI.CsiActive,
			"MT-SMS-CSI notificationToCSE": v.MtSmsCSI.NotificationToCSE, "MT-SMS-CSI csi-Active": v.MtSmsCSI.CsiActive,
			"GPRS-CSI notificationToCSE": s.GprsCSI.NotificationToCSE, "GPRS-CSI csi-Active": s.GprsCSI.CsiActive,
			"SGSN MO-SMS-CSI notificationToCSE": s.MoSmsCSI.NotificationToCSE, "SGSN MO-SMS-CSI csi-Active": s.MoSmsCSI.CsiActive,
			"SGSN MT-SMS-CSI notificationToCSE": s.MtSmsCSI.NotificationToCSE, "SGSN MT-SMS-CSI csi-Active": s.MtSmsCSI.CsiActive,
			"MG-CSI notificationToCSE": s.MgCsi.NotificationToCSE, "MG-CSI csi-Active": s.MgCsi.CsiActive,
		} {
			if p != nil {
				t.Errorf("Marshal sent %s", name)
			}
		}
	})
	t.Run("SendRoutingInfoRes", func(t *testing.T) {
		w := sriRespWire(t, gmscCamel())
		g := w.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo
		setFlags(&g.TCSI.NotificationToCSE, &g.TCSI.CsiActive)
		setFlags(&g.OCSI.NotificationToCSE, &g.OCSI.CsiActive)
		setFlags(&g.DCsi.NotificationToCSE, &g.DCsi.CsiActive)
		got := parseSriRespGmsc(t, w)
		wantEqual(t, "GmscCamelSubscriptionInfo", gmscCamel(), got)
	})
}
