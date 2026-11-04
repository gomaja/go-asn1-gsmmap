package gsmmap

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

func servingReportArg(nodes []SMServingNodeAddress) *ReportSMDeliveryStatus {
	return &ReportSMDeliveryStatus{
		MSISDN: "31611234567", MSISDNNature: 0x10, MSISDNPlan: 1,
		ServiceCentreAddress: "31611111111", ServiceCentreAddressNature: 0x10, ServiceCentreAddressPlan: 1,
		SmDeliveryOutcome: SmDeliveryAbsentSubscriber, FailedSMServingNodes: nodes,
	}
}

type servingNodeCase struct {
	name, path string
	message    func([]SMServingNodeAddress) marshaler
	parse      func([]byte) (marshaler, error)
	nodes      func(marshaler) []SMServingNodeAddress
	toWire     func([]SMServingNodeAddress) (*gsm_map.SMServingNodeAddressList, error)
	fromWire   func(*gsm_map.SMServingNodeAddressList) ([]SMServingNodeAddress, error)
	wire       func(*gsm_map.SMServingNodeAddressList) semWire
}

func servingNodeCases() []servingNodeCase {
	argWire := func(nodes *gsm_map.SMServingNodeAddressList) *gsm_map.ReportSMDeliveryStatusArg {
		return &gsm_map.ReportSMDeliveryStatusArg{
			Msisdn:               []byte{0x91, 0x13, 0x16, 0x21, 0x43, 0x65, 0xf7},
			ServiceCentreAddress: []byte{0x91, 0x13, 0x16, 0x11, 0x11, 0x11, 0xf1},
			SmDeliveryOutcome:    gsm_map.SMDeliveryOutcomeAbsentSubscriber, FailedSMServingNodes: nodes,
		}
	}
	return []servingNodeCase{
		{
			name: "failedSMServingNodes", path: "failedSMServingNodes",
			message: func(nodes []SMServingNodeAddress) marshaler { return servingReportArg(nodes) },
			parse:   asParser(ParseReportSMDeliveryStatus),
			nodes:   func(m marshaler) []SMServingNodeAddress { return m.(*ReportSMDeliveryStatus).FailedSMServingNodes },
			toWire: func(nodes []SMServingNodeAddress) (*gsm_map.SMServingNodeAddressList, error) {
				w, err := convertReportSMDeliveryStatusToArg(servingReportArg(nodes))
				if err != nil {
					return nil, err
				}
				return w.FailedSMServingNodes, nil
			},
			fromWire: func(nodes *gsm_map.SMServingNodeAddressList) ([]SMServingNodeAddress, error) {
				v, err := convertArgToReportSMDeliveryStatus(argWire(nodes))
				if err != nil {
					return nil, err
				}
				return v.FailedSMServingNodes, nil
			},
			wire: func(nodes *gsm_map.SMServingNodeAddressList) semWire { return argWire(nodes) },
		},
		{
			name: "registeredSMServingNodes", path: "registeredSMServingNodes",
			message: func(nodes []SMServingNodeAddress) marshaler {
				return &ReportSMDeliveryStatusRes{RegisteredSMServingNodes: nodes}
			},
			parse: asParser(ParseReportSMDeliveryStatusRes),
			nodes: func(m marshaler) []SMServingNodeAddress {
				return m.(*ReportSMDeliveryStatusRes).RegisteredSMServingNodes
			},
			toWire: func(nodes []SMServingNodeAddress) (*gsm_map.SMServingNodeAddressList, error) {
				w, err := convertReportSMDeliveryStatusResToRes(&ReportSMDeliveryStatusRes{RegisteredSMServingNodes: nodes})
				if err != nil {
					return nil, err
				}
				return w.RegisteredSMServingNodes, nil
			},
			fromWire: func(nodes *gsm_map.SMServingNodeAddressList) ([]SMServingNodeAddress, error) {
				v, err := convertResToReportSMDeliveryStatusRes(&gsm_map.ReportSMDeliveryStatusRes{RegisteredSMServingNodes: nodes})
				if err != nil {
					return nil, err
				}
				return v.RegisteredSMServingNodes, nil
			},
			wire: func(nodes *gsm_map.SMServingNodeAddressList) semWire {
				return &gsm_map.ReportSMDeliveryStatusRes{RegisteredSMServingNodes: nodes}
			},
		},
	}
}

