package gsmmap

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ----------------------------------------------------------------------------
// MC-SS-Info
// ----------------------------------------------------------------------------

func TestMCSSInfo_RoundTrip(t *testing.T) {
	in := &MCSSInfo{
		SsCode:   SsCode(0x21),
		SsStatus: HexBytes{0x01, 0x02},
		NbrSB:    7,
		NbrUser:  3,
	}
	w, err := convertMCSSInfoToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToMCSSInfo(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestMCSSInfo_NbrSBOutOfRange(t *testing.T) {
	for _, v := range []int{0, 1, 8, 100} {
		in := &MCSSInfo{SsStatus: HexBytes{0x01}, NbrSB: v, NbrUser: 1}
		_, err := strictWire(convertMCSSInfoToWire(in))
		if !matchesConstraint(err, "nbrSB", "(2..7)") {
			t.Fatalf("NbrSB=%d: want BER constraint error, got %v", v, err)
		}
	}
}

func TestMCSSInfo_NbrUserOutOfRange(t *testing.T) {
	for _, v := range []int{0, 8, 100} {
		in := &MCSSInfo{SsStatus: HexBytes{0x01}, NbrSB: 2, NbrUser: v}
		_, err := strictWire(convertMCSSInfoToWire(in))
		if !matchesConstraint(err, "nbrUser", "(1..7)") {
			t.Fatalf("NbrUser=%d: want BER constraint error, got %v", v, err)
		}
	}
}

func TestMCSSInfo_SsStatusInvalid(t *testing.T) {
	in := &MCSSInfo{SsStatus: HexBytes{}, NbrSB: 2, NbrUser: 1}
	_, err := strictWire(convertMCSSInfoToWire(in))
	if !matchesConstraint(err, "ss-Status", "SIZE (1..5)") {
		t.Fatalf("want BER constraint error, got %v", err)
	}
}

func TestMCSSInfo_NilPassthrough(t *testing.T) {
	w, err := convertMCSSInfoToWire(nil)
	if err != nil || w != nil {
		t.Fatalf("toWire nil: w=%v err=%v", w, err)
	}
	o, err := convertWireToMCSSInfo(nil)
	if err != nil || o != nil {
		t.Fatalf("fromWire nil: o=%v err=%v", o, err)
	}
}

func TestMCSSInfo_DecodeRejectsMissingSsCode(t *testing.T) {
	w := &gsm_map.MCSSInfo{
		SsCode:   gsm_map.SSCode{},
		SsStatus: gsm_map.ExtSSStatus{0x01},
		NbrSB:    2,
		NbrUser:  1,
	}
	wantConstraintError(t, strictDecodeWire(w), "ss-Code", "SIZE (1)")
}

// ----------------------------------------------------------------------------
// CSG-SubscriptionData / list / VPLMN-list
// ----------------------------------------------------------------------------

func makeCSGEntry() CSGSubscriptionData {
	return CSGSubscriptionData{
		CsgId:          HexBytes{0x12, 0x34, 0x56, 0x60}, // 27-bit BIT STRING (4 octets)
		CsgIdBitLength: 27,
		ExpirationDate: HexBytes{0x17, 0x0a, 0x01, 0x00}, // Time is SIZE (4), TS 29.002 V19.1.0 §17.7.8
		LipaAllowedAPNList: []HexBytes{
			{'a', 'p', 'n'}, // 3 octets, in 2..63 range
		},
		PlmnId: HexBytes{0x62, 0xf2, 0x10},
	}
}

func TestCSGSubscriptionData_RoundTrip(t *testing.T) {
	in := makeCSGEntry()
	w, err := strictWire(convertCSGSubscriptionDataToWire(&in))
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToCSGSubscriptionData(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestCSGSubscriptionData_BitLengthZeroRejected(t *testing.T) {
	// Fix 5: silent coercion of CsgIdBitLength=0 → 27 has been removed
	// because it caused encode/decode round-trip mutation.
	in := makeCSGEntry()
	in.CsgIdBitLength = 0
	_, err := convertCSGSubscriptionDataToWire(&in)
	if !errors.Is(err, ErrCSGIdInvalidSize) {
		t.Fatalf("want ErrCSGIdInvalidSize for BitLength=0, got %v", err)
	}
}

func TestCSGSubscriptionData_BadCsgId(t *testing.T) {
	cases := []struct {
		name string
		in   CSGSubscriptionData
	}{
		{"wrong octets", CSGSubscriptionData{CsgId: HexBytes{0x01, 0x02}, CsgIdBitLength: 27}},
		{"wrong bits", CSGSubscriptionData{CsgId: HexBytes{0x01, 0x02, 0x03, 0x04}, CsgIdBitLength: 32}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := convertCSGSubscriptionDataToWire(&tc.in)
			if !errors.Is(err, ErrCSGIdInvalidSize) {
				t.Fatalf("want ErrCSGIdInvalidSize, got %v", err)
			}
		})
	}
}

func TestCSGSubscriptionData_BadAPN(t *testing.T) {
	in := makeCSGEntry()
	in.LipaAllowedAPNList = []HexBytes{{'a'}} // 1 octet, below the 2..63 range
	_, err := convertCSGSubscriptionDataToWire(&in)
	if !errors.Is(err, ErrAPNInvalidSize) {
		t.Fatalf("want ErrAPNInvalidSize, got %v", err)
	}
}

func TestCSGSubscriptionData_EmptyAPNList(t *testing.T) {
	in := makeCSGEntry()
	in.LipaAllowedAPNList = []HexBytes{} // present but empty → not allowed (use nil)
	_, err := strictWire(convertCSGSubscriptionDataToWire(&in))
	if !matchesConstraint(err, "lipa-AllowedAPNList", "SIZE (1..50)") {
		t.Fatalf("want BER constraint error, got %v", err)
	}
}

func TestCSGSubscriptionData_OverMaxAPNList(t *testing.T) {
	in := makeCSGEntry()
	apns := make([]HexBytes, 51)
	for i := range apns {
		apns[i] = HexBytes{'a', 'p', 'n'}
	}
	in.LipaAllowedAPNList = apns
	_, err := strictWire(convertCSGSubscriptionDataToWire(&in))
	if !matchesConstraint(err, "lipa-AllowedAPNList", "SIZE (1..50)") {
		t.Fatalf("want BER constraint error, got %v", err)
	}
}

func TestCSGSubscriptionData_BadPlmnId(t *testing.T) {
	in := makeCSGEntry()
	in.PlmnId = HexBytes{0x01, 0x02} // too short
	_, err := strictWire(convertCSGSubscriptionDataToWire(&in))
	if !matchesConstraint(err, "plmn-Id", "SIZE (3)") {
		t.Fatalf("want BER constraint error, got %v", err)
	}
}

func TestCSGSubscriptionDataList_RoundTrip(t *testing.T) {
	in := CSGSubscriptionDataList{makeCSGEntry(), makeCSGEntry()}
	w, err := strictWire(convertCSGSubscriptionDataListToWire(in))
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToCSGSubscriptionDataList(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch")
	}
}

func TestCSGSubscriptionDataList_BoundsRejected(t *testing.T) {
	empty := CSGSubscriptionDataList{}
	_, err := strictWire(convertCSGSubscriptionDataListToWire(empty))
	if !matchesConstraint(err, "csg-SubscriptionDataList", "SIZE (1..50)") {
		t.Fatalf("empty: want size error, got %v", err)
	}
	too := make(CSGSubscriptionDataList, 50+1)
	for i := range too {
		too[i] = makeCSGEntry()
	}
	_, err = strictWire(convertCSGSubscriptionDataListToWire(too))
	if !matchesConstraint(err, "csg-SubscriptionDataList", "SIZE (1..50)") {
		t.Fatalf("over-max: want size error, got %v", err)
	}
}

func TestVPLMNCSGSubscriptionDataList_RoundTrip(t *testing.T) {
	in := VPLMNCSGSubscriptionDataList{makeCSGEntry()}
	w, err := strictWire(convertVPLMNCSGSubscriptionDataListToWire(in))
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToVPLMNCSGSubscriptionDataList(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch")
	}
}

func TestVPLMNCSGSubscriptionDataList_BoundsRejected(t *testing.T) {
	_, err := strictWire(convertVPLMNCSGSubscriptionDataListToWire(VPLMNCSGSubscriptionDataList{}))
	if !matchesConstraint(err, "vplmn-Csg-SubscriptionDataList", "SIZE (1..50)") {
		t.Fatalf("empty: want size error, got %v", err)
	}
	too := make(VPLMNCSGSubscriptionDataList, 50+1)
	for i := range too {
		too[i] = makeCSGEntry()
	}
	_, err = strictWire(convertVPLMNCSGSubscriptionDataListToWire(too))
	if !matchesConstraint(err, "vplmn-Csg-SubscriptionDataList", "SIZE (1..50)") {
		t.Fatalf("over-max: want size error, got %v", err)
	}
}

// ----------------------------------------------------------------------------
// AdjacentAccessRestrictionData / list
// ----------------------------------------------------------------------------

func makeAdjacentEntry() AdjacentAccessRestrictionData {
	return AdjacentAccessRestrictionData{
		PlmnId: HexBytes{0x62, 0xf2, 0x10},
		AccessRestrictionData: AccessRestrictionData{
			UtranNotAllowed: true,
			GeranNotAllowed: true,
		},
		ExtAccessRestrictionData: &ExtAccessRestrictionData{NrAsSecondaryRATNotAllowed: true},
	}
}

func TestAdjacentAccessRestrictionData_RoundTrip(t *testing.T) {
	in := makeAdjacentEntry()
	w, err := convertAdjacentAccessRestrictionDataToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToAdjacentAccessRestrictionData(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestAdjacentAccessRestrictionData_BadPlmnId(t *testing.T) {
	in := makeAdjacentEntry()
	in.PlmnId = HexBytes{0x01}
	_, err := strictWire(convertAdjacentAccessRestrictionDataToWire(&in))
	if !matchesConstraint(err, "plmnId", "SIZE (3)") {
		t.Fatalf("want BER constraint error, got %v", err)
	}
}

func TestAdjacentAccessRestrictionDataList_BoundsRejected(t *testing.T) {
	_, err := strictWire(convertAdjacentAccessRestrictionDataListToWire(AdjacentAccessRestrictionDataList{}))
	if !matchesConstraint(err, "adjacentAccessRestrictionDataList", "SIZE (1..50)") {
		t.Fatalf("empty: want size error, got %v", err)
	}
	too := make(AdjacentAccessRestrictionDataList, 50+1)
	for i := range too {
		too[i] = makeAdjacentEntry()
	}
	_, err = strictWire(convertAdjacentAccessRestrictionDataListToWire(too))
	if !matchesConstraint(err, "adjacentAccessRestrictionDataList", "SIZE (1..50)") {
		t.Fatalf("over-max: want size error, got %v", err)
	}
}

func TestAdjacentAccessRestrictionDataList_RoundTrip(t *testing.T) {
	in := AdjacentAccessRestrictionDataList{makeAdjacentEntry(), makeAdjacentEntry()}
	w, err := convertAdjacentAccessRestrictionDataListToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToAdjacentAccessRestrictionDataList(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch")
	}
}

// ----------------------------------------------------------------------------
// IMSI-GroupId / list
// ----------------------------------------------------------------------------

func makeIMSIGroupEntry() IMSIGroupId {
	return IMSIGroupId{
		GroupServiceID: 0xDEADBEEF,
		PlmnId:         HexBytes{0x62, 0xf2, 0x10},
		LocalGroupID:   HexBytes{0x01, 0x02, 0x03, 0x04},
	}
}

func TestIMSIGroupId_RoundTrip(t *testing.T) {
	in := makeIMSIGroupEntry()
	w, err := convertIMSIGroupIdToWire(&in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToIMSIGroupId(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(&in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, *out)
	}
}

func TestIMSIGroupId_PlmnIdInvalid(t *testing.T) {
	in := makeIMSIGroupEntry()
	in.PlmnId = HexBytes{0x01, 0x02}
	_, err := strictWire(convertIMSIGroupIdToWire(&in))
	if !matchesConstraint(err, "plmnId", "SIZE (3)") {
		t.Fatalf("want BER constraint error, got %v", err)
	}
}

func TestIMSIGroupId_LocalGroupIDInvalid(t *testing.T) {
	in := makeIMSIGroupEntry()
	in.LocalGroupID = HexBytes{}
	_, err := strictWire(convertIMSIGroupIdToWire(&in))
	if !matchesConstraint(err, "local-Group-ID", "SIZE (1..10)") {
		t.Fatalf("empty: want BER constraint error, got %v", err)
	}
	in.LocalGroupID = make(HexBytes, 11)
	_, err = strictWire(convertIMSIGroupIdToWire(&in))
	if !matchesConstraint(err, "local-Group-ID", "SIZE (1..10)") {
		t.Fatalf("over-max: want BER constraint error, got %v", err)
	}
}

func TestIMSIGroupIdList_BoundsRejected(t *testing.T) {
	_, err := strictWire(convertIMSIGroupIdListToWire(IMSIGroupIdList{}))
	if !matchesConstraint(err, "imsi-Group-Id-List", "SIZE (1..50)") {
		t.Fatalf("empty: want size error, got %v", err)
	}
	too := make(IMSIGroupIdList, 50+1)
	for i := range too {
		too[i] = makeIMSIGroupEntry()
	}
	_, err = strictWire(convertIMSIGroupIdListToWire(too))
	if !matchesConstraint(err, "imsi-Group-Id-List", "SIZE (1..50)") {
		t.Fatalf("over-max: want size error, got %v", err)
	}
}

func TestIMSIGroupIdList_RoundTrip(t *testing.T) {
	in := IMSIGroupIdList{makeIMSIGroupEntry(), makeIMSIGroupEntry()}
	w, err := convertIMSIGroupIdListToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToIMSIGroupIdList(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch")
	}
}

// ----------------------------------------------------------------------------
// EDRX-Cycle-Length / list
// ----------------------------------------------------------------------------

func TestEDRXCycleLength_RoundTrip(t *testing.T) {
	in := &EDRXCycleLength{
		RatType:              UsedRatEUTRAN,
		EDRXCycleLengthValue: HexBytes{0x09},
	}
	w, err := convertEDRXCycleLengthToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToEDRXCycleLength(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestEDRXCycleLength_ValueWrongSize(t *testing.T) {
	cases := []HexBytes{{}, {0x01, 0x02}}
	for _, v := range cases {
		in := &EDRXCycleLength{RatType: UsedRatNBIOT, EDRXCycleLengthValue: v}
		_, err := strictWire(convertEDRXCycleLengthToWire(in))
		if !matchesConstraint(err, "eDRX-Cycle-Length-Value", "SIZE (1)") {
			t.Fatalf("len=%d: want BER constraint error, got %v", len(v), err)
		}
	}
}

func TestEDRXCycleLength_PreservesUnknownRAT(t *testing.T) {
	// Postel's law: spec is extensible — preserve unknown values.
	in := &EDRXCycleLength{RatType: UsedRatType(99), EDRXCycleLengthValue: HexBytes{0xff}}
	w, err := convertEDRXCycleLengthToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToEDRXCycleLength(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if out.RatType != 99 {
		t.Fatalf("want preserved RatType=99, got %d", out.RatType)
	}
}

func TestEDRXCycleLengthList_BoundsRejected(t *testing.T) {
	_, err := strictWire(convertEDRXCycleLengthListToWire(EDRXCycleLengthList{}))
	if !matchesConstraint(err, "eDRX-Cycle-Length-List", "SIZE (1..8)") {
		t.Fatalf("empty: want size error, got %v", err)
	}
	too := make(EDRXCycleLengthList, 8+1)
	for i := range too {
		too[i] = EDRXCycleLength{RatType: UsedRatEUTRAN, EDRXCycleLengthValue: HexBytes{0x09}}
	}
	_, err = strictWire(convertEDRXCycleLengthListToWire(too))
	if !matchesConstraint(err, "eDRX-Cycle-Length-List", "SIZE (1..8)") {
		t.Fatalf("over-max: want size error, got %v", err)
	}
}

func TestEDRXCycleLengthList_RoundTrip(t *testing.T) {
	in := EDRXCycleLengthList{
		{RatType: UsedRatEUTRAN, EDRXCycleLengthValue: HexBytes{0x05}},
		{RatType: UsedRatNBIOT, EDRXCycleLengthValue: HexBytes{0x0a}},
	}
	w, err := convertEDRXCycleLengthListToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToEDRXCycleLengthList(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch")
	}
}

// ----------------------------------------------------------------------------
// Reset-Id-List
// ----------------------------------------------------------------------------

func TestResetIdList_RoundTrip(t *testing.T) {
	in := ResetIdList{
		HexBytes{0x01},
		HexBytes{0x01, 0x02, 0x03, 0x04},
	}
	w, err := convertResetIdListToWire(in)
	if err != nil {
		t.Fatalf("toWire: %v", err)
	}
	out, err := convertWireToResetIdList(w)
	if err != nil {
		t.Fatalf("fromWire: %v", err)
	}
	if !reflect.DeepEqual(in, out) {
		t.Fatalf("mismatch:\nin=%+v\nout=%+v", in, out)
	}
}

func TestResetIdList_BoundsRejected(t *testing.T) {
	_, err := strictWire(convertResetIdListToWire(ResetIdList{}))
	if !matchesConstraint(err, "reset-Id-List", "SIZE (1..50)") {
		t.Fatalf("empty: want size error, got %v", err)
	}
	too := make(ResetIdList, 50+1)
	for i := range too {
		too[i] = HexBytes{0x01}
	}
	_, err = strictWire(convertResetIdListToWire(too))
	if !matchesConstraint(err, "reset-Id-List", "SIZE (1..50)") {
		t.Fatalf("over-max: want size error, got %v", err)
	}
}

func TestResetIdList_PerEntrySize(t *testing.T) {
	cases := []HexBytes{{}, make(HexBytes, 5)}
	for _, c := range cases {
		_, err := convertResetIdListToWire(ResetIdList{c})
		if !errors.Is(err, ErrResetIdInvalidSize) {
			t.Fatalf("len=%d: want ErrResetIdInvalidSize, got %v", len(c), err)
		}
	}
}
