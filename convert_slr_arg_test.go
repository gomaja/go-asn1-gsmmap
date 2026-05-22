// convert_slr_arg_test.go
//
// Tests for the top-level SubscriberLocationReportArg converter and its
// Marshal()/ParseSubscriberLocationReport() entry points (opCode 86).
// Round-trip (struct→wire→struct), BER round-trip (Marshal→Parse), and
// targeted negative cases.
package gsmmap

import (
	"errors"
	"reflect"
	"testing"
)

// minimalSLRArg is the smallest valid arg: the three mandatory fields
// plus one subscriber identity (IMSI), satisfying the spec's
// "one of msisdn or imsi" rule.
func minimalSLRArg() *SubscriberLocationReportArg {
	return &SubscriberLocationReportArg{
		LcsEvent: LCSEventEmergencyCallOrigination,
		LcsClientID: LCSClientID{
			LcsClientType: LCSClientTypeEmergencyServices,
		},
		LcsLocationInfo: LCSLocationInfo{
			NetworkNodeNumber:       "31650000000",
			NetworkNodeNumberNature: 0x10,
			NetworkNodeNumberPlan:   0x01,
		},
		IMSI: "204080000000001",
	}
}

func TestSLRArgRoundTrip(t *testing.T) {
	age := int64(7)
	svcType := int64(42)
	seq := SequenceNumber(100)
	baro := UtranBaroPressureMeas(101325)
	acc := AccuracyFulfilmentRequestedAccuracyFulfilled

	cases := []struct {
		name string
		in   *SubscriberLocationReportArg
	}{
		{"minimal (IMSI identity)", minimalSLRArg()},
		{"minimal (MSISDN identity)", &SubscriberLocationReportArg{
			LcsEvent:        LCSEventMoLr,
			LcsClientID:     LCSClientID{LcsClientType: LCSClientTypeValueAddedServices},
			LcsLocationInfo: LCSLocationInfo{NetworkNodeNumber: "31650000000", NetworkNodeNumberNature: 0x10, NetworkNodeNumberPlan: 0x01},
			MSISDN:          "31612345678",
			MSISDNNature:    0x10,
			MSISDNPlan:      0x01,
		}},
		{"emergency routing (na-ESRD + na-ESRK)", &SubscriberLocationReportArg{
			LcsEvent:        LCSEventEmergencyCallOrigination,
			LcsClientID:     LCSClientID{LcsClientType: LCSClientTypeEmergencyServices},
			LcsLocationInfo: LCSLocationInfo{NetworkNodeNumber: "31650000000", NetworkNodeNumberNature: 0x10, NetworkNodeNumberPlan: 0x01},
			IMSI:            "204080000000001",
			NaESRD:          "8005551212",
			NaESRDNature:    0x10,
			NaESRDPlan:      0x01,
			NaESRK:          "8005556789",
			NaESRKNature:    0x10,
			NaESRKPlan:      0x01,
		}},
		{"with location estimate + age + IMEI", &SubscriberLocationReportArg{
			LcsEvent:              LCSEventDeferredmtLrResponse,
			LcsClientID:           LCSClientID{LcsClientType: LCSClientTypePlmnOperatorServices},
			LcsLocationInfo:       LCSLocationInfo{NetworkNodeNumber: "31650000000", NetworkNodeNumberNature: 0x10, NetworkNodeNumberPlan: 0x01},
			IMSI:                  "204080000000001",
			IMEI:                  "490154203237518",
			LocationEstimate:      HexBytes{0x04, 0x10, 0x20, 0x30, 0x40, 0x50, 0x60},
			AgeOfLocationEstimate: &age,
		}},
		{"CGI cell id", &SubscriberLocationReportArg{
			LcsEvent:        LCSEventMoLr,
			LcsClientID:     LCSClientID{LcsClientType: LCSClientTypeEmergencyServices},
			LcsLocationInfo: LCSLocationInfo{NetworkNodeNumber: "31650000000", NetworkNodeNumberNature: 0x10, NetworkNodeNumberPlan: 0x01},
			IMSI:            "204080000000001",
			CellGlobalId:    HexBytes{0x12, 0xf4, 0x10, 0x00, 0x01, 0x00, 0x02},
		}},
		{"LAI cell id", &SubscriberLocationReportArg{
			LcsEvent:        LCSEventMoLr,
			LcsClientID:     LCSClientID{LcsClientType: LCSClientTypeEmergencyServices},
			LcsLocationInfo: LCSLocationInfo{NetworkNodeNumber: "31650000000", NetworkNodeNumberNature: 0x10, NetworkNodeNumberPlan: 0x01},
			IMSI:            "204080000000001",
			LAI:             HexBytes{0x12, 0xf4, 0x10, 0x00, 0x01},
		}},
		{"NULL flags + h-gmlc + service type", &SubscriberLocationReportArg{
			LcsEvent:                  LCSEventMoLr,
			LcsClientID:               LCSClientID{LcsClientType: LCSClientTypeEmergencyServices},
			LcsLocationInfo:           LCSLocationInfo{NetworkNodeNumber: "31650000000", NetworkNodeNumberNature: 0x10, NetworkNodeNumberPlan: 0x01},
			IMSI:                      "204080000000001",
			HGmlcAddress:              "192.0.2.10",
			LcsServiceTypeID:          &svcType,
			SaiPresent:                true,
			PseudonymIndicator:        true,
			MoLrShortCircuitIndicator: true,
		}},
		{"positioning + velocity + sequence + baro + civic", &SubscriberLocationReportArg{
			LcsEvent:                       LCSEventMoLr,
			LcsClientID:                    LCSClientID{LcsClientType: LCSClientTypeEmergencyServices},
			LcsLocationInfo:                LCSLocationInfo{NetworkNodeNumber: "31650000000", NetworkNodeNumberNature: 0x10, NetworkNodeNumberPlan: 0x01},
			IMSI:                           "204080000000001",
			GeranPositioningData:           HexBytes{0x01, 0x02},
			UtranPositioningData:           HexBytes{0x01, 0x02, 0x03},
			AccuracyFulfilmentIndicator:    &acc,
			VelocityEstimate:               HexBytes{0x01, 0x02, 0x03, 0x04},
			SequenceNumber:                 &seq,
			GeranGANSSpositioningData:      HexBytes{0x05, 0x06},
			UtranGANSSpositioningData:      HexBytes{0x07},
			UtranAdditionalPositioningData: HexBytes{0x08},
			UtranBaroPressureMeas:          &baro,
			UtranCivicAddress:              HexBytes("civic-xml"),
		}},
		{"deferredmt-lrData + serving node handover", &SubscriberLocationReportArg{
			LcsEvent:        LCSEventDeferredmtLrResponse,
			LcsClientID:     LCSClientID{LcsClientType: LCSClientTypeEmergencyServices},
			LcsLocationInfo: LCSLocationInfo{NetworkNodeNumber: "31650000000", NetworkNodeNumberNature: 0x10, NetworkNodeNumberPlan: 0x01},
			IMSI:            "204080000000001",
			DeferredmtLrData: &DeferredmtLrData{
				DeferredLocationEventType: DeferredLocationEventType{MsAvailable: true},
			},
			TargetServingNodeForHandover: &ServingNodeAddress{
				MscNumber:       "31640000000",
				MscNumberNature: 0x10,
				MscNumberPlan:   0x01,
			},
			LcsReferenceNumber: HexBytes{0x2a},
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			wire, err := convertSubscriberLocationReportArgToWire(tc.in)
			if err != nil {
				t.Fatalf("toWire: %v", err)
			}
			got, err := convertWireToSubscriberLocationReportArg(wire)
			if err != nil {
				t.Fatalf("toStruct: %v", err)
			}
			if !reflect.DeepEqual(tc.in, got) {
				t.Errorf("round-trip mismatch:\n in=%+v\ngot=%+v", tc.in, got)
			}
		})
	}
}