// TS 29.002 V19.1.0 §17.7.6: [18]/[0] replace the list SEQUENCE tag;
// each element is the [0] number or [1] Diameter SEQUENCE CHOICE alternative.
// Hand lengths: mandatory Arg = 9+9+3 = 21; number TLV = 9, list = 11,
// so Arg content = 32 (0x20), Res content = 11 (0x0b). Diameter "node.test"/"node.test"
// has two eleven-octet fields, element = 24, list = 26, Arg content = 47.
func TestReportSMServingNodesWire(t *testing.T) {
	number := []SMServingNodeAddress{{NetworkNodeNumber: "31612222222", NetworkNodeNumberNature: 0x10, NetworkNodeNumberPlan: 1}}
	diameter := []SMServingNodeAddress{{DiameterAddress: &NetworkNodeDiameterAddress{DiameterName: HexBytes("node.test"), DiameterRealm: HexBytes("node.test")}}}
	for i, tc := range servingNodeCases() {
		for _, vector := range []struct {
			name  string
			nodes []SMServingNodeAddress
			wires [2]string
		}{
			{"number", number, [2]string{"30200407911316214365f70407911316111111f10a0101b2098007911316222222f2", "300ba0098007911316222222f2"}},
			{"Diameter", diameter, [2]string{"302f0407911316214365f70407911316111111f10a0101b218a11680096e6f64652e7465737481096e6f64652e74657374", "301aa018a11680096e6f64652e7465737481096e6f64652e74657374"}},
		} {
			t.Run(tc.name+"/"+vector.name, func(t *testing.T) {
				want, err := hex.DecodeString(vector.wires[i])
				if err != nil {
					t.Fatal(err)
				}
				in := tc.message(vector.nodes)
				got, err := tc.parse(want)
				if err != nil {
					t.Fatalf("Parse literal: %v", err)
				}
				if !reflect.DeepEqual(got, in) {
					t.Fatalf("Parse = %#v, want %#v", got, in)
				}
				encoded, err := in.Marshal()
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(encoded, want) {
					t.Fatalf("Marshal = %x, want %x", encoded, want)
				}
				replayed, err := got.Marshal()
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(replayed, want) {
					t.Fatalf("parsed Marshal = %x, want %x", replayed, want)
				}
			})
		}
	}
}

// TS 29.002 V19.1.0 §17.7.6 preserves list order and both CHOICE alternatives.
func TestReportSMServingNodesConverters(t *testing.T) {
	public := []SMServingNodeAddress{
		{NetworkNodeNumber: "12", NetworkNodeNumberNature: 0x20, NetworkNodeNumberPlan: 3},
		{DiameterAddress: &NetworkNodeDiameterAddress{DiameterName: HexBytes("node.test"), DiameterRealm: HexBytes("realm.test")}},
		{NetworkNodeNumber: "3", NetworkNodeNumberNature: 0x40, NetworkNodeNumberPlan: 9},
	}
	wire := &gsm_map.SMServingNodeAddressList{Values: []gsm_map.SMServingNodeAddress{
		gsm_map.NewSMServingNodeAddressNetworkNodeNumber([]byte{0xa3, 0x21}),
		gsm_map.NewSMServingNodeAddressDiameterAddress(gsm_map.NetworkNodeDiameterAddress{DiameterName: []byte("node.test"), DiameterRealm: []byte("realm.test")}),
		gsm_map.NewSMServingNodeAddressNetworkNodeNumber([]byte{0xc9, 0xf3}),
	}}
	for _, tc := range servingNodeCases() {
		t.Run(tc.name, func(t *testing.T) {
			gotWire, err := tc.toWire(public)
			if err != nil {
				t.Fatal(err)
			}
			if gotWire == nil || len(gotWire.Values) != len(wire.Values) {
				t.Fatalf("to wire = %#v, want three elements", gotWire)
			}
			for i, want := range wire.Values {
				got := gotWire.Values[i]
				if got.Choice != want.Choice {
					t.Fatalf("element %d choice = %d, want %d", i, got.Choice, want.Choice)
				}
				if want.NetworkNodeNumber != nil {
					if got.NetworkNodeNumber == nil || got.DiameterAddress != nil || !bytes.Equal(*got.NetworkNodeNumber, *want.NetworkNodeNumber) {
						t.Fatalf("element %d number = %#v, want %x", i, got, *want.NetworkNodeNumber)
					}
				} else if got.DiameterAddress == nil || got.NetworkNodeNumber != nil ||
					!bytes.Equal(got.DiameterAddress.DiameterName, want.DiameterAddress.DiameterName) ||
					!bytes.Equal(got.DiameterAddress.DiameterRealm, want.DiameterAddress.DiameterRealm) {
					t.Fatalf("element %d Diameter address = %#v", i, got)
				}
			}
			gotPublic, err := tc.fromWire(wire)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(gotPublic, public) {
				t.Fatalf("from wire = %#v, want %#v", gotPublic, public)
			}
			senderMobilityRoundTrip(t, tc.message(public), tc.parse)
		})
	}
}

