package gsmmap

import (
	"errors"
	"fmt"
	"math/bits"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/runtime"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func senderLCSCheck(t *testing.T, in marshaler, parse func([]byte) (marshaler, error), want error) {
	t.Helper()
	data, err := in.Marshal()
	if want != nil {
		if !errors.Is(err, want) {
			t.Fatalf("Marshal: %v, want %v", err, want)
		}
		return
	}
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if _, err := parse(data); err != nil {
		t.Fatalf("Parse: %v", err)
	}
}

func senderPSLArg() *ProvideSubscriberLocationArg {
	a := semPSLArg()
	a.IMSI = "001010123456789"
	return a
}

// TS 29.002 V19.1.0 §17.7.6 LocationInfoWithLMSI excludes overlapping node indicators.
func TestSenderSriSmNodeIndicators(t *testing.T) {
	for mask := 0; mask < 16; mask++ {
		t.Run(fmt.Sprint(mask), func(t *testing.T) {
			a := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "123", GprsNodeIndicator: mask&1 != 0, ImsNodeIndicator: mask&2 != 0, Smsf3gppAddressIndicator: mask&4 != 0, SmsfNon3gppAddressIndicator: mask&8 != 0}}
			var want error
			if bits.OnesCount(uint(mask)) > 1 {
				want = ErrSriSmRespNodeIndicatorsMutuallyExclusive
			}
			senderLCSCheck(t, a, asParser(ParseSriSmResp), want)
		})
	}
}

// TS 29.002 V19.1.0 §17.7.6 AdditionalNumber and ThirdNumber must use different CHOICE types.
func TestSenderSriSmAdditionalNumberTypes(t *testing.T) {
	nums := []*AdditionalNumber{nil, {MscNumber: "123"}, {SgsnNumber: "456"}}
	for i, a := range nums {
		for j, b := range nums {
			t.Run(fmt.Sprintf("%d/%d", i, j), func(t *testing.T) {
				in := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "789", AdditionalNumber: a, ThirdNumber: b}}
				var want error
				if i == j && i > 0 {
					want = ErrSriSmRespAdditionalNumbersSameType
				}
				// Different digits do not make equal number types legal.
				if b != nil {
					third := *b
					if third.MscNumber != "" {
						third.MscNumber = "789"
					} else {
						third.SgsnNumber = "789"
					}
					in.LocationInfoWithLMSI.ThirdNumber = &third
				}
				senderLCSCheck(t, in, asParser(ParseSriSmResp), want)
			})
		}
	}
}