// TestSLRArgBERRoundTrip exercises the public Marshal()/Parse() path
// through real BER encoding.
func TestSLRArgBERRoundTrip(t *testing.T) {
	in := minimalSLRArg()
	in.LocationEstimate = HexBytes{0x04, 0x10, 0x20, 0x30}
	in.MSISDN = "31612345678"
	in.MSISDNNature = 0x10
	in.MSISDNPlan = 0x01

	data, err := in.Marshal()
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Marshal produced empty output")
	}
	got, err := ParseSubscriberLocationReport(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if !reflect.DeepEqual(in, got) {
		t.Errorf("BER round-trip mismatch:\n in=%+v\ngot=%+v", in, got)
	}
}

func TestSLRArgEncodeNegative(t *testing.T) {
	cases := []struct {
		name string
		mut  func(a *SubscriberLocationReportArg)
		want error
	}{
		{"nil arg", nil, ErrSLRArgNil},
		{"LcsEvent out of range", func(a *SubscriberLocationReportArg) { a.LcsEvent = LCSEvent(99) }, ErrLCSEventInvalid},
		{"LcsClientType out of range", func(a *SubscriberLocationReportArg) { a.LcsClientID.LcsClientType = LCSClientType(99) }, ErrLCSClientTypeInvalid},
		{"LcsLocationInfo empty node", func(a *SubscriberLocationReportArg) { a.LcsLocationInfo.NetworkNodeNumber = "" }, ErrLCSLocationInfoNetworkNodeEmpty},
		{"IMSI too short", func(a *SubscriberLocationReportArg) { a.IMSI = "1234" }, ErrSLRArgIMSIInvalidSize},
		{"IMSI too long", func(a *SubscriberLocationReportArg) { a.IMSI = "1234567890123456" }, ErrSLRArgIMSIInvalidSize},
		{"IMEI wrong length", func(a *SubscriberLocationReportArg) { a.IMEI = "12345" }, ErrSLRArgIMEIInvalidSize},
		{"LocationEstimate too long", func(a *SubscriberLocationReportArg) { a.LocationEstimate = make(HexBytes, 21) }, ErrExtGeographicalInformationSize},
		{"AddLocationEstimate too long", func(a *SubscriberLocationReportArg) { a.AddLocationEstimate = make(HexBytes, 92) }, ErrAddGeographicalInformationSize},
		{"LcsReferenceNumber wrong size", func(a *SubscriberLocationReportArg) { a.LcsReferenceNumber = HexBytes{0x01, 0x02} }, ErrLCSReferenceNumberInvalidSize},
		{"GeranPositioningData too short", func(a *SubscriberLocationReportArg) { a.GeranPositioningData = HexBytes{0x01} }, ErrPositioningDataInformationSize},
		{"UtranPositioningData too short", func(a *SubscriberLocationReportArg) { a.UtranPositioningData = HexBytes{0x01, 0x02} }, ErrUtranPositioningDataInfoSize},
		{"CellGlobalId wrong size", func(a *SubscriberLocationReportArg) { a.CellGlobalId = HexBytes{0x01} }, ErrPSLResCellGlobalIdSize},
		{"CGI and LAI both set", func(a *SubscriberLocationReportArg) {
			a.CellGlobalId = make(HexBytes, 7)
			a.LAI = make(HexBytes, 5)
		}, ErrSLRArgCellGlobalIdAndLAIMutex},
		{"LcsServiceTypeID out of range", func(a *SubscriberLocationReportArg) { v := int64(128); a.LcsServiceTypeID = &v }, ErrSLRArgLcsServiceTypeIDOutOfRange},
		{"AccuracyFulfilmentIndicator out of range", func(a *SubscriberLocationReportArg) {
			v := AccuracyFulfilmentIndicator(9)
			a.AccuracyFulfilmentIndicator = &v
		}, ErrAccuracyFulfilmentIndicatorInvalid},
		{"VelocityEstimate too short", func(a *SubscriberLocationReportArg) { a.VelocityEstimate = HexBytes{0x01} }, ErrVelocityEstimateSize},
		{"SequenceNumber too low", func(a *SubscriberLocationReportArg) { v := SequenceNumber(0); a.SequenceNumber = &v }, ErrSequenceNumberOutOfRange},
		{"SequenceNumber too high", func(a *SubscriberLocationReportArg) { v := SequenceNumber(8640000); a.SequenceNumber = &v }, ErrSequenceNumberOutOfRange},
		{"UtranBaroPressureMeas out of range", func(a *SubscriberLocationReportArg) { v := UtranBaroPressureMeas(1); a.UtranBaroPressureMeas = &v }, ErrUtranBaroPressureMeasOutOfRange},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var in *SubscriberLocationReportArg
			if tc.mut != nil {
				in = minimalSLRArg()
				tc.mut(in)
			}
			_, err := convertSubscriberLocationReportArgToWire(in)
			if !errors.Is(err, tc.want) {
				t.Errorf("want errors.Is(_, %v), got %v", tc.want, err)
			}
		})
	}
}