// TS 29.002 V19.1.0 §17.7.6 specifies SIZE (1..5); a non-nil empty
// slice is present and must reach the codec rather than becoming absent.
func TestReportSMServingNodesSize(t *testing.T) {
	for _, tc := range servingNodeCases() {
		for _, n := range []int{0, 5, 6} {
			t.Run(fmt.Sprintf("%s/%d", tc.name, n), func(t *testing.T) {
				nodes := make([]SMServingNodeAddress, n)
				wire := &gsm_map.SMServingNodeAddressList{Values: make([]gsm_map.SMServingNodeAddress, n)}
				for i := range nodes {
					nodes[i] = SMServingNodeAddress{NetworkNodeNumber: "1", NetworkNodeNumberNature: 0x10, NetworkNodeNumberPlan: 1}
					wire.Values[i] = gsm_map.NewSMServingNodeAddressNetworkNodeNumber([]byte{0x91, 0xf1})
				}
				data, err := tc.message(nodes).Marshal()
				if n == 5 {
					if err != nil {
						t.Fatal(err)
					}
					got, parseErr := tc.parse(data)
					if parseErr != nil {
						t.Fatal(parseErr)
					}
					if !reflect.DeepEqual(tc.nodes(got), nodes) {
						t.Fatalf("Parse nodes = %#v", tc.nodes(got))
					}
					return
				}
				wantValidationError(t, err, tc.path, "SIZE (1..5)", n)
				got, err := tc.parse(tolerantBER(t, tc.wire(wire)))
				wantValidationError(t, err, tc.path, "SIZE (1..5)", n)
				if got != nil {
					t.Fatalf("Parse = %#v, want nil", got)
				}
			})
		}
	}
}

// TS 29.002 V19.1.0 §17.7.6 selects exactly one SMServingNodeAddress alternative.
func TestReportSMServingNodesChoiceErrors(t *testing.T) {
	for _, tc := range servingNodeCases() {
		t.Run(tc.name, func(t *testing.T) {
			for _, bad := range []struct {
				name string
				node SMServingNodeAddress
				want error
			}{
				{"none", SMServingNodeAddress{}, ErrSMServingNodeAddressNoAlternative},
				{"both", SMServingNodeAddress{NetworkNodeNumber: "1", DiameterAddress: &NetworkNodeDiameterAddress{DiameterName: HexBytes("node.test"), DiameterRealm: HexBytes("node.test")}}, ErrSMServingNodeAddressMultipleAlternatives},
			} {
				t.Run(bad.name, func(t *testing.T) {
					_, err := tc.message([]SMServingNodeAddress{{NetworkNodeNumber: "1"}, bad.node}).Marshal()
					if !errors.Is(err, bad.want) || !strings.Contains(err.Error(), "[1]") {
						t.Fatalf("Marshal = %v, want indexed %v", err, bad.want)
					}
				})
			}
			for _, choice := range []int{0, 99} {
				_, err := tc.fromWire(&gsm_map.SMServingNodeAddressList{Values: []gsm_map.SMServingNodeAddress{{Choice: choice}}})
				if !errors.Is(err, ErrSMServingNodeAddressUnknownAlternative) || !strings.Contains(err.Error(), "[0]") {
					t.Fatalf("choice %d = %v", choice, err)
				}
			}
		})
	}
}