// TS 29.002 V19.1.0 §17.7.6 ReportSM-DeliveryStatusArg excludes these nine field pairs.
func TestSenderReportSMOutcomeFields(t *testing.T) {
	pairs := []struct {
		name        string
		left, right func(*ReportSMDeliveryStatus)
	}{
		{"GPRS/additional outcome", func(a *ReportSMDeliveryStatus) { a.DeliveryOutcomeIndicator = true }, func(a *ReportSMDeliveryStatus) { a.AdditionalSMDeliveryOutcome = new(SmDeliverySuccessfulTransfer) }},
		{"GPRS/additional diagnostic", func(a *ReportSMDeliveryStatus) { a.DeliveryOutcomeIndicator = true }, func(a *ReportSMDeliveryStatus) { a.AdditionalAbsentSubscriberDiagnosticSM = new(1) }},
		{"IMS/GPRS", func(a *ReportSMDeliveryStatus) { a.IpSmGwIndicator = true }, func(a *ReportSMDeliveryStatus) { a.DeliveryOutcomeIndicator = true }},
		{"IMS/outcome", func(a *ReportSMDeliveryStatus) { a.IpSmGwIndicator = true }, func(a *ReportSMDeliveryStatus) { a.IpSmGwSmDeliveryOutcome = new(SmDeliverySuccessfulTransfer) }},
		{"IMS/diagnostic", func(a *ReportSMDeliveryStatus) { a.IpSmGwIndicator = true }, func(a *ReportSMDeliveryStatus) { a.IpSmGwAbsentSubscriberDiagnosticSM = new(1) }},
		{"SMSF3/outcome", func(a *ReportSMDeliveryStatus) { a.Smsf3gppDeliveryOutcomeIndicator = true }, func(a *ReportSMDeliveryStatus) { a.Smsf3gppDeliveryOutcome = new(SmDeliverySuccessfulTransfer) }},
		{"SMSF3/diagnostic", func(a *ReportSMDeliveryStatus) { a.Smsf3gppDeliveryOutcomeIndicator = true }, func(a *ReportSMDeliveryStatus) { a.Smsf3gppAbsentSubscriberDiagnosticSM = new(1) }},
		{"SMSFnon3/outcome", func(a *ReportSMDeliveryStatus) { a.SmsfNon3gppDeliveryOutcomeIndicator = true }, func(a *ReportSMDeliveryStatus) { a.SmsfNon3gppDeliveryOutcome = new(SmDeliverySuccessfulTransfer) }},
		{"SMSFnon3/diagnostic", func(a *ReportSMDeliveryStatus) { a.SmsfNon3gppDeliveryOutcomeIndicator = true }, func(a *ReportSMDeliveryStatus) { a.SmsfNon3gppAbsentSubscriberDiagnosticSM = new(1) }},
	}
	for _, p := range pairs {
		t.Run(p.name, func(t *testing.T) {
			for mask := 0; mask < 4; mask++ {
				a := &ReportSMDeliveryStatus{MSISDN: "123", ServiceCentreAddress: "456", SmDeliveryOutcome: SmDeliverySuccessfulTransfer}
				if mask&1 != 0 {
					p.left(a)
				}
				if mask&2 != 0 {
					p.right(a)
				}
				var want error
				if mask == 3 {
					want = ErrReportSMDeliveryStatusConflictingOutcomeFields
				}
				senderLCSCheck(t, a, asParser(ParseReportSMDeliveryStatus), want)
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.6 InformServiceCentreArg requires a primary diagnostic before its additional diagnostic.
func TestSenderISCAdditionalDiagnostic(t *testing.T) {
	for mask := 0; mask < 4; mask++ {
		a := &InformServiceCentre{Smsf3gppAbsentSubscriberDiagnosticSM: new(1), SmsfNon3gppAbsentSubscriberDiagnosticSM: new(1)}
		if mask&1 != 0 {
			a.AbsentSubscriberDiagnosticSM = new(1)
		}
		if mask&2 != 0 {
			a.AdditionalAbsentSubscriberDiagnosticSM = new(1)
		}
		var want error
		if mask == 2 {
			want = ErrIscAdditionalDiagnosticWithoutPrimary
		}
		senderLCSCheck(t, a, asParser(ParseInformServiceCentre), want)
	}
}

// TS 29.002 V19.1.0 §17.7.13 requires an IMSI or MSISDN in both LCS arguments.
func TestSenderLCSSubscriberIdentity(t *testing.T) {
	for mask := 0; mask < 4; mask++ {
		p := senderPSLArg()
		s := minimalSLRArg()
		p.IMSI = ""
		s.IMSI = ""
		if mask&1 != 0 {
			p.IMSI = "001010123456789"
			s.IMSI = p.IMSI
		}
		if mask&2 != 0 {
			p.MSISDN = "123"
			s.MSISDN = p.MSISDN
		}
		var want error
		if mask == 0 {
			want = ErrLCSSubscriberIdentityMissing
		}
		senderLCSCheck(t, p, asParser(ParseProvideSubscriberLocation), want)
		senderLCSCheck(t, s, asParser(ParseSubscriberLocationReport), want)
	}
}

// TS 29.002 V19.1.0 §17.7.13 requires a reference for activated/cancelled PSL and deferred SLR responses.
func TestSenderLCSDeferredReference(t *testing.T) {
	for event := LocationEstimateType(0); event <= 5; event++ {
		for _, present := range []bool{false, true} {
			p := senderPSLArg()
			p.LocationType.LocationEstimateType = event
			if present {
				p.LcsReferenceNumber = HexBytes{1}
			}
			var want error
			if !present && (event == LocationEstimateActivateDeferredLocation || event == LocationEstimateCancelDeferredLocation) {
				want = ErrLCSDeferredReferenceNumberMissing
			}
			senderLCSCheck(t, p, asParser(ParseProvideSubscriberLocation), want)
		}
	}
	for event := LCSEvent(0); event <= 5; event++ {
		for _, present := range []bool{false, true} {
			s := minimalSLRArg()
			s.LcsEvent = event
			if event == LCSEventDeferredmtLrResponse {
				s.DeferredmtLrData = &DeferredmtLrData{}
			}
			if present {
				s.LcsReferenceNumber = HexBytes{1}
			}
			var want error
			if !present && event == LCSEventDeferredmtLrResponse {
				want = ErrLCSDeferredReferenceNumberMissing
			}
			senderLCSCheck(t, s, asParser(ParseSubscriberLocationReport), want)
		}
	}
}

// TS 29.002 V19.1.0 §17.7.13 LCSClientName requires DCS 00001111; RequestorID and Codeword are independent.
func TestSenderLCSClientNameDCS(t *testing.T) {
	for v := 0; v < 256; v++ {
		name := &LCSClientName{DataCodingScheme: USSDDataCodingScheme(v), NameString: HexBytes{0x41}}
		p := senderPSLArg()
		p.LcsClientID = &LCSClientID{LcsClientName: name, LcsRequestorID: &LCSRequestorID{DataCodingScheme: 0x48, RequestorIDString: HexBytes{0x41}}}
		p.LcsCodeword = &LCSCodeword{DataCodingScheme: 0x48, LcsCodewordString: HexBytes{0x41}}
		s := minimalSLRArg()
		s.LcsClientID = *p.LcsClientID
		var want error
		if v != 0x0f {
			want = ErrLCSClientNameDataCodingSchemeInvalid
		}
		senderLCSCheck(t, p, asParser(ParseProvideSubscriberLocation), want)
		senderLCSCheck(t, s, asParser(ParseSubscriberLocationReport), want)
	}
}

// TS 29.002 V19.1.0 §17.7.13 excludes the GPRS indicator with dummy network node 0.
func TestSenderLCSDummyNode(t *testing.T) {
	for _, node := range []string{"0", "123"} {
		for _, flag := range []bool{false, true} {
			loc := LCSLocationInfo{NetworkNodeNumber: node, GprsNodeIndicator: flag}
			r := &SriLcsResp{TargetMS: SubscriberIdentity{IMSI: "001010123456789"}, LcsLocationInfo: loc}
			s := minimalSLRArg()
			s.LcsLocationInfo = loc
			d := minimalSLRArg()
			d.LcsEvent = LCSEventDeferredmtLrResponse
			d.LcsReferenceNumber = HexBytes{1}
			d.DeferredmtLrData = &DeferredmtLrData{TerminationCause: new(TerminationMtLrRestart), LcsLocationInfo: &loc}
			var want error
			if node == "0" && flag {
				want = ErrLCSLocationInfoGPRSIndicatorWithDummyNode
			}
			senderLCSCheck(t, r, asParser(ParseSriLcsResp), want)
			senderLCSCheck(t, s, asParser(ParseSubscriberLocationReport), want)
			senderLCSCheck(t, d, asParser(ParseSubscriberLocationReport), want)
		}
	}
}

// TS 29.002 V19.1.0 §17.7.13 makes DeferredmtLrData present iff the event is deferredmt-lrResponse.
func TestSenderSLRDeferredDataPresence(t *testing.T) {
	for event := LCSEvent(0); event <= 5; event++ {
		for _, present := range []bool{false, true} {
			s := minimalSLRArg()
			s.LcsEvent = event
			s.LcsReferenceNumber = HexBytes{1}
			if present {
				s.DeferredmtLrData = &DeferredmtLrData{}
			}
			var want error
			if (event == LCSEventDeferredmtLrResponse) != present {
				want = ErrSLRArgDeferredmtLrDataPresence
			}
			senderLCSCheck(t, s, asParser(ParseSubscriberLocationReport), want)
		}
	}
}

// TS 29.002 V19.1.0 §17.7.13 DeferredmtLrData permits nested location only with mt-lrRestart.
func TestSenderSLRDeferredLocationRequiresRestart(t *testing.T) {
	for cause := -1; cause <= 9; cause++ {
		for _, present := range []bool{false, true} {
			s := minimalSLRArg()
			s.LcsEvent = LCSEventDeferredmtLrResponse
			s.LcsReferenceNumber = HexBytes{1}
			s.DeferredmtLrData = &DeferredmtLrData{}
			if cause >= 0 {
				s.DeferredmtLrData.TerminationCause = new(TerminationCause(cause))
			}
			if present {
				s.DeferredmtLrData.LcsLocationInfo = &LCSLocationInfo{NetworkNodeNumber: "123"}
			}
			var want error
			if present && cause != 4 {
				want = ErrDeferredmtLrDataLocationInfoRequiresRestart
			}
			senderLCSCheck(t, s, asParser(ParseSubscriberLocationReport), want)
		}
	}
}

// TS 29.002 V19.1.0 §17.7.13 SubscriberLocationReport-Res excludes both routing identifiers.
func TestSenderSLRResponseRoutingIdentifiers(t *testing.T) {
	for mask := 0; mask < 4; mask++ {
		a := &SubscriberLocationReportRes{}
		if mask&1 != 0 {
			a.NaESRK = "123"
		}
		if mask&2 != 0 {
			a.NaESRD = "456"
		}
		var want error
		if mask == 3 {
			want = ErrSLRResRoutingIdentifiersMutuallyExclusive
		}
		senderLCSCheck(t, a, asParser(ParseSubscriberLocationReportRes), want)
	}
}

func senderLCSGeography(t *testing.T, value HexBytes, add bool, want error) {
	t.Helper()
	r := &ProvideSubscriberLocationRes{LocationEstimate: HexBytes{0}}
	s := minimalSLRArg()
	if add {
		r.AddLocationEstimate = value
		s.AddLocationEstimate = value
	} else {
		r.LocationEstimate = value
		s.LocationEstimate = value
	}
	senderLCSCheck(t, r, asParser(ParseProvideSubscriberLocationRes), want)
	senderLCSCheck(t, s, asParser(ParseSubscriberLocationReport), want)
}

// TS 29.002 V19.1.0 §17.7.13; TS 23.032 V19.0.0 §7.3 figures define exact shape lengths.
func TestSenderLCSGeographicalShapes(t *testing.T) {
	lengths := map[byte]int{0: 7, 1: 8, 3: 11, 8: 9, 9: 14, 10: 13, 11: 13, 12: 18, 13: 13, 14: 18}
	for shape := byte(0); shape < 16; shape++ {
		n, ok := lengths[shape]
		if !ok {
			n = 7
		}
		d := make(HexBytes, n)
		d[0] = shape << 4
		extOK := shape == 0 || shape == 1 || shape == 3 || shape == 9 || shape == 10
		var extWant, addWant error
		if !extOK {
			extWant = ErrExtGeographicalInformationInvalid
		}
		if !ok {
			addWant = ErrAddGeographicalInformationInvalid
		}
		senderLCSGeography(t, d, false, extWant)
		senderLCSGeography(t, d, true, addWant)
		if ok {
			for _, delta := range []int{-1, 1} {
				bad := make(HexBytes, n+delta)
				bad[0] = shape << 4
				senderLCSGeography(t, bad, true, ErrAddGeographicalInformationInvalid)
				senderLCSGeography(t, bad, false, ErrExtGeographicalInformationInvalid)
			}
		}
	}
	for _, count := range []byte{2, 3, 15} {
		d := make(HexBytes, 1+6*int(count))
		d[0] = 0x50 | count
		var want error
		if count == 2 {
			want = ErrAddGeographicalInformationInvalid
		}
		senderLCSGeography(t, d, true, want)
	}
	// §7.3.5: altitude direction is a data bit, not a spare.
	d := make(HexBytes, 9)
	d[0] = 0x80
	d[7] = 0xff
	senderLCSGeography(t, d, true, nil)
	// §17.7.13: valid Ext does not prohibit sending valid Add; Parse discards Add.
	in := &ProvideSubscriberLocationRes{LocationEstimate: make(HexBytes, 7), AddLocationEstimate: make(HexBytes, 7)}
	senderLCSCheck(t, in, asParser(ParseProvideSubscriberLocationRes), nil)
	// §17.7.13: a deferred PSL response's location estimate is ignored.
	in = &ProvideSubscriberLocationRes{LocationEstimate: HexBytes{0xf0, 1}, DeferredmtLrResponseIndicator: true}
	senderLCSCheck(t, in, asParser(ParseProvideSubscriberLocationRes), nil)
}

// TS 23.032 V19.0.0 §7.3 figures 6,6a,9,10 and high accuracy figures explicitly zero these bits.
func TestSenderLCSGeographicalCoding(t *testing.T) {
	coding := []struct {
		shape         byte
		n             int
		zeros, angles []int
	}{
		{1, 8, []int{7}, nil}, {3, 11, []int{7, 8, 10}, []int{9}}, {9, 14, []int{9, 10, 12, 13}, []int{11}},
		{10, 13, []int{9, 12}, []int{10, 11}}, {11, 13, []int{12}, []int{11}},
		{12, 18, []int{9, 15, 17}, []int{14}}, {13, 13, nil, []int{11}}, {14, 18, []int{9}, []int{14}},
	}
	for _, c := range coding {
		for _, i := range c.zeros {
			for _, mask := range []byte{0x80, 0x40} {
				if i != 9 || (c.shape != 12 && c.shape != 14) {
					if mask == 0x40 {
						continue
					}
				}
				d := make(HexBytes, c.n)
				d[0] = c.shape << 4
				d[i] = mask
				senderLCSGeography(t, d, true, ErrAddGeographicalInformationInvalid)
				if c.shape < 11 {
					senderLCSGeography(t, d, false, ErrExtGeographicalInformationInvalid)
				}
			}
		}
		for _, i := range c.angles {
			for _, v := range []byte{0, 179, 180, 255} {
				d := make(HexBytes, c.n)
				d[0] = c.shape << 4
				d[i] = v
				var want error
				if v > 179 {
					want = ErrAddGeographicalInformationInvalid
				}
				senderLCSGeography(t, d, true, want)
				if c.shape < 11 {
					var wantExt error
					if v > 179 {
						wantExt = ErrExtGeographicalInformationInvalid
					}
					senderLCSGeography(t, d, false, wantExt)
				}
			}
		}
	}
	// §§7.3.3b,7.3.6b: U, HU and VU are independent uncertainty selectors.
	for flags := byte(0); flags < 4; flags++ {
		d := make(HexBytes, 18)
		d[0] = 0xe0
		d[15] = (flags & 1) << 7
		d[17] = (flags & 2) << 6
		senderLCSGeography(t, d, true, nil)
	}
	for _, shape := range []byte{11, 13} {
		for uncertainty := 0; uncertainty < 256; uncertainty++ {
			d := make(HexBytes, 13)
			d[0] = shape << 4
			d[9] = byte(uncertainty)
			d[10] = byte(uncertainty)
			d[12] = 127
			if shape == 13 {
				d[12] |= 0x80
			}
			senderLCSGeography(t, d, true, nil)
		}
	}
	// §6.3a: high accuracy altitude is signed 22-bit with these inclusive limits.
	for _, shape := range []byte{12, 14} {
		for _, alt := range []int32{-64001, -64000, 1280000, 1280001} {
			d := make(HexBytes, 18)
			d[0] = shape << 4
			n := uint32(alt) & 0x3fffff
			d[9] = byte(n >> 16)
			d[10] = byte(n >> 8)
			d[11] = byte(n)
			var want error
			if alt < -64000 || alt > 1280000 {
				want = ErrAddGeographicalInformationInvalid
			}
			senderLCSGeography(t, d, true, want)
		}
	}
	// §6.5 confidence 101..127 is not a mandatory prohibition; header spare nibble has no explicit zero.
	d := make(HexBytes, 11)
	d[0] = 0x3f
	d[10] = 127
	senderLCSGeography(t, d, false, nil)
	senderLCSGeography(t, d, true, nil)
}

// TS 29.002 V19.1.0 §17.7.13; TS 23.032 V19.0.0 §8.6 velocity type is the high nibble.
func TestSenderLCSVelocityTypes(t *testing.T) {
	// TS 29.002 V19.1.0 §17.7.13: types 0 to 3 take 4, 5, 5 and 7 octets.
	lengths := map[byte]int{0: 4, 1: 5, 2: 5, 3: 7}
	for typ := byte(0); typ < 16; typ++ {
		for n := 4; n <= 7; n++ {
			v := make(HexBytes, n)
			v[0] = typ<<4 | 3
			r := &ProvideSubscriberLocationRes{LocationEstimate: make(HexBytes, 7), VelocityEstimate: v}
			s := minimalSLRArg()
			s.VelocityEstimate = v
			var want error
			switch {
			case typ > 3:
				want = ErrVelocityEstimateTypeInvalid
			case n != lengths[typ]:
				want = ErrVelocityEstimateLengthInvalid
			}
			senderLCSCheck(t, r, asParser(ParseProvideSubscriberLocationRes), want)
			senderLCSCheck(t, s, asParser(ParseSubscriberLocationReport), want)
		}
	}
}

// TS 29.002 V19.1.0 §17.7.13 Parse preserves an LCSClientName's received DCS.
func TestSenderLCSNameParsePreservesDCS(t *testing.T) {
	a := senderPSLArg()
	a.LcsClientID = &LCSClientID{LcsClientName: &LCSClientName{DataCodingScheme: 0x0f, NameString: HexBytes{0x41}}}
	w := semMust(convertProvideSubscriberLocationArgToWire(a))
	w.LcsClientID.LcsClientName.DataCodingScheme[0] = 0x48
	data := semMust(w.MarshalBER())
	got := semMust(ParseProvideSubscriberLocation(data))
	if !reflect.DeepEqual(got.LcsClientID.LcsClientName.NameString, HexBytes{0x41}) || got.LcsClientID.LcsClientName.DataCodingScheme != 0x48 {
		t.Fatalf("name not preserved: %+v", got.LcsClientID.LcsClientName)
	}
	senderLCSCheck(t, got, asParser(ParseProvideSubscriberLocation), ErrLCSClientNameDataCodingSchemeInvalid)
}

// TS 29.002 V19.1.0 §§17.7.6,17.7.13: these sender violations remain received values.
func senderLCSSemLenientCases() []semLenientCase {
	return []semLenientCase{
		{ErrVelocityEstimateLengthInvalid, asParser(ParseProvideSubscriberLocationRes), func(*testing.T) semWire {
			v := gsm_map.VelocityEstimate{0x03, 0, 0, 0, 0, 0, 0} // type 0 in seven octets
			return &gsm_map.ProvideSubscriberLocationRes{LocationEstimate: make(gsm_map.ExtGeographicalInformation, 7), VelocityEstimate: &v}
		}},
		{ErrSriSmRespNodeIndicatorsMutuallyExclusive, asParser(ParseSriSmResp), func(*testing.T) semWire {
			w := semMust(convertSriSmRespToRes(&SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "123"}}))
			w.LocationInfoWithLMSI.GprsNodeIndicator = &struct{}{}
			w.LocationInfoWithLMSI.ImsNodeIndicator = &struct{}{}
			return w
		}},
		{ErrSriSmRespAdditionalNumbersSameType, asParser(ParseSriSmResp), func(*testing.T) semWire {
			w := semMust(convertSriSmRespToRes(&SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "123", AdditionalNumber: &AdditionalNumber{MscNumber: "123"}}}))
			n := gsm_map.NewAdditionalNumberMscNumber([]byte{0x80, 0x54, 0xf6})
			w.LocationInfoWithLMSI.ThirdNumber = &n
			return w
		}},
		{ErrReportSMDeliveryStatusConflictingOutcomeFields, asParser(ParseReportSMDeliveryStatus), func(*testing.T) semWire {
			w := semMust(convertReportSMDeliveryStatusToArg(&ReportSMDeliveryStatus{MSISDN: "123", ServiceCentreAddress: "456", SmDeliveryOutcome: SmDeliverySuccessfulTransfer, AdditionalSMDeliveryOutcome: new(SmDeliverySuccessfulTransfer)}))
			w.DeliveryOutcomeIndicator = &struct{}{}
			return w
		}},
		{ErrIscAdditionalDiagnosticWithoutPrimary, asParser(ParseInformServiceCentre), func(*testing.T) semWire {
			return &gsm_map.InformServiceCentreArg{AdditionalAbsentSubscriberDiagnosticSM: new(gsm_map.AbsentSubscriberDiagnosticSM(1))}
		}},
		{ErrLCSLocationInfoGPRSIndicatorWithDummyNode, asParser(ParseSriLcsResp), func(*testing.T) semWire {
			w := semMust(convertSriLcsRespToRes(&SriLcsResp{TargetMS: SubscriberIdentity{IMSI: "001010123456789"}, LcsLocationInfo: LCSLocationInfo{NetworkNodeNumber: "0"}}))
			w.LcsLocationInfo.GprsNodeIndicator = &struct{}{}
			return w
		}},
		{ErrLCSLocationInfoGPRSIndicatorWithDummyNode, asParser(ParseSubscriberLocationReport), func(*testing.T) semWire {
			a := minimalSLRArg()
			a.LcsLocationInfo.NetworkNodeNumber = "0"
			w := semMust(convertSubscriberLocationReportArgToWire(a))
			w.LcsLocationInfo.GprsNodeIndicator = &struct{}{}
			return w
		}},
		{ErrLCSLocationInfoGPRSIndicatorWithDummyNode, asParser(ParseSubscriberLocationReport), func(*testing.T) semWire {
			a := minimalSLRArg()
			a.LcsEvent = LCSEventDeferredmtLrResponse
			a.LcsReferenceNumber = HexBytes{1}
			a.DeferredmtLrData = &DeferredmtLrData{TerminationCause: new(TerminationMtLrRestart), LcsLocationInfo: &LCSLocationInfo{NetworkNodeNumber: "0"}}
			w := semMust(convertSubscriberLocationReportArgToWire(a))
			w.DeferredmtLrData.LcsLocationInfo.GprsNodeIndicator = &struct{}{}
			return w
		}},
		{ErrLCSSubscriberIdentityMissing, asParser(ParseProvideSubscriberLocation), func(*testing.T) semWire {
			w := semMust(convertProvideSubscriberLocationArgToWire(senderPSLArg()))
			w.Imsi = nil
			w.Msisdn = nil
			return w
		}},
		{ErrLCSSubscriberIdentityMissing, asParser(ParseSubscriberLocationReport), func(*testing.T) semWire {
			w := semMust(convertSubscriberLocationReportArgToWire(minimalSLRArg()))
			w.Imsi, w.Msisdn = nil, nil
			return w
		}},
		{ErrLCSDeferredReferenceNumberMissing, asParser(ParseProvideSubscriberLocation), func(*testing.T) semWire {
			w := semMust(convertProvideSubscriberLocationArgToWire(senderPSLArg()))
			w.LocationType.LocationEstimateType = LocationEstimateActivateDeferredLocation
			return w
		}},
		{ErrLCSDeferredReferenceNumberMissing, asParser(ParseSubscriberLocationReport), func(*testing.T) semWire {
			a := minimalSLRArg()
			a.LcsEvent = LCSEventDeferredmtLrResponse
			a.LcsReferenceNumber = HexBytes{1}
			a.DeferredmtLrData = &DeferredmtLrData{}
			w := semMust(convertSubscriberLocationReportArgToWire(a))
			w.LcsReferenceNumber = nil
			return w
		}},
		{ErrLCSClientNameDataCodingSchemeInvalid, asParser(ParseProvideSubscriberLocation), func(*testing.T) semWire {
			a := senderPSLArg()
			a.LcsClientID = &LCSClientID{LcsClientName: &LCSClientName{DataCodingScheme: 0x0f, NameString: HexBytes{0x41}}}
			w := semMust(convertProvideSubscriberLocationArgToWire(a))
			w.LcsClientID.LcsClientName.DataCodingScheme[0] = 0x48
			return w
		}},
		{ErrLCSClientNameDataCodingSchemeInvalid, asParser(ParseSubscriberLocationReport), func(*testing.T) semWire {
			a := minimalSLRArg()
			a.LcsClientID.LcsClientName = &LCSClientName{DataCodingScheme: 0x0f, NameString: HexBytes{0x41}}
			w := semMust(convertSubscriberLocationReportArgToWire(a))
			w.LcsClientID.LcsClientName.DataCodingScheme[0] = 0x48
			return w
		}},
		{ErrSLRArgDeferredmtLrDataPresence, asParser(ParseSubscriberLocationReport), func(*testing.T) semWire {
			w := semMust(convertSubscriberLocationReportArgToWire(minimalSLRArg()))
			w.DeferredmtLrData = &gsm_map.DeferredmtLrData{DeferredLocationEventType: runtime.BitString{Bytes: []byte{0}, BitLength: 5}}
			return w
		}},
		{ErrSLRArgDeferredmtLrDataPresence, asParser(ParseSubscriberLocationReport), func(*testing.T) semWire {
			a := minimalSLRArg()
			a.LcsEvent = LCSEventDeferredmtLrResponse
			a.LcsReferenceNumber = HexBytes{1}
			a.DeferredmtLrData = &DeferredmtLrData{}
			w := semMust(convertSubscriberLocationReportArgToWire(a))
			w.DeferredmtLrData = nil
			return w
		}},
		{ErrDeferredmtLrDataLocationInfoRequiresRestart, asParser(ParseSubscriberLocationReport), func(*testing.T) semWire {
			a := minimalSLRArg()
			a.LcsEvent = LCSEventDeferredmtLrResponse
			a.LcsReferenceNumber = HexBytes{1}
			a.DeferredmtLrData = &DeferredmtLrData{}
			w := semMust(convertSubscriberLocationReportArgToWire(a))
			w.DeferredmtLrData.LcsLocationInfo = &w.LcsLocationInfo
			return w
		}},
	}
}

// TS 29.002 V19.1.0 §17.7.6 permits each outcome or node indicator alone.
func TestSenderReportSMIndicatorsRoundTrip(t *testing.T) {
	for _, field := range []string{"IpSmGwIndicator", "Smsf3gppDeliveryOutcomeIndicator", "SmsfNon3gppDeliveryOutcomeIndicator", "DeliveryOutcomeIndicator", "GprsSupportIndicator"} {
		t.Run(field, func(t *testing.T) {
			in := minimalReportSMDeliveryStatus()
			reflect.ValueOf(in).Elem().FieldByName(field).SetBool(true)
			senderMobilityRoundTrip(t, in, asParser(ParseReportSMDeliveryStatus))
		})
	}
}

func TestSenderSriSmIndicatorsRoundTrip(t *testing.T) {
	for _, field := range []string{"GprsNodeIndicator", "ImsNodeIndicator", "Smsf3gppAddressIndicator", "SmsfNon3gppAddressIndicator"} {
		t.Run(field, func(t *testing.T) {
			in := &SriSmResp{IMSI: "001010123456789", LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: "123"}}
			reflect.ValueOf(&in.LocationInfoWithLMSI).Elem().FieldByName(field).SetBool(true)
			senderMobilityRoundTrip(t, in, asParser(ParseSriSmResp))
		})
	}
}
