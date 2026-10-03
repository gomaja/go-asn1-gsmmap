package gsmmap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ----------------------------------------------------------------------------
// AMBR
// ----------------------------------------------------------------------------

func ptrInt64(v int64) *int64 { return &v }

func TestAMBR_RoundTrip(t *testing.T) {
	in := &AMBR{
		MaxRequestedBandwidthUL:         1_000_000,
		MaxRequestedBandwidthDL:         5_000_000,
		ExtendedMaxRequestedBandwidthUL: ptrInt64(20_000),
		ExtendedMaxRequestedBandwidthDL: ptrInt64(100_000),
	}
	w, err := convertAMBRToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToAMBR(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestAMBR_NegativeRejected(t *testing.T) {
	cases := []*AMBR{
		{MaxRequestedBandwidthUL: -1, MaxRequestedBandwidthDL: 0},
		{MaxRequestedBandwidthUL: 0, MaxRequestedBandwidthDL: -1},
		{MaxRequestedBandwidthUL: 1, MaxRequestedBandwidthDL: 1, ExtendedMaxRequestedBandwidthUL: ptrInt64(-1)},
		{MaxRequestedBandwidthUL: 1, MaxRequestedBandwidthDL: 1, ExtendedMaxRequestedBandwidthDL: ptrInt64(-1)},
	}
	for i, c := range cases {
		_, err := convertAMBRToWire(c)
		if !errors.Is(err, ErrAMBRBandwidthOutOfRange) {
			t.Fatalf("case %d: want ErrAMBRBandwidthOutOfRange, got %v", i, err)
		}
	}
}

func TestAMBR_NilPassthrough(t *testing.T) {
	w, err := convertAMBRToWire(nil)
	if err != nil || w != nil {
		t.Fatalf("toWire nil: w=%v err=%v", w, err)
	}
	o, err := convertWireToAMBR(nil)
	if err != nil || o != nil {
		t.Fatalf("fromWire nil: o=%v err=%v", o, err)
	}
}

// ----------------------------------------------------------------------------
// PDP-Context
// ----------------------------------------------------------------------------

func makePDPContext() PDPContext {
	siptoP := SIPTOAboveRanAllowed
	siptoLP := SIPTOAtLocalNetworkAllowed
	lipaP := LIPAConditional
	nidd := NIDDSCEFBasedDataDelivery
	return PDPContext{
		PdpContextId:               1,
		PdpType:                    HexBytes{0xf1, 0x21}, // 2 octets
		PdpAddress:                 HexBytes{0x0a, 0x01, 0x02, 0x03},
		QosSubscribed:              HexBytes{0x09, 0x01, 0x02},
		VplmnAddressAllowed:        true,
		Apn:                        HexBytes{'a', 'p', 'n', '.', 'e', 'x'},
		ExtQoSSubscribed:           HexBytes{0xff, 0xee},
		PdpChargingCharacteristics: HexBytes{0x08, 0x00},
		Ext2QoSSubscribed:          HexBytes{0xaa},
		Ext3QoSSubscribed:          HexBytes{0xbb},
		Ext4QoSSubscribed:          HexBytes{0xcc},
		ApnOiReplacement:           HexBytes("apn-oi-9b"), // 9 octets, exactly the lower bound
		ExtPdpType:                 HexBytes{0xf1, 0x57},
		ExtPdpAddress:              HexBytes{0x0a, 0x01},
		Ambr: &AMBR{
			MaxRequestedBandwidthUL: 1_000_000,
			MaxRequestedBandwidthDL: 5_000_000,
		},
		SiptoPermission:             &siptoP,
		LipaPermission:              &lipaP,
		RestorationPriority:         HexBytes{0x05},
		SiptoLocalNetworkPermission: &siptoLP,
		NIDDMechanism:               &nidd,
		SCEFID:                      HexBytes("scef.example.com"), // 16 octets, in 9..255
	}
}

func TestPDPContext_RoundTrip(t *testing.T) {
	in := makePDPContext()
	w, err := convertPDPContextToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToPDPContext(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestPDPContext_MinimalRoundTrip(t *testing.T) {
	in := PDPContext{
		PdpContextId:  1,
		PdpType:       HexBytes{0xf1, 0x21},
		QosSubscribed: HexBytes{0x09, 0x00, 0x00},
		Apn:           HexBytes{'a', 'p'},
	}
	w, err := convertPDPContextToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToPDPContext(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestPDPContext_ContextIdOutOfRange(t *testing.T) {
	for _, id := range []int{0, 51, 100} {
		in := makePDPContext()
		in.PdpContextId = id
		_, err := strictWire(convertPDPContextToWire(&in))
		if !matchesConstraint(err, "pdp-ContextId", "(1..50)") {
			t.Fatalf("id=%d: want BER constraint error, got %v", id, err)
		}
	}
}

func TestPDPContext_FieldSizeViolations(t *testing.T) {
	cases := []struct {
		name       string
		mut        func(*PDPContext)
		want       error
		path       string
		constraint string
	}{
		{"PdpType wrong size", func(p *PDPContext) { p.PdpType = HexBytes{0x01} }, nil, "pdp-Type", "SIZE (2)"},
		{"QosSubscribed empty", func(p *PDPContext) { p.QosSubscribed = HexBytes{} }, nil, "qos-Subscribed", "SIZE (3)"},
		{"QosSubscribed too short", func(p *PDPContext) { p.QosSubscribed = HexBytes{0x01, 0x02} }, nil, "qos-Subscribed", "SIZE (3)"},
		{"QosSubscribed too long", func(p *PDPContext) { p.QosSubscribed = HexBytes{0x01, 0x02, 0x03, 0x04} }, nil, "qos-Subscribed", "SIZE (3)"},
		{"ExtQoSSubscribed empty", func(p *PDPContext) { p.ExtQoSSubscribed = HexBytes{} }, nil, "ext-QoS-Subscribed", "SIZE (1..9)"},
		{"ExtQoSSubscribed too long", func(p *PDPContext) { p.ExtQoSSubscribed = make(HexBytes, 10) }, nil, "ext-QoS-Subscribed", "SIZE (1..9)"},
		{"Ext2QoSSubscribed empty", func(p *PDPContext) { p.Ext2QoSSubscribed = HexBytes{} }, nil, "ext2-QoS-Subscribed", "SIZE (1..3)"},
		{"Ext2QoSSubscribed too long", func(p *PDPContext) { p.Ext2QoSSubscribed = HexBytes{0x01, 0x02, 0x03, 0x04} }, nil, "ext2-QoS-Subscribed", "SIZE (1..3)"},
		{"Ext3QoSSubscribed empty", func(p *PDPContext) { p.Ext3QoSSubscribed = HexBytes{} }, nil, "ext3-QoS-Subscribed", "SIZE (1..2)"},
		{"Ext3QoSSubscribed too long", func(p *PDPContext) { p.Ext3QoSSubscribed = HexBytes{0x01, 0x02, 0x03} }, nil, "ext3-QoS-Subscribed", "SIZE (1..2)"},
		{"Ext4QoSSubscribed empty", func(p *PDPContext) { p.Ext4QoSSubscribed = HexBytes{} }, nil, "ext4-QoS-Subscribed", "SIZE (1)"},
		{"Ext4QoSSubscribed too long", func(p *PDPContext) { p.Ext4QoSSubscribed = HexBytes{0x01, 0x02} }, nil, "ext4-QoS-Subscribed", "SIZE (1)"},
		{"PdpAddress empty", func(p *PDPContext) { p.PdpAddress = HexBytes{} }, nil, "pdp-Address", "SIZE (1..16)"},
		{"PdpAddress too long", func(p *PDPContext) { p.PdpAddress = make(HexBytes, 17) }, nil, "pdp-Address", "SIZE (1..16)"},
		{"ExtPdpType wrong", func(p *PDPContext) { p.ExtPdpType = HexBytes{0x01} }, nil, "ext-pdp-Type", "SIZE (2)"},
		{"ExtPdpAddress too long", func(p *PDPContext) { p.ExtPdpAddress = make(HexBytes, 17) }, nil, "ext-pdp-Address", "SIZE (1..16)"},
		{"PdpChargingChars wrong", func(p *PDPContext) { p.PdpChargingCharacteristics = HexBytes{0x01} }, nil, "pdp-ChargingCharacteristics", "SIZE (2)"},
		{"ApnOiReplacement too short", func(p *PDPContext) { p.ApnOiReplacement = HexBytes("short") }, nil, "apn-oi-Replacement", "SIZE (9..100)"},
		{"RestorationPriority wrong", func(p *PDPContext) { p.RestorationPriority = HexBytes{0x01, 0x02} }, nil, "restoration-Priority", "SIZE (1)"},
		{"SCEFID too short", func(p *PDPContext) { p.SCEFID = HexBytes("short") }, nil, "sCEF-ID", "SIZE (9..255)"},
		{"Apn too short", func(p *PDPContext) { p.Apn = HexBytes{'a'} }, nil, "apn", "SIZE (2..63)"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := makePDPContext()
			tc.mut(&in)
			_, err := strictWire(convertPDPContextToWire(&in))
			if !matchesExpected(err, tc.want, tc.path, tc.constraint) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
}

func TestPDPContext_DecoderRejectsContextIdOutOfRange(t *testing.T) {
	// Strict decoding must report the context ID range violation as precisely
	// as strict encoding does.
	for _, id := range []int64{0, 51, 100} {
		w := &gsm_map.PDPContext{
			PdpContextId:  id,
			PdpType:       gsm_map.PDPType{0xf1, 0x21},
			QosSubscribed: gsm_map.QoSSubscribed{0x09, 0x00, 0x00},
			Apn:           gsm_map.APN{'a', 'p'},
		}
		wantConstraintError(t, strictDecodeWire(w), "pdp-ContextId", "(1..50)")
	}
}

func TestPDPContext_ExtPdpAddressRequiresPdpAddress(t *testing.T) {
	in := makePDPContext()
	in.PdpAddress = nil // remove pdp-Address
	// in.ExtPdpAddress is still populated by makePDPContext
	_, err := convertPDPContextToWire(&in)
	if !errors.Is(err, ErrExtPDPAddressWithoutPDPAddress) {
		t.Fatalf("want ErrExtPDPAddressWithoutPDPAddress, got %v", err)
	}
}

func TestPDPContext_ExtQoSHierarchy(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*PDPContext)
	}{
		{"Ext2 without Ext", func(p *PDPContext) { p.ExtQoSSubscribed = nil }},
		{"Ext3 without Ext2", func(p *PDPContext) { p.Ext2QoSSubscribed = nil }},
		{"Ext4 without Ext3", func(p *PDPContext) { p.Ext3QoSSubscribed = nil }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := makePDPContext()
			tc.mut(&in)
			_, err := convertPDPContextToWire(&in)
			if !errors.Is(err, ErrExtQoSHierarchyViolated) {
				t.Fatalf("want ErrExtQoSHierarchyViolated, got %v", err)
			}
		})
	}
}

func TestPDPContext_EnumOutOfRange(t *testing.T) {
	cases := []struct {
		name string
		mut  func(*PDPContext)
		want error
	}{
		{"SiptoPermission", func(p *PDPContext) {
			v := SIPTOPermission(99)
			p.SiptoPermission = &v
		}, ErrSIPTOPermissionInvalid},
		{"SiptoLocalNetworkPermission", func(p *PDPContext) {
			v := SIPTOLocalNetworkPermission(99)
			p.SiptoLocalNetworkPermission = &v
		}, ErrSIPTOLocalNetworkPermissionInvalid},
		{"LipaPermission", func(p *PDPContext) {
			v := LIPAPermission(99)
			p.LipaPermission = &v
		}, ErrLIPAPermissionInvalid},
		{"NIDDMechanism", func(p *PDPContext) {
			v := NIDDMechanism(99)
			p.NIDDMechanism = &v
		}, ErrNIDDMechanismInvalid},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := makePDPContext()
			tc.mut(&in)
			_, err := convertPDPContextToWire(&in)
			if !errors.Is(err, tc.want) {
				t.Fatalf("want %v, got %v", tc.want, err)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// GPRSDataList / GPRSSubscriptionData
// ----------------------------------------------------------------------------

func TestGPRSDataList_RoundTrip(t *testing.T) {
	in := GPRSDataList{makePDPContext(), makePDPContext()}
	in[1].PdpContextId = 2
	w, err := convertGPRSDataListToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToGPRSDataList(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch")
	}
}

func TestGPRSDataList_BoundsRejected(t *testing.T) {
	_, err := strictWire(convertGPRSDataListToWire(GPRSDataList{}))
	if !matchesConstraint(err, "gprsDataList", "SIZE (1..50)") {
		t.Fatalf("empty: want size error, got %v", err)
	}
	too := make(GPRSDataList, 51)
	for i := range too {
		p := makePDPContext()
		p.PdpContextId = (i % 50) + 1
		too[i] = p
	}
	_, err = strictWire(convertGPRSDataListToWire(too))
	if !matchesConstraint(err, "gprsDataList", "SIZE (1..50)") {
		t.Fatalf("over-max: want size error, got %v", err)
	}
}

func TestGPRSSubscriptionData_RoundTrip(t *testing.T) {
	in := &GPRSSubscriptionData{
		CompleteDataListIncluded: true,
		GprsDataList:             GPRSDataList{makePDPContext()},
		ApnOiReplacement:         HexBytes("apn-oi-9b"),
	}
	w, err := convertGPRSSubscriptionDataToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToGPRSSubscriptionData(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestGPRSSubscriptionData_MissingList(t *testing.T) {
	in := &GPRSSubscriptionData{} // GprsDataList nil
	_, err := convertGPRSSubscriptionDataToWire(in)
	if !errors.Is(err, ErrGPRSSubscriptionDataMissingList) {
		t.Fatalf("want ErrGPRSSubscriptionDataMissingList, got %v", err)
	}
	w := &gsm_map.GPRSSubscriptionData{} // wire side without list
	_, err = convertWireToGPRSSubscriptionData(w)
	if !errors.Is(err, ErrGPRSSubscriptionDataMissingList) {
		t.Fatalf("decode want ErrGPRSSubscriptionDataMissingList, got %v", err)
	}
}

func TestGPRSSubscriptionData_BadAPNOI(t *testing.T) {
	in := &GPRSSubscriptionData{
		GprsDataList:     GPRSDataList{makePDPContext()},
		ApnOiReplacement: HexBytes("short"),
	}
	_, err := strictWire(convertGPRSSubscriptionDataToWire(in))
	if !matchesConstraint(err, "apn-oi-Replacement", "SIZE (9..100)") {
		t.Fatalf("want BER constraint error, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// LSAData / LSADataList / LSAInformation
// ----------------------------------------------------------------------------

func makeLSAData() LSAData {
	return LSAData{
		LsaIdentity:            HexBytes{0x01, 0x02, 0x03},
		LsaAttributes:          HexBytes{0xff},
		LsaActiveModeIndicator: true,
	}
}

func TestLSAData_RoundTrip(t *testing.T) {
	in := makeLSAData()
	w, err := convertLSADataToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToLSAData(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestLSAData_FieldSize(t *testing.T) {
	cases := []struct {
		mut        func(*LSAData)
		path       string
		constraint string
	}{
		{func(l *LSAData) { l.LsaIdentity = HexBytes{0x01} }, "lsaIdentity", "SIZE (3)"},
		{func(l *LSAData) { l.LsaIdentity = HexBytes{0x01, 0x02, 0x03, 0x04} }, "lsaIdentity", "SIZE (3)"},
		{func(l *LSAData) { l.LsaAttributes = HexBytes{} }, "lsaAttributes", "SIZE (1)"},
		{func(l *LSAData) { l.LsaAttributes = HexBytes{0x01, 0x02} }, "lsaAttributes", "SIZE (1)"},
	}
	for _, tc := range cases {
		in := makeLSAData()
		tc.mut(&in)
		_, err := strictWire(convertLSADataToWire(&in))
		if !matchesConstraint(err, tc.path, tc.constraint) {
			t.Fatalf("want %s %s, got %v", tc.path, tc.constraint, err)
		}
	}
}

func TestLSADataList_RoundTrip(t *testing.T) {
	in := LSADataList{makeLSAData(), makeLSAData()}
	w, err := convertLSADataListToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToLSADataList(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch")
	}
}

func TestLSADataList_BoundsRejected(t *testing.T) {
	_, err := strictWire(convertLSADataListToWire(LSADataList{}))
	if !matchesConstraint(err, "lsaDataList", "SIZE (1..20)") {
		t.Fatalf("empty: want size error, got %v", err)
	}
	too := make(LSADataList, 21)
	for i := range too {
		too[i] = makeLSAData()
	}
	_, err = strictWire(convertLSADataListToWire(too))
	if !matchesConstraint(err, "lsaDataList", "SIZE (1..20)") {
		t.Fatalf("over-max: want size error, got %v", err)
	}
}

func TestLSAInformation_RoundTrip(t *testing.T) {
	indicator := LSAAccessOutsideRestricted
	in := &LSAInformation{
		CompleteDataListIncluded: true,
		LsaOnlyAccessIndicator:   &indicator,
		LsaDataList:              LSADataList{makeLSAData()},
	}
	w, err := convertLSAInformationToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToLSAInformation(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestLSAInformation_MinimalRoundTrip(t *testing.T) {
	// All optional fields omitted → minimal SEQUENCE
	in := &LSAInformation{}
	w, err := convertLSAInformationToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToLSAInformation(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestLSAInformation_OnlyAccessIndicatorOutOfRange(t *testing.T) {
	v := LSAOnlyAccessIndicator(99)
	in := &LSAInformation{LsaOnlyAccessIndicator: &v}
	_, err := convertLSAInformationToWire(in)
	if !errors.Is(err, ErrLSAOnlyAccessIndicatorInvalid) {
		t.Fatalf("want ErrLSAOnlyAccessIndicatorInvalid, got %v", err)
	}
}