// Absence is unchanged by the §17.7.6 extension: nil slices omit the tags.
// Hand lengths: legacy Arg content is 21 (0x15); empty Res content is zero.
func TestReportSMServingNodesAbsent(t *testing.T) {
	for i, tc := range servingNodeCases() {
		t.Run(tc.name, func(t *testing.T) {
			legacy := []string{"30150407911316214365f70407911316111111f10a0101", "3000"}[i]
			want, err := hex.DecodeString(legacy)
			if err != nil {
				t.Fatal(err)
			}
			wire, err := tc.toWire(nil)
			if err != nil || wire != nil {
				t.Fatalf("to wire = %#v, %v", wire, err)
			}
			nodes, err := tc.fromWire(nil)
			if err != nil || nodes != nil {
				t.Fatalf("from wire = %#v, %v", nodes, err)
			}
			data, err := tc.message(nil).Marshal()
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(data, want) {
				t.Fatalf("Marshal = %x, want %x", data, want)
			}
			got, err := tc.parse(want)
			if err != nil {
				t.Fatal(err)
			}
			if tc.nodes(got) != nil {
				t.Fatalf("Parse nodes = %#v, want nil", tc.nodes(got))
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.6 selects a number; §17.7.8 permits its
// header without digits, but the public CHOICE cannot preserve that presence.
// Hand lengths: header-only element = 3, list = 5, Arg content = 26;
// filler-only element = 4, list = 6, Arg content = 27.
func TestReportSMServingNodesDigitlessWire(t *testing.T) {
	for i, tc := range servingNodeCases() {
		for _, wires := range [][2]string{
			{"301a0407911316214365f70407911316111111f10a0101b203800191", "3005a003800191"},
			{"301b0407911316214365f70407911316111111f10a0101b204800291ff", "3006a004800291ff"},
		} {
			t.Run(tc.name+"/"+wires[i], func(t *testing.T) {
				data, err := hex.DecodeString(wires[i])
				if err != nil {
					t.Fatal(err)
				}
				got, err := tc.parse(data)
				if !errors.Is(err, ErrSMServingNodeAddressNetworkNodeNumberDecodedEmpty) {
					t.Fatalf("Parse = %v", err)
				}
				if got != nil {
					t.Fatalf("Parse = %#v, want nil", got)
				}
			})
		}
	}
}

// TS 29.002 V19.1.0 §17.1.4 keeps received values. The sender's
// DiameterIdentity ASCII restriction (§17.7.8) must not reject Parse.
func TestReportSMServingNodesDiameterLenient(t *testing.T) {
	for _, tc := range servingNodeCases() {
		for _, member := range []string{"name", "realm"} {
			t.Run(tc.name+"/"+member, func(t *testing.T) {
				for pos := range 9 {
					for invalid := 0x80; invalid <= 0xff; invalid++ {
						name, realm := []byte("node.test"), []byte("node.test")
						if member == "name" {
							name[pos] = byte(invalid)
						} else {
							realm[pos] = byte(invalid)
						}
						wire := &gsm_map.SMServingNodeAddressList{Values: []gsm_map.SMServingNodeAddress{
							gsm_map.NewSMServingNodeAddressDiameterAddress(gsm_map.NetworkNodeDiameterAddress{DiameterName: name, DiameterRealm: realm}),
						}}
						data, err := tc.wire(wire).MarshalBER()
						if err != nil {
							t.Fatal(err)
						}
						got, err := tc.parse(data)
						if err != nil {
							t.Fatalf("pos %d byte %02x: %v", pos, invalid, err)
						}
						nodes := tc.nodes(got)
						if len(nodes) != 1 || nodes[0].DiameterAddress == nil {
							t.Fatalf("Parse nodes = %#v", nodes)
						}
						if !bytes.Equal(nodes[0].DiameterAddress.DiameterName, name) || !bytes.Equal(nodes[0].DiameterAddress.DiameterRealm, realm) {
							t.Fatal("Parse changed Diameter octets")
						}
						if _, err := got.Marshal(); !errors.Is(err, ErrDiameterIdentityNotASCII) {
							t.Fatalf("Marshal = %v, want ASCII error", err)
						}
					}
				}
			})
		}
	}
}

// TS 29.002 V19.1.0 §17.1.4 keeps received values; §17.7.8 assigns
// address nature/plan values. Reserved assignments are retained on receive
// and rejected by the sender without normalizing the stored header fields.
func TestReportSMServingNodesAddressLenient(t *testing.T) {
	for _, tc := range servingNodeCases() {
		for _, header := range []struct {
			value, nature, plan byte
			want                error
		}{
			{0xd1, 0x50, 1, ErrAddressNatureInvalid},
			{0xf1, 0x70, 1, ErrAddressNatureInvalid},
			{0x92, 0x10, 2, ErrAddressPlanInvalid},
			{0x9f, 0x10, 15, ErrAddressPlanInvalid},
		} {
			t.Run(fmt.Sprintf("%s/%02x", tc.name, header.value), func(t *testing.T) {
				wire := &gsm_map.SMServingNodeAddressList{Values: []gsm_map.SMServingNodeAddress{
					gsm_map.NewSMServingNodeAddressNetworkNodeNumber([]byte{header.value, 0x21}),
				}}
				want := []SMServingNodeAddress{{NetworkNodeNumber: "12", NetworkNodeNumberNature: header.nature, NetworkNodeNumberPlan: header.plan}}
				converted, err := tc.fromWire(wire)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(converted, want) {
					t.Fatalf("converter = %#v, want %#v", converted, want)
				}
				data, err := tc.wire(wire).MarshalBER()
				if err != nil {
					t.Fatal(err)
				}
				got, err := tc.parse(data)
				if err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(tc.nodes(got), want) {
					t.Fatalf("Parse = %#v, want %#v", tc.nodes(got), want)
				}
				if _, err := got.Marshal(); !errors.Is(err, header.want) {
					t.Fatalf("Marshal = %v, want %v", err, header.want)
				}
			})
		}
	}
}

// Both lists appear in JSON: null when absent, and each entry with its one
// alternative set, Diameter octets as hex like every other byte field.
func TestReportSMServingNodesJSON(t *testing.T) {
	const entries = `[{"NetworkNodeNumber":"12","NetworkNodeNumberNature":16,"NetworkNodeNumberPlan":1,"DiameterAddress":null},` +
		`{"NetworkNodeNumber":"","NetworkNodeNumberNature":0,"NetworkNodeNumberPlan":0,"DiameterAddress":{"DiameterName":"6e","DiameterRealm":"72"}}]`
	nodes := []SMServingNodeAddress{
		{NetworkNodeNumber: "12", NetworkNodeNumberNature: 0x10, NetworkNodeNumberPlan: 1},
		{DiameterAddress: &NetworkNodeDiameterAddress{DiameterName: HexBytes("n"), DiameterRealm: HexBytes("r")}},
	}
	for _, tc := range []struct {
		name, key string
		present   any
		absent    any
		decoded   any
	}{
		{"failedSMServingNodes", "FailedSMServingNodes", &ReportSMDeliveryStatus{FailedSMServingNodes: nodes}, &ReportSMDeliveryStatus{}, &ReportSMDeliveryStatus{}},
		{"registeredSMServingNodes", "RegisteredSMServingNodes", &ReportSMDeliveryStatusRes{RegisteredSMServingNodes: nodes}, &ReportSMDeliveryStatusRes{}, &ReportSMDeliveryStatusRes{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := json.Marshal(tc.absent)
			if err != nil {
				t.Fatal(err)
			}
			if want := `"` + tc.key + `":null`; !strings.Contains(string(got), want) {
				t.Errorf("absent JSON %s lacks %s", got, want)
			}
			got, err = json.Marshal(tc.present)
			if err != nil {
				t.Fatal(err)
			}
			if want := `"` + tc.key + `":` + entries; !strings.Contains(string(got), want) {
				t.Errorf("JSON %s lacks %s", got, want)
			}
			if err := json.Unmarshal(got, tc.decoded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(tc.decoded, tc.present) {
				t.Errorf("round trip = %#v, want %#v", tc.decoded, tc.present)
			}
		})
	}
}
