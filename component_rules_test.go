package gsmmap

import (
	"encoding/hex"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	phase2 "github.com/gomaja/go-asn1/telecom/ss7/gsm_map/gsm0902_v4_19_1"
)

// These pairs are transcribed from the 35 NU>1, OA>1, OP>1 and NU1 NU2
// annotations in GSM 09.02 V4.19.1 ASN.1.txt, independently of the rule table.
func TestComponentPhase2RuleSources(t *testing.T) {
	want := []string{
		"MAP-CH-DataTypes:2445", "MAP-CommonDataTypes:3288", "MAP-CommonDataTypes:3311",
		"MAP-ER-DataTypes:3096", "MAP-Errors:1897", "MAP-Errors:1929", "MAP-Errors:1958",
		"MAP-MS-DataTypes:2075", "MAP-MS-DataTypes:2296", "MAP-MobileServiceOperations:1003",
		"MAP-MobileServiceOperations:1050", "MAP-MobileServiceOperations:1109",
		"MAP-Protocol:717", "MAP-Protocol:725", "MAP-Protocol:742", "MAP-Protocol:759",
		"MAP-Protocol:761", "MAP-Protocol:779", "MAP-Protocol:790", "MAP-Protocol:798", "MAP-Protocol:801",
		"MAP-SM-DataTypes:2945", "MAP-SM-DataTypes:2952", "MAP-SM-DataTypes:2973", "MAP-SM-DataTypes:2989",
		"MAP-SS-DataTypes:2693", "MAP-SS-DataTypes:2731", "MAP-SS-DataTypes:2733",
		"MAP-SupplementaryServiceOperations:1435", "MAP-SupplementaryServiceOperations:1443",
		"MAP-SupplementaryServiceOperations:1459", "MAP-SupplementaryServiceOperations:1469",
		"MAP-SupplementaryServiceOperations:1486", "MAP-SupplementaryServiceOperations:1513",
		"MAP-SupplementaryServiceOperations:1539",
	}
	var got []string
	for _, r := range componentVersionRules() {
		if r.edition == EditionPhase2 {
			got = append(got, fmt.Sprintf("%s:%d", r.module, r.line))
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Phase 2 rule sources = %v; want %v", got, want)
	}
}

type componentRuleCase struct {
	id        string
	acn       OID
	role      componentRole
	op, error int64
	bad, good string
	other     OID // The same forbidden wire is valid outside the rule's version.
}

func componentRuleCases() []componentRuleCase {
	const p2 OID = "0.4.0.0.1.0.1.2"
	const ts2 OID = "0.4.0.0.1.0.19.2"
	const ts3 OID = "0.4.0.0.1.0.19.3"
	cases := []componentRuleCase{
		// ProvideRoamingNumberArg: [0] IMSI (3 octets), optional [1] address (2).
		{"p2-2445", p2, componentArgument, 4, 0, "30058003214365", "3009800321436581029121", "0.4.0.0.1.0.1.1"},
		// The [5] ExternalSignalInfo contains ENUMERATED and one-octet SignalInfo.
		{"p2-3288", p2, componentArgument, 4, 0, "3011800321436581029121a5060a0103040100", "3011800321436581029121a5060a0101040100", ts3},
		// UpdateLocationArg: IMSI, LocationInfo CHOICE [0]/[1], VLR address.
		{"p2-3311", p2, componentArgument, 2, 0, "300d04032143658002912104029143", "300d04032143658102912104029143", "0.4.0.0.1.0.1.1"},
		// SM-DeliveryFailureCause: ENUMERATED alternative / SEQUENCE alternative.
		{"p2-3096", p2, componentError, 46, 32, "0a0100", "30030a0100", "0.4.0.0.1.0.1.1"},
		{"p2-1897", p2, componentError, 2, 8, "", "0a0100", "0.4.0.0.1.0.1.1"},                             // RoamingNotAllowedCause.
		{"p2-1929", p2, componentError, 4, 27, "010100", "", "0.4.0.0.1.0.1.1"},                            // AbsentSubscriber BOOLEAN.
		{"p2-1958", p2, componentError, 12, 19, "820100", "", ""},                                          // SS-SubscriptionOption [2].
		{"p2-2075", p2, componentResult, 2, 0, "04029121", "300404029121", "0.4.0.0.1.0.1.1"},              // HLR address / extensible SEQUENCE.
		{"p2-2296", p2, componentArgument, 37, 0, "30070a010004029121", "300404029121", "0.4.0.0.1.0.1.1"}, // ResetArg.
		{"p2-1003", p2, componentError, 3, 5, "", "", "0.4.0.0.1.0.1.1"},
		{"p2-1050", p2, componentArgument, 29, 0, "", "30060a0101040100", "0.4.0.0.1.0.1.1"},
		{"p2-1109", p2, componentError, 43, 36, "", "", "0.4.0.0.1.0.1.1"},
		// RoutingInfoForSM-Arg: [0] address, [1] BOOLEAN, [2] address, optional [5] teleservice.
		{"p2-2945", p2, componentArgument, 45, 0, "300e8002912181010082029143850110", "300b8002912181010082029143", "0.4.0.0.1.0.1.1"},
		// RoutingInfoForSM-Res: IMSI, [0] LocationInfoWithLMSI, optional [2] BOOLEAN.
		{"p2-2952", p2, componentResult, 45, 0, "300e0403214365a004810291218201ff", "300b0403214365a00481029121", ""},
		// ForwardSM-Arg: DA CHOICE [3]/[0], OA [4] service centre, SignalInfo.
		{"p2-2973", p2, componentArgument, 46, 0, "300b8302912184029143040100", "300c800321436584029143040100", "0.4.0.0.1.0.1.1"},
		{"p2-2989", p2, componentArgument, 47, 0, "30080402912104029143", "300b04029121040291430a0100", "0.4.0.0.1.0.1.1"},
		{"p2-2693", p2, componentResult, 14, 0, "81029121", "800100", "0.4.0.0.1.0.1.1"}, // CHOICE address / one-octet SS-Status.
		{"p2-2731", p2, componentArgument, 18, 0, "0a0103", "0a0100", "0.4.0.0.1.0.1.1"},
		{"p2-2733", p2, componentArgument, 18, 0, "0a0104", "0a0100", "0.4.0.0.1.0.1.1"},
		{"p2-1435", p2, componentError, 10, 1, "", "", "0.4.0.0.1.0.1.1"},
		{"p2-1443", p2, componentError, 10, 19, "", "", "0.4.0.0.1.0.1.1"},
		{"p2-1459", p2, componentError, 11, 1, "", "", "0.4.0.0.1.0.1.1"},
		{"p2-1469", p2, componentError, 11, 19, "", "", "0.4.0.0.1.0.1.1"},
		{"p2-1486", p2, componentError, 12, 1, "", "", "0.4.0.0.1.0.1.1"},
		{"p2-1513", p2, componentError, 13, 1, "", "", "0.4.0.0.1.0.1.1"},
		{"p2-1539", p2, componentError, 14, 1, "", "", "0.4.0.0.1.0.1.1"},
		// These two CHOICEs distinguish the primitive version-2 cause from an extensible SEQUENCE.
		{"ts2-callBarred", ts2, componentError, 12, 13, "3000", "0a0100", ts3},
		{"ts2-systemFailure", ts2, componentError, 12, 34, "3000", "0a0100", ts3},
		{"ts3-callBarred", ts3, componentError, 12, 13, "0a0100", "3000", ts2},
		{"ts3-systemFailure", ts3, componentError, 12, 34, "0a0100", "3000", ts2},
		// ProvideRoamingNumberArg.[5] ExternalSignalInfo has an empty ExtensionContainer SEQUENCE.
		{"ts2-externalSignalInfo", ts2, componentArgument, 4, 0, "3013800321436581029121a5080a01010401003000", "3011800321436581029121a5060a0101040100", ts3},
	}
	for _, x := range []struct {
		line int
		op   int64
	}{{717, 28}, {725, 30}, {742, 9}, {759, 52}, {761, 35}, {779, 19}, {790, 54}, {798, 48}, {801, 49}} {
		cases = append(cases, componentRuleCase{fmt.Sprintf("p2-%d", x.line), p2, componentArgument, x.op, 0, "", "", "0.4.0.0.1.0.1.1"})
	}
	for _, x := range []struct {
		name      string
		op, error int64
	}{
		{"DataMissingParam", 12, 35}, {"UnexpectedDataParam", 12, 36}, {"FacilityNotSupParam", 50, 21},
		{"UnknownSubscriberParam", 71, 1}, {"UnidentifiedSubParam", 50, 5}, {"IllegalSubscriberParam", 44, 9},
		{"IllegalEquipmentParam", 44, 12}, {"BearerServNotProvParam", 12, 10}, {"TeleservNotProvParam", 12, 11},
		{"AbsentSubscriberParam", 4, 27}, {"IllegalSSOperationParam", 12, 16}, {"SSNotAvailableParam", 62, 18},
		{"SSSubscriptionViolationParam", 12, 19},
	} {
		// All 13 source parameter types are SEQUENCEs with only optional root components.
		cases = append(cases, componentRuleCase{"ts2-" + x.name, ts2, componentError, x.op, x.error, "3000", "", ts3})
	}
	return cases
}

func TestComponentVersionRules(t *testing.T) {
	cases := componentRuleCases()
	seen := make(map[string]bool)
	for _, tc := range cases {
		seen[tc.id] = true
		t.Run(tc.id, func(t *testing.T) {
			codec := componentRuleCodec(t, tc.acn)
			bad := componentRuleHex(t, tc.bad)
			_, err := componentRuleParse(codec, tc.role, tc.op, tc.error, bad)
			if !errors.Is(err, ErrVersionRule) {
				t.Fatalf("Parse(%x) = %v; want ErrVersionRule", bad, err)
			}
			if !strings.Contains(err.Error(), codec.Edition().String()) {
				t.Fatalf("error lacks edition: %v", err)
			}
			for _, rule := range componentVersionRules() {
				if rule.id == tc.id && (!strings.Contains(err.Error(), rule.module) || !strings.Contains(err.Error(), rule.text)) {
					t.Fatalf("error lacks source/comment: %v", err)
				}
				if rule.id == tc.id {
					location := "section " + rule.clause
					if rule.line != 0 {
						location = fmt.Sprintf("ASN.1.txt line %d", rule.line)
						if rule.endLine != rule.line {
							location = fmt.Sprintf("ASN.1.txt lines %d-%d", rule.line, rule.endLine)
						}
					}
					if !strings.Contains(err.Error(), location) {
						t.Fatalf("error lacks exact source location %q: %v", location, err)
					}
				}
			}
			value := componentRuleNative(t, codec, tc, bad)
			if _, err = componentRuleMarshal(codec, tc.role, tc.op, tc.error, value); !errors.Is(err, ErrVersionRule) {
				t.Fatalf("Marshal(%T) = %v; want ErrVersionRule", value, err)
			}
			// A forbidden operation or op/error pair has no conforming value in this
			// version; the no-parameter component in Phase 1 demonstrates its boundary.
			if tc.bad == "" && tc.good == "" && tc.other != "" {
				componentRuleRoundTrip(t, componentRuleCodec(t, tc.other), tc, bad)
			} else {
				componentRuleRoundTrip(t, codec, tc, componentRuleHex(t, tc.good))
			}
			if tc.other != "" && (tc.bad != "" || tc.good != "") {
				componentRuleRoundTrip(t, componentRuleCodec(t, tc.other), tc, bad)
			}
		})
	}
	for _, rule := range componentVersionRules() {
		if !seen[rule.id] {
			t.Errorf("rule %s has no breaking/conforming case", rule.id)
		}
	}
}

func componentRuleCodec(t *testing.T, acn OID) ComponentCodec {
	t.Helper()
	c, err := ComponentCodecFor(acn)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func componentRuleHex(t *testing.T, wire string) []byte {
	t.Helper()
	b, err := hex.DecodeString(wire)
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func componentRuleParse(c ComponentCodec, role componentRole, op, e int64, data []byte) (any, error) {
	switch role {
	case componentArgument:
		return c.ParseArgument(op, data)
	case componentResult:
		return c.ParseResult(op, data)
	default:
		return c.ParseError(op, e, data)
	}
}

func componentRuleMarshal(c ComponentCodec, role componentRole, op, e int64, value any) ([]byte, error) {
	switch role {
	case componentArgument:
		return c.MarshalArgument(op, value)
	case componentResult:
		return c.MarshalResult(op, value)
	default:
		return c.MarshalError(op, e, value)
	}
}

func componentRuleNative(t *testing.T, c ComponentCodec, tc componentRuleCase, data []byte) any {
	t.Helper()
	if len(data) == 0 {
		return nil
	}
	if c.Edition() == EditionPhase2 {
		operation, _ := phase2.LookupMacroOperation(tc.op)
		typeCodec := operation.Argument
		if tc.role == componentResult {
			typeCodec = operation.Result
		}
		if tc.role == componentError {
			errorInfo, _ := phase2.LookupMacroError(tc.error)
			typeCodec = errorInfo.Parameter
		}
		value, err := typeCodec.DecodeBER(data)
		if err != nil {
			t.Fatalf("source wire failed native decoder: %v", err)
		}
		return value
	}
	other := componentRuleCodec(t, tc.other)
	value, err := componentRuleParse(other, tc.role, tc.op, tc.error, data)
	if err != nil {
		t.Fatalf("source wire failed unaffected version: %v", err)
	}
	return value
}

func componentRuleRoundTrip(t *testing.T, c ComponentCodec, tc componentRuleCase, wire []byte) {
	t.Helper()
	value, err := componentRuleParse(c, tc.role, tc.op, tc.error, wire)
	if err != nil {
		t.Fatalf("conforming Parse(%x), edition %s: %v", wire, c.Edition(), err)
	}
	encoded, err := componentRuleMarshal(c, tc.role, tc.op, tc.error, value)
	if err != nil {
		t.Fatalf("conforming Marshal(%T): %v", value, err)
	}
	if !reflect.DeepEqual(encoded, wire) && (len(encoded) != 0 || len(wire) != 0) {
		t.Fatalf("conforming Marshal = %x; want %x", encoded, wire)
	}
}

func TestComponentForbiddenOperationAllRoles(t *testing.T) {
	c := componentRuleCodec(t, "0.4.0.0.1.0.1.2")
	for _, op := range []int64{28, 30, 9, 52, 35, 19, 54, 48, 49} {
		for _, role := range []componentRole{componentArgument, componentResult, componentError} {
			t.Run(fmt.Sprintf("operation%d-role%d", op, role), func(t *testing.T) {
				if _, err := componentRuleParse(c, role, op, 34, nil); !errors.Is(err, ErrVersionRule) {
					t.Fatalf("Parse = %v; want ErrVersionRule", err)
				}
				if _, err := componentRuleMarshal(c, role, op, 34, nil); !errors.Is(err, ErrVersionRule) {
					t.Fatalf("Marshal = %v; want ErrVersionRule", err)
				}
			})
		}
	}
}

func TestComponentVersionFourRules(t *testing.T) {
	c := componentRuleCodec(t, "0.4.0.0.1.0.19.4")
	for _, tc := range componentRuleCases() {
		if tc.id == "ts3-callBarred" || tc.id == "ts3-systemFailure" {
			t.Run(tc.id, func(t *testing.T) {
				data := componentRuleHex(t, tc.bad)
				if _, err := componentRuleParse(c, tc.role, tc.op, tc.error, data); !errors.Is(err, ErrVersionRule) {
					t.Fatalf("version 4 Parse = %v; want ErrVersionRule", err)
				}
				if _, err := componentRuleMarshal(c, tc.role, tc.op, tc.error, componentRuleNative(t, c, tc, data)); !errors.Is(err, ErrVersionRule) {
					t.Fatalf("version 4 Marshal = %v; want ErrVersionRule", err)
				}
				componentRuleRoundTrip(t, c, tc, componentRuleHex(t, tc.good))
			})
		}
	}
}

func TestComponentRuleRecursiveWalk(t *testing.T) {
	type node struct {
		Next     *node
		Value    any
		Children []node
		Octets   []byte
		Ignored  any `asn1:"-"`
		private  any
	}
	var rule componentVersionRule
	for _, r := range componentVersionRules() {
		if r.id == "p2-3288" {
			rule = r
			break
		}
	}
	cycle := &node{}
	cycle.Next = cycle
	badCycle := &node{Value: phase2.ProtocolId(3)}
	badCycle.Next = badCycle
	for _, tc := range []struct {
		name  string
		value any
		want  bool
	}{
		{"invalid", nil, false},
		{"nilPointer", (*node)(nil), false},
		{"nilInterface", node{}, false},
		{"cycle", cycle, false},
		{"cycleWithViolation", badCycle, true},
		{"sliceViolation", node{Children: []node{{Value: phase2.ProtocolId(3)}}}, true},
		{"sliceConforms", node{Children: []node{{Value: phase2.ProtocolId(1)}}}, false},
		{"opaqueOctets", node{Octets: []byte{3}, Ignored: phase2.ProtocolId(3), private: phase2.ProtocolId(3)}, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := rule.breaksValue(reflect.ValueOf(tc.value), make(map[uintptr]bool)); got != tc.want {
				t.Fatalf("recursive walk = %v; want %v", got, tc.want)
			}
		})
	}
}
