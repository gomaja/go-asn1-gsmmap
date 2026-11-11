package gsmmap

import (
	"bytes"
	"encoding/hex"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	phase1 "github.com/gomaja/go-asn1/telecom/ss7/gsm_map/gsm0902_v3_11_0"
	phase2 "github.com/gomaja/go-asn1/telecom/ss7/gsm_map/gsm0902_v4_19_1"
	asn1runtime "github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

func TestComponentApplicationContexts(t *testing.T) {
	cases := []struct {
		acn     OID
		version int
		edition Edition
	}{
		{"0.4.0.0.1.0.1.1", 1, EditionPhase1},
		{"0.4.0.0.1.0.1.2", 2, EditionPhase2},
		{"0.4.0.0.1.0.1.3", 3, EditionTS29002},
		{"0.4.0.0.1.0.1.4", 4, EditionTS29002},
		{"", 1, EditionPhase1},
	}
	for _, family := range []int{18, 19, 23, 26} {
		for v := 1; v <= 3; v++ {
			e := EditionTS29002
			if v == 1 {
				e = EditionPhase1
			}
			cases = append(cases, struct {
				acn     OID
				version int
				edition Edition
			}{OID(fmt.Sprintf("0.4.0.0.1.0.%d.%d", family, v)), v, e})
		}
	}
	for _, tc := range cases {
		name := string(tc.acn)
		if name == "" {
			name = "empty_acn"
		}
		t.Run(name, func(t *testing.T) {
			c, err := ComponentCodecFor(tc.acn)
			if err != nil {
				t.Fatal(err)
			}
			if c.Edition() != tc.edition || c.Version() != tc.version {
				t.Fatalf("got %v v%d, want %v v%d", c.Edition(), c.Version(), tc.edition, tc.version)
			}
			if tc.acn != "" {
				v, err := ApplicationContextVersion(tc.acn)
				if err != nil || v != tc.version {
					t.Fatalf("version: %d %v", v, err)
				}
			}
		})
	}
	for i, acn := range []OID{"0.4.0.0.1.1.1.1", "0.4.0.0.1.0.19", "0.4.0.0.1.0.19.2.1", "0.4.0.0.1.0.19.0", "0.4.0.0.1.0.19.02", "0.4.0.0.1.0.19.18446744073709551616", "garbage", MapDialogueAS()} {
		t.Run(fmt.Sprintf("invalid_%d_%s", i, acn), func(t *testing.T) {
			if _, err := ApplicationContextVersion(acn); !errors.Is(err, ErrNotMAPApplicationContext) {
				t.Fatalf("version error: %v", err)
			}
			if _, err := ComponentCodecFor(acn); !errors.Is(err, ErrNotMAPApplicationContext) {
				t.Fatalf("codec error: %v", err)
			}
		})
	}
}

func TestComponentEditionNames(t *testing.T) {
	for _, tc := range []struct {
		e    Edition
		name string
	}{{EditionPhase1, "GSM 09.02 V3.11.0"}, {EditionPhase2, "GSM 09.02 V4.19.1"}, {EditionTS29002, "3GPP TS 29.002 V19.1.0"}, {0, "Edition(0)"}, {255, "Edition(255)"}} {
		if got := tc.e.String(); got != tc.name {
			t.Errorf("%d: %s", tc.e, got)
		}
	}
}

func TestComponentContextBoundaries(t *testing.T) {
	var got []string
	for family := range componentVersion2Families() {
		got = append(got, family)
	}
	slices.Sort(got)
	if !reflect.DeepEqual(got, []string{"18", "19", "23", "26"}) {
		t.Fatalf("v2-only families %v", got)
	}
	for _, acn := range []OID{"", "0.4.0.0.1.0.19.02", "bad"} {
		if _, err := ApplicationContextVersion(acn); !errors.Is(err, asn1runtime.ErrInvalidOID) {
			t.Errorf("%q parser error not wrapped: %v", acn, err)
		}
	}
	// OID arcs are unbounded. Only the version must fit in int; an unassigned
	// family, including one above uint64, is allowed (membership is the caller's).
	c, err := ComponentCodecFor("0.4.0.0.1.0.18446744073709551616.2")
	if err != nil || c.Edition() != EditionPhase2 {
		t.Fatalf("large family: %v %v", c, err)
	}
	if _, err := ApplicationContextVersion(OID("0.4.0.0.1.0.1." + fmt.Sprint(int(^uint(0)>>1)))); err != nil {
		t.Fatal(err)
	}
}

func componentHex(t testing.TB, s string) []byte {
	t.Helper()
	v, err := hex.DecodeString(s)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

type componentWireCase struct {
	name          string
	acn           OID
	role          componentRole
	op, errorCode int64
	wire          string
	wantType      reflect.Type
}

// All lengths below count the child TLVs by hand. The Phase 1 fields come
// from V3.11.0's generated macro types (the source PDF has no text layer).
func componentWires() []componentWireCase {
	const p1 OID = "0.4.0.0.1.0.1.1"
	const p2 OID = "0.4.0.0.1.0.1.2"
	const ts2 OID = "0.4.0.0.1.0.19.2"
	const ts3 OID = "0.4.0.0.1.0.20.3"
	return []componentWireCase{
		// RoutingInfoForSM-Arg: [0] address 4 + [1] BOOLEAN 3 + [2] address 4 = 11.
		{"phase2_sri_arg", p2, componentArgument, 45, 0, "300b8002912181010082029143", reflect.TypeFor[phase2.RoutingInfoForSMArg]()},
		// IMSI TLV 5 + [0] LocationInfoWithLMSI TLV 6 = 11.
		{"phase2_sri_res", p2, componentResult, 45, 0, "300b0403214365a00481029121", reflect.TypeFor[phase2.RoutingInfoForSMRes]()},
		// ForwardSM-Arg: DA 5 + OA NULL 2 + UI 3 + moreMessages NULL 2 = 12.
		{"phase2_forward_sm", p2, componentArgument, 46, 0, "300c800321436585000401000500", reflect.TypeFor[phase2.ForwardSMArg]()},
		// SendAuthenticationInfoArg is IMSI, primitive OCTET STRING of 3 bytes.
		{"phase2_sai_arg", p2, componentArgument, 56, 0, "0403214365", reflect.TypeFor[phase2.BERValue[phase2.SendAuthenticationInfoArg]]()},
		// AuthenticationSet: RAND 18 + SRES 6 + KC 10 = 34; list child 36.
		{"phase2_sai_res", p2, componentResult, 56, 0, "302430220410000102030405060708090a0b0c0d0e0f04040102030404080102030405060708", reflect.TypeFor[phase2.SendAuthenticationInfoRes]()},
		// UpdateLocationArg: IMSI 5 + LocationInfo [1] 4 + VLR address 4 = 13.
		{"phase2_update_arg", p2, componentArgument, 2, 0, "300d04032143658102912104029143", reflect.TypeFor[phase2.UpdateLocationArg]()},
		// ExtensibleUpdateLocationRes: HLR address TLV 4; CHOICE uses SEQUENCE tag.
		{"phase2_update_res", p2, componentResult, 2, 0, "300404029121", reflect.TypeFor[phase2.UpdateLocationRes]()},
		// ResetArg: networkResource absent; HLR address TLV 4.
		{"phase2_reset", p2, componentArgument, 37, 0, "300404029121", reflect.TypeFor[phase2.ResetArg]()},
		// ExternalSignalInfo: ProtocolId ENUMERATED 3 + signal OCTET STRING 3 = 6.
		{"phase2_send_end", p2, componentArgument, 29, 0, "30060a0101040100", reflect.TypeFor[phase2.ExternalSignalInfo]()},
		// CallBarringCause operatorBarring(1), ENUMERATED of one octet.
		{"phase2_call_barred", p2, componentError, 45, 13, "0a0101", reflect.TypeFor[phase2.BERValue[phase2.CallBarringCause]]()},
		// SRI argument's 11 root octets + unknown context tag 127 TLV 4 = 15.
		{"phase2_unknown_addition", p2, componentArgument, 45, 0, "300f80029121810100820291439f7f0100", reflect.TypeFor[phase2.RoutingInfoForSMArg]()},
		// V3.11.0 UpdateLocationArgument: IMSI 5 + MSC [1] 4 + VLR 4 = 13.
		{"phase1_update_arg", p1, componentArgument, 2, 0, "300d04032143658102912104029143", reflect.TypeFor[phase1.UpdateLocationArgument]()},
		// UpdateLocation RESULT is IsdnAddressString, OCTET STRING of 2 bytes.
		{"phase1_update_res", p1, componentResult, 2, 0, "04029121", reflect.TypeFor[phase1.BERValue[phase1.IsdnAddressString]]()},
		// ForwardShortMessageArgument: DA IMSI [0] 5 + OA service centre [4] 4 + UI 3 = 12.
		{"phase1_forward_sm", p1, componentArgument, 46, 0, "300c800321436584029121040100", reflect.TypeFor[phase1.ForwardShortMessageArgument]()},
		// SendParametersArgument: SubscriberId IMSI [0] 5 + RequestParameters list (3+2) = 10.
		{"phase1_parameters_arg", p1, componentArgument, 9, 0, "300a800321436530030a0100", reflect.TypeFor[phase1.SendParametersArgument]()},
		// SentParameters: one SentParameter IMSI [0] OCTET STRING TLV of 5 bytes.
		{"phase1_parameters_res", p1, componentResult, 9, 0, "30058003214365", reflect.TypeFor[phase1.SentParameters]()},
		// SendRoutingInfoForSMArgument: address 4 + priority 3 + SC address 4 = 11.
		{"phase1_sri_arg", p1, componentArgument, 45, 0, "300b8002912181010082029143", reflect.TypeFor[phase1.SendRoutingInfoForSMArgument]()},
		// V3.11.0 CallBarred.parameter ENUMERATED one octet, not the TS CHOICE.
		{"phase1_call_barred", p1, componentError, 45, 13, "0a0101", reflect.TypeFor[phase1.BERValue[phase1.CallBarredParameter]]()},
		// TS USSD-Arg: DCS 3 + string 3 + msisdn [0] 4 = 10.
		{"ts2_ussd_arg", ts2, componentArgument, 59, 0, "300a04010f04014180029121", reflect.TypeFor[gsm_map.USSDArg]()},
		// USSD-Res: DCS 3 + string 3 = 6.
		{"ts2_ussd_res", ts2, componentResult, 59, 0, "300604010f040141", reflect.TypeFor[gsm_map.USSDRes]()},
		// AlertServiceCentreArg: MSISDN 4 + SC address 4 + IMSI 5 = 13.
		{"ts2_alert_imsi", "0.4.0.0.1.0.23.2", componentArgument, 64, 0, "300d04029121040291430403214365", reflect.TypeFor[gsm_map.AlertServiceCentreArg]()},
		// RegisterSS-Arg: mandatory SS-Code OCTET STRING of 1 byte; optional fields absent.
		{"ts2_register_ss", "0.4.0.0.1.0.18.2", componentArgument, 10, 0, "3003040111", reflect.TypeFor[gsm_map.RegisterSSArg]()},
		// sendIMSI argument is ISDN-AddressString (2 bytes); result IMSI (3 bytes).
		{"ts2_send_imsi_arg", "0.4.0.0.1.0.26.2", componentArgument, 58, 0, "04029121", reflect.TypeFor[gsm_map.BERValue[gsm_map.ISDNAddressString]]()},
		{"ts2_send_imsi_res", "0.4.0.0.1.0.26.2", componentResult, 58, 0, "0403214365", reflect.TypeFor[gsm_map.BERValue[gsm_map.IMSI]]()},
		// TS RoutingInfoForSM-Arg has the same mandatory root TLVs as Phase 2.
		{"ts3_sri_arg", ts3, componentArgument, 45, 0, "300b8002912181010082029143", reflect.TypeFor[gsm_map.RoutingInfoForSMArg]()},
		// The unknown TLV remains in ExtData_ with replay explicitly disabled.
		{"ts3_unknown_addition", ts3, componentArgument, 45, 0, "300f80029121810100820291439f7f0100", reflect.TypeFor[gsm_map.RoutingInfoForSMArg]()},
	}
}

func componentParse(c ComponentCodec, role componentRole, op, errorCode int64, data []byte) (any, error) {
	switch role {
	case componentArgument:
		return c.ParseArgument(op, data)
	case componentResult:
		return c.ParseResult(op, data)
	default:
		return c.ParseError(op, errorCode, data)
	}
}
func componentMarshal(c ComponentCodec, role componentRole, op, errorCode int64, value any) ([]byte, error) {
	switch role {
	case componentArgument:
		return c.MarshalArgument(op, value)
	case componentResult:
		return c.MarshalResult(op, value)
	default:
		return c.MarshalError(op, errorCode, value)
	}
}
func TestComponentRoundTrips(t *testing.T) {
	for _, tc := range componentWires() {
		t.Run(tc.name, func(t *testing.T) {
			c, err := ComponentCodecFor(tc.acn)
			if err != nil {
				t.Fatal(err)
			}
			wire := componentHex(t, tc.wire)
			value, err := componentParse(c, tc.role, tc.op, tc.errorCode, wire)
			if err != nil {
				t.Fatal(err)
			}
			if reflect.TypeOf(value) != tc.wantType {
				t.Fatalf("type %T, want %v", value, tc.wantType)
			}
			encoded, err := componentMarshal(c, tc.role, tc.op, tc.errorCode, value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(encoded, wire) {
				t.Fatalf("got %x, want %x", encoded, wire)
			}
		})
	}
}

func TestComponentCatalogues(t *testing.T) {
	for _, tc := range []struct {
		edition   Edition
		acn       OID
		ops, errs int
	}{
		{EditionPhase1, "0.4.0.0.1.0.1.1", 53, 39}, {EditionPhase2, "0.4.0.0.1.0.1.2", 46, 38}, {EditionTS29002, "0.4.0.0.1.0.1.3", 70, 56},
	} {
		t.Run(tc.edition.String(), func(t *testing.T) {
			c, _ := ComponentCodecFor(tc.acn)
			cat := componentCatalogues()[tc.edition]
			if len(cat.operations) != tc.ops || len(cat.errors) != tc.errs {
				t.Fatalf("counts operations %d errors %d", len(cat.operations), len(cat.errors))
			}
			for code, op := range cat.operations {
				t.Run(fmt.Sprintf("%s_%d", op.name, code), func(t *testing.T) {
					forbidden := tc.edition == EditionPhase2 && slices.Contains([]int64{28, 30, 9, 52, 35, 19, 54, 48, 49}, code)
					for _, role := range []componentRole{componentArgument, componentResult} {
						var want error
						if forbidden || (tc.edition == EditionPhase2 && code == 29 && role == componentArgument) {
							want = ErrVersionRule
						} else if role == componentResult && !op.resultPresent {
							want = ErrComponentNoResult
						}
						v, err := componentParse(c, role, code, 0, nil)
						if v != nil || !errors.Is(err, want) {
							t.Errorf("absent parse role%d: %T %v, want %v", role, v, err, want)
						}
						data, err := componentMarshal(c, role, code, 0, nil)
						if data != nil || !errors.Is(err, want) {
							t.Errorf("absent marshal role%d: %x %v, want %v", role, data, err, want)
						}
					}
					for errorCode := range op.errors {
						var want error
						pairForbidden := map[int64][]int64{3: {5}, 43: {36}, 10: {1, 19}, 11: {1, 19}, 12: {1}, 13: {1}, 14: {1}}
						if tc.edition == EditionPhase2 && (forbidden || errorCode == 8 || slices.Contains(pairForbidden[code], errorCode)) {
							want = ErrVersionRule
						}
						v, err := c.ParseError(code, errorCode, nil)
						if v != nil || !errors.Is(err, want) {
							t.Errorf("error%d absent parse: %T %v, want %v", errorCode, v, err, want)
						}
						data, err := c.MarshalError(code, errorCode, nil)
						if data != nil || !errors.Is(err, want) {
							t.Errorf("error%d absent marshal: %x %v, want %v", errorCode, data, err, want)
						}
					}
				})
			}
		})
	}
}

func TestComponentPhaseReferences(t *testing.T) {
	t.Run("Phase1", func(t *testing.T) {
		for _, op := range phase1.MacroOperations() {
			table := componentCatalogues()[EditionPhase1].operations[op.Code]
			if table.name != op.Name || (table.argument != nil) != (op.Argument != nil) || (table.result != nil) != (op.Result != nil) || table.resultPresent != op.ResultPresent {
				t.Fatalf("%s dispatch differs", op.Name)
			}
			want := make(map[int64]bool)
			for _, ref := range op.Errors {
				matches := 0
				var code int64
				for _, e := range phase1.MacroErrors() {
					if e.Module == ref.DefinitionModule && e.Definition == ref.Definition {
						matches++
						code = e.Code
					}
				}
				if matches != 1 {
					t.Errorf("%s reference %+v resolves %d times", op.Name, ref, matches)
				}
				want[code] = true
			}
			// The codec's ERRORS set is exactly the resolved source list.
			if !maps.Equal(table.errors, want) {
				t.Errorf("%s ERRORS %v, want %v", op.Name, table.errors, want)
			}
		}
	})
	t.Run("Phase2", func(t *testing.T) {
		for _, op := range phase2.MacroOperations() {
			table := componentCatalogues()[EditionPhase2].operations[op.Code]
			if table.name != op.Name || (table.argument != nil) != (op.Argument != nil) || (table.result != nil) != (op.Result != nil) || table.resultPresent != op.ResultPresent {
				t.Fatalf("%s dispatch differs", op.Name)
			}
			want := make(map[int64]bool)
			for _, ref := range op.Errors {
				matches := 0
				var code int64
				for _, e := range phase2.MacroErrors() {
					if e.Module == ref.DefinitionModule && e.Definition == ref.Definition {
						matches++
						code = e.Code
					}
				}
				if matches != 1 {
					t.Errorf("%s reference %+v resolves %d times", op.Name, ref, matches)
				}
				want[code] = true
			}
			// The codec's ERRORS set is exactly the resolved source list.
			if !maps.Equal(table.errors, want) {
				t.Errorf("%s ERRORS %v, want %v", op.Name, table.errors, want)
			}
		}
	})
	// sendAuthenticationInfo (56) lists SystemFailure, DataMissing,
	// UnexpectedDataValue and UnknownSubscriber (V4.19.1 ASN.1.txt lines
	// 1092-1096), not sm-DeliveryFailure (32).
	c, err := ComponentCodecFor("0.4.0.0.1.0.14.2")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.ParseError(56, 32, nil); !errors.Is(err, ErrComponentErrorNotListed) {
		t.Errorf("ParseError(56, 32) = %v; want ErrComponentErrorNotListed", err)
	}
	if _, err := c.MarshalError(56, 32, nil); !errors.Is(err, ErrComponentErrorNotListed) {
		t.Errorf("MarshalError(56, 32) = %v; want ErrComponentErrorNotListed", err)
	}
	// A value reference's resolved definition has the same identity as a macro
	// reference. Ambiguous and missing definitions are never admitted.
	defs := []componentErrorDefinition{{"M", "E", 7}, {"N", "E", 8}}
	if code, ok := componentResolveError("M", "E", defs); !ok || code != 7 {
		t.Fatalf("resolved %d %v", code, ok)
	}
	if _, ok := componentResolveError("M", "missing", defs); ok {
		t.Fatal("resolved missing error")
	}
	if _, ok := componentResolveError("M", "E", append(defs, defs[0])); ok {
		t.Fatal("resolved ambiguous error")
	}
}

func TestComponentTSCatalogueBindings(t *testing.T) {
	cat := componentCatalogues()[EditionTS29002]
	ops := gsm_map.Operations()
	errs := gsm_map.Errors()
	if len(ops) != len(cat.operations) || len(errs) != len(cat.errors) {
		t.Fatal("incomplete TS bindings")
	}
	for _, info := range ops {
		op, ok := cat.operations[int64(info.Code)]
		if !ok || op.name != info.Name {
			t.Fatalf("missing %s", info.Name)
		}
		for _, role := range []struct {
			typ *componentType
			ref *gsm_map.TypeRef
		}{{op.argument, info.Argument}, {op.result, info.Result}} {
			if (role.typ != nil) != (role.ref != nil) {
				t.Fatalf("%s codec presence", info.Name)
			}
			if role.ref != nil && role.typ.name != role.ref.ASN1Name {
				t.Fatalf("%s codec name", info.Name)
			}
		}
		if op.resultPresent != info.ReturnResult || len(op.errors) != len(info.Errors) {
			t.Fatalf("%s RESULT/ERRORS differs", info.Name)
		}
		for _, e := range info.Errors {
			if !op.errors[int64(e.Code)] {
				t.Errorf("%s missing error %s", info.Name, e.Name)
			}
		}
	}
	for _, info := range errs {
		e, ok := cat.errors[int64(info.Code)]
		if !ok || e.name != info.Name {
			t.Fatalf("missing error %s", info.Name)
		}
		if (e.parameter != nil) != (info.Parameter != nil) {
			t.Fatalf("%s parameter presence", info.Name)
		}
		if info.Parameter != nil && e.parameter.name != info.Parameter.ASN1Name {
			t.Fatalf("%s parameter name", info.Name)
		}
	}
}

func TestComponentAllRoleValueTypes(t *testing.T) {
	for edition := EditionPhase1; edition <= EditionTS29002; edition++ {
		c, _ := ComponentCodecFor(OID(fmt.Sprintf("0.4.0.0.1.0.1.%d", edition)))
		cat := componentCatalogues()[edition]
		check := func(role componentRole, op, code int64, typ *componentType) {
			if typ == nil {
				return
			}
			t.Run(fmt.Sprintf("%d_%d_%d_%d_%s", edition, role, op, code, typ.name), func(t *testing.T) {
				if typ.valueType == nil {
					t.Fatal("no dynamic type witness")
				}
				value := reflect.Zero(typ.valueType).Interface()
				_, err := componentMarshal(c, role, op, code, value)
				// Invalid zero values may fail ASN.1 constraints or version rules, but
				// the actual typed codec return type must pass the pre-encoding guard.
				if errors.Is(err, ErrComponentValueType) {
					t.Fatalf("decoder's type %T refused: %v", value, err)
				}
				_, err = componentMarshal(c, role, op, code, reflect.New(typ.valueType).Interface())
				if !errors.Is(err, ErrComponentValueType) && !errors.Is(err, ErrVersionRule) {
					t.Fatalf("pointer accepted or wrong error: %v", err)
				}
			})
		}
		for op, info := range cat.operations {
			check(componentArgument, op, 0, info.argument)
			if info.resultPresent {
				check(componentResult, op, 0, info.result)
			}
			for code := range info.errors {
				check(componentError, op, code, cat.errors[code].parameter)
			}
		}
	}
}

func TestComponentSentinels(t *testing.T) {
	var zero ComponentCodec
	if zero.Edition() != 0 || zero.Version() != 0 {
		t.Fatal("zero accessors")
	}
	for role := componentArgument; role <= componentError; role++ {
		for _, data := range [][]byte{nil, {5, 0}} {
			if _, err := componentParse(zero, role, 0, 0, data); !errors.Is(err, ErrComponentCodecUnset) {
				t.Fatalf("unset parse %v", err)
			}
		}
		if _, err := componentMarshal(zero, role, 0, 0, struct{}{}); !errors.Is(err, ErrComponentCodecUnset) {
			t.Fatalf("unset marshal %v", err)
		}
	}
	for edition := EditionPhase1; edition <= EditionTS29002; edition++ {
		c, _ := ComponentCodecFor(OID(fmt.Sprintf("0.4.0.0.1.0.1.%d", edition)))
		for role := componentArgument; role <= componentError; role++ {
			if _, err := componentParse(c, role, -1, -1, nil); !errors.Is(err, ErrComponentUnknownOperation) {
				t.Errorf("unknown parse: %v", err)
			}
			if _, err := componentMarshal(c, role, -1, -1, nil); !errors.Is(err, ErrComponentUnknownOperation) {
				t.Errorf("unknown marshal: %v", err)
			}
		}
		for _, tc := range []struct {
			role     componentRole
			op, code int64
			want     error
		}{{componentError, 2, -1, ErrComponentUnknownError}, {componentError, 2, 13, ErrComponentErrorNotListed}, {componentResult, 37, 0, ErrComponentNoResult}, {componentArgument, 45, 0, ErrComponentValueType}} {
			var v any
			if tc.want == ErrComponentNoResult && edition == EditionTS29002 {
				continue
			}
			if tc.want == ErrComponentValueType {
				v = "wrong type"
			}
			if _, err := componentMarshal(c, tc.role, tc.op, tc.code, v); !errors.Is(err, tc.want) {
				t.Errorf("edition%d marshal %+v: %v", edition, tc, err)
			}
			if tc.want != ErrComponentValueType {
				if _, err := componentParse(c, tc.role, tc.op, tc.code, nil); !errors.Is(err, tc.want) {
					t.Errorf("edition%d parse %+v: %v", edition, tc, err)
				}
			}
		}
		// Find all source roles without a type and exercise unexpected presence.
		for op, info := range componentCatalogues()[edition].operations {
			if info.argument == nil {
				if _, err := c.ParseArgument(op, []byte{5, 0}); !errors.Is(err, ErrComponentUnexpectedParameter) && !errors.Is(err, ErrVersionRule) {
					t.Errorf("unexpected parse argument %v", err)
				}
				if _, err := c.MarshalArgument(op, struct{}{}); !errors.Is(err, ErrComponentUnexpectedParameter) && !errors.Is(err, ErrVersionRule) {
					t.Errorf("unexpected marshal argument %v", err)
				}
			}
			if info.resultPresent && info.result == nil {
				if _, err := c.ParseResult(op, []byte{5, 0}); !errors.Is(err, ErrComponentUnexpectedParameter) && !errors.Is(err, ErrVersionRule) {
					t.Errorf("unexpected parse result %v", err)
				}
				if _, err := c.MarshalResult(op, struct{}{}); !errors.Is(err, ErrComponentUnexpectedParameter) && !errors.Is(err, ErrVersionRule) {
					t.Errorf("unexpected marshal result %v", err)
				}
			}
			for code := range info.errors {
				if componentCatalogues()[edition].errors[code].parameter == nil {
					if _, err := c.ParseError(op, code, []byte{5, 0}); !errors.Is(err, ErrComponentUnexpectedParameter) && !errors.Is(err, ErrVersionRule) {
						t.Errorf("unexpected error parameter parse %v", err)
					}
					if _, err := c.MarshalError(op, code, struct{}{}); !errors.Is(err, ErrComponentUnexpectedParameter) && !errors.Is(err, ErrVersionRule) {
						t.Errorf("unexpected error parameter marshal %v", err)
					}
				}
			}
		}
	}
}

func TestComponentFreshEncoding(t *testing.T) {
	for _, acn := range []OID{"0.4.0.0.1.0.1.1", "0.4.0.0.1.0.1.2", "0.4.0.0.1.0.1.3"} {
		t.Run(string(acn), func(t *testing.T) {
			c, _ := ComponentCodecFor(acn)
			// BER BOOLEAN true may be any nonzero octet. Fresh encoding writes FF;
			// replay would preserve 01. Child sizes: 4 + 3 + 4 = 11.
			value, err := c.ParseArgument(45, componentHex(t, "300b8002912181010182029143"))
			if err != nil {
				t.Fatal(err)
			}
			wire, err := c.MarshalArgument(45, value)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(wire, componentHex(t, "300b800291218101ff82029143")) {
				t.Fatalf("replayed instead of fresh: %x", wire)
			}
		})
	}
	// V3.11.0 §6.1 allows indefinite constructors; fresh encoding uses definite.
	c, _ := ComponentCodecFor("")
	value, err := c.ParseResult(9, componentHex(t, "308080032143650000"))
	if err != nil {
		t.Fatal(err)
	}
	wire, err := c.MarshalResult(9, value)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(wire, componentHex(t, "30058003214365")) {
		t.Fatalf("indefinite replay: %x", wire)
	}
}

func TestComponentCodecDiagnostics(t *testing.T) {
	for _, acn := range []OID{"0.4.0.0.1.0.1.1", "0.4.0.0.1.0.1.2", "0.4.0.0.1.0.1.3"} {
		t.Run(string(acn), func(t *testing.T) {
			c, _ := ComponentCodecFor(acn)
			for _, tc := range []struct {
				role     componentRole
				op, code int64
			}{{componentArgument, 45, 0}, {componentResult, 45, 0}, {componentError, 45, 13}} {
				typ, err := c.componentParameter(tc.op, tc.code, tc.role)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := componentParse(c, tc.role, tc.op, tc.code, []byte{0xff}); err == nil || !strings.HasPrefix(err.Error(), "decoding "+typ.name+": ") {
					t.Fatalf("decode wrapper: %v", err)
				}
				if tc.role != componentError {
					value := reflect.Zero(typ.valueType).Interface()
					if _, err := componentMarshal(c, tc.role, tc.op, tc.code, value); err == nil || !strings.HasPrefix(err.Error(), "encoding "+typ.name+": ") {
						t.Fatalf("encode wrapper: %v", err)
					}
				}
			}
			// The shortest definite length and primitive OCTET STRING requirements
			// are each edition's own strict profile; none of these APIs is tolerant.
			for _, wire := range []string{"30810b8002912181010082029143", "300da0040402912181010082029143"} {
				b := componentHex(t, wire)
				if _, err := c.ParseArgument(45, b); err == nil {
					t.Fatalf("strict profile accepted %x", b)
				}
			}
		})
	}
}

func TestComponentNativeRoleTypeWitnesses(t *testing.T) {
	check := func(edition Edition, name string, typ *componentType, encode func(any, ...ber.EncodeOption) ([]byte, error)) {
		t.Helper()
		if typ == nil {
			return
		}
		_, err := encode(reflect.Zero(typ.valueType).Interface())
		if err != nil && strings.Contains(err.Error(), "native macro role expects") {
			t.Errorf("%v %s witness %v is not decoder's type: %v", edition, name, typ.valueType, err)
		}
	}
	for _, op := range phase1.MacroOperations() {
		info := componentCatalogues()[EditionPhase1].operations[op.Code]
		if op.Argument != nil {
			check(EditionPhase1, op.Name+" argument", info.argument, op.Argument.EncodeBER)
		}
		if op.Result != nil {
			check(EditionPhase1, op.Name+" result", info.result, op.Result.EncodeBER)
		}
	}
	for _, e := range phase1.MacroErrors() {
		if e.Parameter != nil {
			check(EditionPhase1, e.Name, componentCatalogues()[EditionPhase1].errors[e.Code].parameter, e.Parameter.EncodeBER)
		}
	}
	for _, op := range phase2.MacroOperations() {
		info := componentCatalogues()[EditionPhase2].operations[op.Code]
		if op.Argument != nil {
			check(EditionPhase2, op.Name+" argument", info.argument, op.Argument.EncodeBER)
		}
		if op.Result != nil {
			check(EditionPhase2, op.Name+" result", info.result, op.Result.EncodeBER)
		}
	}
	for _, e := range phase2.MacroErrors() {
		if e.Parameter != nil {
			check(EditionPhase2, e.Name, componentCatalogues()[EditionPhase2].errors[e.Code].parameter, e.Parameter.EncodeBER)
		}
	}
}

// componentRepairWire is a RoutingInfoForSM-Arg (root 11 octets) with an
// unknown [127] addition of 8 octets: a SET of INTEGER 2 then INTEGER 1, which
// follows neither the SET nor the SET OF order of X.690 (02/2021) §§11.5, 11.6.
const componentRepairWire = "30168002912181010082029143bf7f083106020102020101"

// go-asn1 copies an unknown addition on fresh encoding only if it passes its
// DER element checks (runtime/ber WithoutBERReplay). This one parses, cannot be
// re-encoded, and re-encodes as the bare root once ExtData_ drops it.
func TestComponentOpaqueAdditionNeedsRepair(t *testing.T) {
	root := componentHex(t, "300b8002912181010082029143")
	for _, acn := range []OID{"0.4.0.0.1.0.20.2", "0.4.0.0.1.0.20.3"} {
		t.Run(string(acn), func(t *testing.T) {
			c, err := ComponentCodecFor(acn)
			if err != nil {
				t.Fatal(err)
			}
			value, err := c.ParseArgument(45, componentHex(t, componentRepairWire))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := c.MarshalArgument(45, value); !errors.Is(err, ber.ErrInvalidValue) {
				t.Fatalf("Marshal = %v; want ber.ErrInvalidValue", err)
			}
			var cleared any
			switch v := value.(type) {
			case phase2.RoutingInfoForSMArg:
				v.ExtData_ = nil
				cleared = v
			case gsm_map.RoutingInfoForSMArg:
				v.ExtData_ = nil
				cleared = v
			default:
				t.Fatalf("value type %T", value)
			}
			got, err := c.MarshalArgument(45, cleared)
			if err != nil || !bytes.Equal(got, root) {
				t.Fatalf("Marshal without ExtData_ = %x, %v; want %x", got, err, root)
			}
		})
	}
}
