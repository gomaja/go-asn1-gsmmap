package gsmmap

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

// TS 29.002 V19.1.0 §17.7.1 CancelLocationArg restricts sending both
// MTRF flags but does not prescribe their rejection on receipt.
func TestReceiverCancelLocationMTRFFlags(t *testing.T) {
	for _, flags := range [][2]bool{{true, true}, {true, false}, {false, true}, {false, false}} {
		t.Run(fmt.Sprint(flags), func(t *testing.T) {
			w := &gsm_map.CancelLocationArg{Identity: gsm_map.NewIdentityImsi(semAddrIMSI), MtrfSupportedAndAuthorized: boolToNullPtr(flags[0]), MtrfSupportedAndNotAuthorized: boolToNullPtr(flags[1])}
			got, err := ParseCancelLocation(strictBER(t, w))
			if err != nil {
				t.Fatal(err)
			}
			if got.MtrfSupportedAndAuthorized != flags[0] || got.MtrfSupportedAndNotAuthorized != flags[1] {
				t.Fatalf("flags lost: %+v", got)
			}
			_, err = got.Marshal()
			if flags[0] && flags[1] {
				if !errors.Is(err, ErrCancelLocMtrfBothSet) {
					t.Fatalf("Marshal: %v", err)
				}
			} else if err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.1 TypeOfUpdate specifies sender applicability.
func TestReceiverCancelLocationTypeOfUpdate(t *testing.T) {
	for _, ct := range []*CancellationType{nil, new(CancellationTypeSubscriptionWithdraw), new(CancellationType(99)), new(CancellationTypeUpdateProcedure), new(CancellationTypeInitialAttachProcedure)} {
		t.Run(fmt.Sprint(ct), func(t *testing.T) {
			tou := TypeOfUpdateMmeChange
			w := &gsm_map.CancelLocationArg{Identity: gsm_map.NewIdentityImsi(semAddrIMSI), CancellationType: ct, TypeOfUpdate: &tou}
			got, err := ParseCancelLocation(strictBER(t, w))
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "CancellationType", ct, got.CancellationType)
			wantEqual(t, "TypeOfUpdate", &tou, got.TypeOfUpdate)
			_, err = got.Marshal()
			switch {
			case ct != nil && *ct == 99:
				if !errors.Is(err, ErrCancelLocInvalidCancellationType) {
					t.Fatalf("Marshal: %v", err)
				}
			case ct == nil || *ct == CancellationTypeSubscriptionWithdraw:
				if !errors.Is(err, ErrCancelLocTypeOfUpdateNotApplicable) {
					t.Fatalf("Marshal: %v", err)
				}
			default:
				if err != nil {
					t.Fatal(err)
				}
			}
		})
	}
}

func mobilityGprsWire(t *testing.T) *gsm_map.UpdateGprsLocationArg {
	t.Helper()
	w, err := convertUpdateGprsLocationToArg(&UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "12345", SGSNAddress: "192.0.2.2", SmsOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	return w
}

// TS 29.002 V19.1.0 §17.7.1 PDN-GW-Update explicitly requires the HSS
// to ignore the enclosing EPS-Info when pdn-gw-Identity is missing.
func TestReceiverEPSInfoMissingPDNIdentity(t *testing.T) {
	for _, tc := range []struct {
		name   string
		update gsm_map.PDNGWUpdate
		keep   bool
	}{
		{"empty", gsm_map.PDNGWUpdate{}, false},
		{"apn only", gsm_map.PDNGWUpdate{Apn: new([]byte{1, 'a'})}, false},
		{"context only", gsm_map.PDNGWUpdate{ContextId: new(int64(1))}, false},
		{"identity without apn or context", gsm_map.PDNGWUpdate{PdnGwIdentity: &gsm_map.PDNGWIdentity{PdnGwIpv4Address: new(gsm_map.PDPAddress{192, 0, 2, 1})}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			w := mobilityGprsWire(t)
			e := gsm_map.NewEPSInfoPdnGwUpdate(tc.update)
			w.EpsInfo = &e
			got, err := ParseUpdateGprsLocation(strictBER(t, w))
			if err != nil {
				t.Fatal(err)
			}
			if (got.EpsInfo != nil) != tc.keep {
				t.Fatalf("EPS-Info = %+v, keep %v", got.EpsInfo, tc.keep)
			}
			if got.IMSI != "001010123456789" || got.SgsnNumber != "12345" || got.SGSNAddress != "192.0.2.2" || !got.SmsOnly {
				t.Fatalf("surrounding fields lost: %+v", got)
			}
			if tc.keep && (got.EpsInfo.PdnGwUpdate.PdnGwIdentity == nil || got.EpsInfo.PdnGwUpdate.APN != nil || got.EpsInfo.PdnGwUpdate.ContextID != nil) {
				t.Fatalf("identity-only update changed: %+v", got.EpsInfo)
			}
		})
	}
}

// 3GPP TS 29.002 V19.1.0 §17.7.1 discards unknown ISR bit assignments.
// ITU-T X.680 (02/2021) §22.7 permits trimming named-bit zero padding.
func TestReceiverISRUnknownBits(t *testing.T) {
	for bitLen := 3; bitLen <= 8; bitLen++ {
		for _, b := range []byte{0, 0xE0, 0xFF, 0x10, 0x08, 0x04, 0x02, 0x01} {
			mask := byte(0xFF << uint(8-bitLen))
			raw := b & mask
			t.Run(fmt.Sprintf("%d/%02x", bitLen, raw), func(t *testing.T) {
				w := mobilityGprsWire(t)
				e := gsm_map.NewEPSInfoIsrInformation(runtime.BitString{Bytes: []byte{raw}, BitLength: bitLen})
				w.EpsInfo = &e
				got, err := ParseUpdateGprsLocation(strictBER(t, w))
				if err != nil {
					t.Fatal(err)
				}
				wantEqual(t, "EpsInfo", &EpsInfo{IsrInformation: &IsrInformation{
					UpdateLocation: raw&0x80 != 0, CancelSGSN: raw&0x40 != 0, InitialAttachIndicator: raw&0x20 != 0,
				}}, got.EpsInfo)

			})
		}
	}
}

// TS 29.002 V19.1.0 §17.7.1 Ext-AccessRestrictionData permits 1..32
// bits without a receiver rule discarding its unnamed positions. The
// decoder removes trailing 0 bits of the named BIT STRING (ITU-T X.680
// (02/2021) §22.7), so Parse reports no BitLength and Marshal sends the
// value without them.
func TestReceiverExtAccessRestrictionUnknownBits(t *testing.T) {
	for _, tc := range []struct {
		raw     []byte
		bits    int
		unknown HexBytes
		sent    runtime.BitString
	}{
		{[]byte{0x20}, 3, HexBytes{0x20}, runtime.BitString{Bytes: []byte{0x20}, BitLength: 3}},
		{[]byte{0, 0, 0, 1}, 32, HexBytes{0, 0, 0, 1}, runtime.BitString{Bytes: []byte{0, 0, 0, 1}, BitLength: 32}},
		{[]byte{0xE0}, 3, HexBytes{0x20}, runtime.BitString{Bytes: []byte{0xE0}, BitLength: 3}},
		{[]byte{0xC0, 0, 0, 0}, 32, nil, runtime.BitString{Bytes: []byte{0xC0}, BitLength: 2}},
		{[]byte{0}, 1, nil, runtime.BitString{Bytes: []byte{0}, BitLength: 1}},
	} {
		t.Run(fmt.Sprintf("%x/%d", tc.raw, tc.bits), func(t *testing.T) {
			bs := runtime.BitString{Bytes: tc.raw, BitLength: tc.bits}
			got, err := ParseInsertSubscriberDataArg(strictBER(t, &gsm_map.InsertSubscriberDataArg{ExtAccessRestrictionData: &bs}))
			if err != nil {
				t.Fatal(err)
			}
			r := got.ExtAccessRestrictionData
			if r == nil {
				t.Fatal("missing restrictions")
			}
			if r.NrAsSecondaryRATNotAllowed != (tc.raw[0]&0x80 != 0) || r.UnlicensedSpectrumAsSecondaryRATNotAllowed != (tc.raw[0]&0x40 != 0) || !bytes.Equal(r.UnknownBits, tc.unknown) {
				t.Fatalf("restrictions = %+v", r)
			}
			data, err := got.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			var w gsm_map.InsertSubscriberDataArg
			if err = w.UnmarshalBER(data); err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "wire restrictions", &tc.sent, w.ExtAccessRestrictionData)
		})
	}
}

// TS 29.002 V19.1.0 §17.7.1 treats unknown ODB-GeneralData assignments like
// unsupported ODB-GeneralData, and §8.8.1.3 has the receiver return unsupported
// categories in the ISD result. Parse and Marshal carry them.
func TestReceiverODBGeneralUnknownBits(t *testing.T) {
	// General Data is [8] in the argument and [4] in the result. A four
	// octet BIT STRING has length 5 including its unused-bit count; its
	// argument wrapper has length 7 and the outer SEQUENCE length is 9.
	for _, tc := range []struct {
		name             string
		arg, res         string
		sentArg, sentRes string
		want             ODBGeneralData
	}{
		{
			name: "named-and-bit30",
			arg:  "3009a80703050080000002", res: "300784050080000002",
			sentArg: "3009a80703050180000002", sentRes: "300784050180000002",
			want: ODBGeneralData{AllOGCallsBarred: true, UnknownBits: HexBytes{0, 0, 0, 2}},
		},
		{
			name: "only-first-unknown",
			arg:  "3009a80703050000000004", res: "300784050000000004",
			sentArg: "3009a80703050200000004", sentRes: "300784050200000004",
			want: ODBGeneralData{UnknownBits: HexBytes{0, 0, 0, 4}},
		},
		{
			name: "only-last-unknown",
			arg:  "3009a80703050000000001", res: "300784050000000001",
			sentArg: "3009a80703050000000001", sentRes: "300784050000000001",
			want: ODBGeneralData{UnknownBits: HexBytes{0, 0, 0, 1}},
		},
		{
			name: "mask-named-positions",
			arg:  "3009a8070305008080808d", res: "30078405008080808d",
			sentArg: "3009a8070305008080808d", sentRes: "30078405008080808d",
			want: ODBGeneralData{
				AllOGCallsBarred: true,
				InterzonalOGCallsAndInternationalOGCallsNotToHPLMNCountryBarred: true,
				RoamerAccessToHPLMNAPBarred:                                     true,
				RegistrationAllCFBarred:                                         true,
				RegistrationInternationalCFBarred:                               true,
				UnknownBits:                                                     HexBytes{0, 0, 0, 5},
			},
		},
		{
			name: "named-only-trailing-zero-octets",
			arg:  "3009a80703050080000000", res: "300784050080000000",
			sentArg: "3007a8050303018000", sentRes: "30058403018000",
			want: ODBGeneralData{AllOGCallsBarred: true},
		},
		{
			name: "last-named-only",
			arg:  "3009a80703050000000008", res: "300784050000000008",
			sentArg: "3009a80703050300000008", sentRes: "300784050300000008",
			want: ODBGeneralData{RegistrationInternationalCFBarred: true},
		},
		{
			name: "padding-bits-are-not-unknown-bits",
			arg:  "3009a80703050300000007", res: "300784050300000007",
			sentArg: "3007a8050303010000", sentRes: "30058403010000",
			want: ODBGeneralData{},
		},
		{
			name: "shortest-size",
			arg:  "3007a8050303018000", res: "30058403018000",
			sentArg: "3007a8050303018000", sentRes: "30058403018000",
			want: ODBGeneralData{AllOGCallsBarred: true},
		},
		{
			name: "zero-bits-beyond-size",
			arg:  "300aa8080306008000000000", res: "30088406008000000000",
			sentArg: "3007a8050303018000", sentRes: "30058403018000",
			want: ODBGeneralData{AllOGCallsBarred: true},
		},
		{
			name: "none",
			arg:  "3009a80703050000000000", res: "300784050000000000",
			sentArg: "3007a8050303010000", sentRes: "30058403010000",
			want: ODBGeneralData{},
		},
	} {
		for _, result := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/result=%t", tc.name, result), func(t *testing.T) {
				if result {
					got, err := ParseInsertSubscriberDataRes(mustHex(t, tc.res))
					if err != nil {
						t.Fatal(err)
					}
					wantEqual(t, "ODBGeneralData", &tc.want, got.OdbGeneralData)
					data, err := got.Marshal()
					if err != nil {
						t.Fatal(err)
					}
					wantEqual(t, "wire", mustHex(t, tc.sentRes), data)
					again, err := ParseInsertSubscriberDataRes(data)
					if err != nil {
						t.Fatal(err)
					}
					wantEqual(t, "reparsed ODBGeneralData", &tc.want, again.OdbGeneralData)
				} else {
					got, err := ParseInsertSubscriberDataArg(mustHex(t, tc.arg))
					if err != nil {
						t.Fatal(err)
					}
					if got.OdbData == nil {
						t.Fatal("missing ODBData")
					}
					wantEqual(t, "ODBGeneralData", &tc.want, got.OdbData.OdbGeneralData)
					data, err := got.Marshal()
					if err != nil {
						t.Fatal(err)
					}
					wantEqual(t, "wire", mustHex(t, tc.sentArg), data)
					again, err := ParseInsertSubscriberDataArg(data)
					if err != nil {
						t.Fatal(err)
					}
					wantEqual(t, "reparsed ODBData", got.OdbData, again.OdbData)
					// A VLR caller can report the received categories in its response.
					echo, err := (&InsertSubscriberDataRes{OdbGeneralData: got.OdbData.OdbGeneralData}).Marshal()
					if err != nil {
						t.Fatal(err)
					}
					wantEqual(t, "echo wire", mustHex(t, tc.sentRes), echo)
				}
			})
		}
	}
}

// TS 29.002 V19.1.0 §17.7.1 gives ODB-HPLMN-Data the same exception rule. The
// ISD result has no HPLMN parameter, so only the argument carries it.
func TestReceiverODBHPLMNUnknownBits(t *testing.T) {
	// ODB-Data contains mandatory General Data (5 octets here), followed
	// by HPLMN Data (7 octets at SIZE 32): [8] length 12, SEQUENCE length 14.
	for _, tc := range []struct {
		name, wire, sent string
		want             ODBHPLMNData
	}{
		{"only-first-unknown", "300ea80c030301000003050008000000", "300ba809030301000003020308", ODBHPLMNData{UnknownBits: HexBytes{8}}},
		{"only-last-unknown", "300ea80c030301000003050000000001", "300ea80c030301000003050000000001", ODBHPLMNData{UnknownBits: HexBytes{0, 0, 0, 1}}},
		{"named-and-unknown", "300ea80c030301000003050098000001", "300ea80c030301000003050098000001", ODBHPLMNData{PLMNSpecificBarringType1: true, PLMNSpecificBarringType4: true, UnknownBits: HexBytes{8, 0, 0, 1}}},
		{"named-only-trailing-zero-octets", "300ea80c030301000003050090000000", "300ba809030301000003020490", ODBHPLMNData{PLMNSpecificBarringType1: true, PLMNSpecificBarringType4: true}},
		{"padding-bits-are-not-unknown-bits", "300ea80c03030100000305040000000f", "300ba809030301000003020400", ODBHPLMNData{}},
		{"zero-bits-beyond-size", "300fa80d03030100000306009000000000", "300ba809030301000003020490", ODBHPLMNData{PLMNSpecificBarringType1: true, PLMNSpecificBarringType4: true}},
		{"none", "300ea80c030301000003050000000000", "300ba809030301000003020400", ODBHPLMNData{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseInsertSubscriberDataArg(mustHex(t, tc.wire))
			if err != nil {
				t.Fatal(err)
			}
			if got.OdbData == nil {
				t.Fatal("missing ODBData")
			}
			wantEqual(t, "ODBHPLMNData", &tc.want, got.OdbData.OdbHPLMNData)
			wantEqual(t, "ODBGeneralData", &ODBGeneralData{}, got.OdbData.OdbGeneralData)
			data, err := got.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "wire", mustHex(t, tc.sent), data)
			again, err := ParseInsertSubscriberDataArg(data)
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "reparsed ODBData", got.OdbData, again.OdbData)
		})
	}
}

// A set bit beyond 31 violates the SIZE of both types (TS 29.002 V19.1.0
// §17.7.1), so these messages never reach UnknownBits.
func TestReceiverODBBitBeyondSize(t *testing.T) {
	for _, tc := range []struct {
		name, wire string
		parse      func([]byte) error
	}{
		{"General in the argument", "300aa8080306008000000080", func(b []byte) error { _, err := ParseInsertSubscriberDataArg(b); return err }},
		{"General in the result", "30088406008000000080", func(b []byte) error { _, err := ParseInsertSubscriberDataRes(b); return err }},
		{"HPLMN in the argument", "300fa80d03030100000306000000000080", func(b []byte) error { _, err := ParseInsertSubscriberDataArg(b); return err }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.parse(mustHex(t, tc.wire)); !errors.Is(err, ber.ErrInvalidValue) {
				t.Errorf("err = %v, want %v", err, ber.ErrInvalidValue)
			}
		})
	}
}

// Named fields override UnknownBits, including a false named field.
// TS 29.002 V19.1.0 §17.7.1 bounds both ODB types at 32 bits.
func TestSenderODBGeneralUnknownBits(t *testing.T) {
	for _, tc := range []struct {
		name     string
		value    ODBGeneralData
		arg, res string
	}{
		{"empty", ODBGeneralData{}, "3007a8050303010000", "30058403010000"},
		{"named-only", ODBGeneralData{AllOGCallsBarred: true, RegistrationInternationalCFBarred: true}, "3009a80703050380000008", "300784050380000008"},
		{"ignore-named-positions", ODBGeneralData{UnknownBits: HexBytes{0xff, 0xff, 0xff, 0xf8}}, "3007a8050303010000", "30058403010000"},
		{"named-fields-win", ODBGeneralData{AllOGCallsBarred: true, RegistrationInternationalCFBarred: true, UnknownBits: HexBytes{0x40, 0, 0, 7, 0}}, "3009a8070305008000000f", "30078405008000000f"},
		{"zero-octets-beyond-limit", ODBGeneralData{AllOGCallsBarred: true, UnknownBits: HexBytes{0, 0, 0, 0, 0}}, "3007a8050303018000", "30058403018000"},
	} {
		for _, result := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/result=%t", tc.name, result), func(t *testing.T) {
				var data []byte
				var err error
				wire := tc.arg
				if result {
					data, err = (&InsertSubscriberDataRes{OdbGeneralData: &tc.value}).Marshal()
					wire = tc.res
				} else {
					data, err = (&InsertSubscriberDataArg{OdbData: &ODBData{OdbGeneralData: &tc.value}}).Marshal()
				}
				if err != nil {
					t.Fatal(err)
				}
				wantEqual(t, "wire", mustHex(t, wire), data)
			})
		}
	}
	for _, result := range []bool{false, true} {
		t.Run(fmt.Sprintf("bit32/result=%t", result), func(t *testing.T) {
			value := &ODBGeneralData{UnknownBits: HexBytes{0, 0, 0, 0, 0x80}}
			var err error
			if result {
				_, err = (&InsertSubscriberDataRes{OdbGeneralData: value}).Marshal()
			} else {
				_, err = (&InsertSubscriberDataArg{OdbData: &ODBData{OdbGeneralData: value}}).Marshal()
			}
			wantValidationError(t, err, "ODB-GeneralData", "SIZE (15..32)", 33)
		})
	}
}

func TestSenderODBHPLMNUnknownBits(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value ODBHPLMNData
		wire  string
	}{
		{"empty", ODBHPLMNData{}, "300ba809030301000003020400"},
		{"named-only", ODBHPLMNData{PLMNSpecificBarringType1: true, PLMNSpecificBarringType4: true}, "300ba809030301000003020490"},
		{"ignore-named-positions", ODBHPLMNData{UnknownBits: HexBytes{0xf0}}, "300ba809030301000003020400"},
		{"named-fields-win", ODBHPLMNData{PLMNSpecificBarringType1: true, UnknownBits: HexBytes{0x78, 0, 0, 1, 0}}, "300ea80c030301000003050088000001"},
		{"zero-octets-beyond-limit", ODBHPLMNData{PLMNSpecificBarringType1: true, UnknownBits: HexBytes{0, 0, 0, 0, 0}}, "300ba809030301000003020480"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			data, err := (&InsertSubscriberDataArg{OdbData: &ODBData{OdbGeneralData: &ODBGeneralData{}, OdbHPLMNData: &tc.value}}).Marshal()
			if err != nil {
				t.Fatal(err)
			}
			wantEqual(t, "wire", mustHex(t, tc.wire), data)
		})
	}
	t.Run("bit32", func(t *testing.T) {
		_, err := (&InsertSubscriberDataArg{OdbData: &ODBData{
			OdbGeneralData: &ODBGeneralData{},
			OdbHPLMNData:   &ODBHPLMNData{UnknownBits: HexBytes{0, 0, 0, 0, 0x80}},
		}}).Marshal()
		wantValidationError(t, err, "ODB-HPLMN-Data", "SIZE (4..32)", 33)
	})
}

// TS 29.002 V19.1.0 §17.7.1 SupportedRAT-Types names bits 0 to 5, nb-iot
// being bit 5: "bits 6 to 7 shall be ignored if received and not
// understood". The decoder removes trailing 0 bits (ITU-T X.680 (02/2021)
// §22.7), down to the SIZE (2..8) lower bound.
func TestReceiverSupportedRATTypesNBIoT(t *testing.T) {
	for _, tc := range []struct {
		raw  byte
		bits int
		want SupportedRATTypes
		sent runtime.BitString
	}{
		{0x04, 6, SupportedRATTypes{NBIoT: true}, runtime.BitString{Bytes: []byte{0x04}, BitLength: 6}},
		{0x0C, 6, SupportedRATTypes{EUTRAN: true, NBIoT: true}, runtime.BitString{Bytes: []byte{0x0C}, BitLength: 6}},
		{0x07, 8, SupportedRATTypes{NBIoT: true}, runtime.BitString{Bytes: []byte{0x04}, BitLength: 6}},
		{0x83, 8, SupportedRATTypes{UTRAN: true}, runtime.BitString{Bytes: []byte{0x80}, BitLength: 2}},
	} {
		t.Run(fmt.Sprintf("%02x/%d", tc.raw, tc.bits), func(t *testing.T) {
			w := mobilityGprsWire(t)
			w.SgsnCapability = &gsm_map.SGSNCapability{SupportedRATTypesIndicator: &runtime.BitString{Bytes: []byte{tc.raw}, BitLength: tc.bits}}
			got, err := ParseUpdateGprsLocation(strictBER(t, w))
			if err != nil {
				t.Fatal(err)
			}
			if got.SGSNCapability == nil {
				t.Fatal("missing SGSNCapability")
			}
			wantEqual(t, "SupportedRATTypesIndicator", &tc.want, got.SGSNCapability.SupportedRATTypesIndicator)
			data, err := got.Marshal()
			if err != nil {
				t.Fatal(err)
			}
			var sent gsm_map.UpdateGprsLocationArg
			if err := sent.UnmarshalBER(data); err != nil {
				t.Fatal(err)
			}
			if sent.SgsnCapability == nil {
				t.Fatal("missing sgsn-Capability")
			}
			wantEqual(t, "sent supportedRAT-TypesIndicator", &tc.sent, sent.SgsnCapability.SupportedRATTypesIndicator)
		})
	}
}

// TS 29.002 V19.1.0 §17.7.1 PDP-Context QoS predecessor and extended
// address requirements restrict sending; independent received fields remain inspectable.
func TestReceiverPDPContextSenderDependencies(t *testing.T) {
	for _, tc := range []struct {
		name string
		mut  func(*gsm_map.PDPContext)
		want error
	}{
		{"Ext2 without Ext", func(w *gsm_map.PDPContext) { w.ExtQoSSubscribed = nil }, ErrExtQoSHierarchyViolated},
		{"Ext3 without Ext2", func(w *gsm_map.PDPContext) { w.Ext2QoSSubscribed = nil }, ErrExtQoSHierarchyViolated},
		{"Ext4 without Ext3", func(w *gsm_map.PDPContext) { w.Ext3QoSSubscribed = nil }, ErrExtQoSHierarchyViolated},
		{"extended IPv4 without base", func(w *gsm_map.PDPContext) {
			w.PdpAddress = nil
			v := gsm_map.PDPAddress{192, 0, 2, 1}
			w.ExtPdpAddress = &v
		}, ErrExtPDPAddressWithoutPDPAddress},
		{"extended IPv6 without base", func(w *gsm_map.PDPContext) {
			w.PdpAddress = nil
			v := make([]byte, 16)
			w.ExtPdpAddress = &v
		}, ErrExtPDPAddressWithoutPDPAddress},
		{"complete", func(*gsm_map.PDPContext) {}, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := makePDPContext()
			w, err := convertPDPContextToWire(&p)
			if err != nil {
				t.Fatal(err)
			}
			tc.mut(w)
			arg := &gsm_map.InsertSubscriberDataArg{GprsSubscriptionData: &gsm_map.GPRSSubscriptionData{GprsDataList: &gsm_map.GPRSDataList{Values: []gsm_map.PDPContext{*w}}}}
			got, err := ParseInsertSubscriberDataArg(strictBER(t, arg))
			if err != nil {
				t.Fatal(err)
			}
			r := got.GprsSubscriptionData.GprsDataList[0]
			want := func(v *[]byte) HexBytes {
				if v == nil {
					return nil
				}
				return HexBytes(*v)
			}
			wantEqual(t, "ExtQoS", want(w.ExtQoSSubscribed), r.ExtQoSSubscribed)
			wantEqual(t, "Ext2QoS", want(w.Ext2QoSSubscribed), r.Ext2QoSSubscribed)
			wantEqual(t, "Ext3QoS", want(w.Ext3QoSSubscribed), r.Ext3QoSSubscribed)
			wantEqual(t, "Ext4QoS", want(w.Ext4QoSSubscribed), r.Ext4QoSSubscribed)
			if w.PdpAddress == nil && r.PdpAddress != nil {
				t.Fatal("base address invented")
			}
			if w.ExtPdpAddress != nil {
				wantEqual(t, "extended address", HexBytes(*w.ExtPdpAddress), r.ExtPdpAddress)
			}
			_, err = got.Marshal()
			if !errors.Is(err, tc.want) {
				t.Fatalf("Marshal = %v, want %v", err, tc.want)
			}
		})
	}
}

// TS 29.002 V19.1.0 §17.7.1 DestinationNumberCriteria restricts list
// absence in sent components without prescribing receiver rejection.
func TestReceiverDestinationNumberCriteriaListAbsence(t *testing.T) {
	for _, where := range []string{"ISD", "SRI"} {
		for _, lists := range []int{0, 1, 2, 3} {
			t.Run(fmt.Sprintf("%s/%d", where, lists), func(t *testing.T) {
				d := &gsm_map.DestinationNumberCriteria{MatchType: MatchTypeInhibiting}
				if lists&1 != 0 {
					d.DestinationNumberList = &gsm_map.DestinationNumberList{Values: []gsm_map.ISDNAddressString{{0x91, 0x21}}}
				}
				if lists&2 != 0 {
					d.DestinationNumberLengthList = &gsm_map.DestinationNumberLengthList{Values: []int64{2}}
				}
				criteria := &gsm_map.OBcsmCamelTDPCriteriaList{Values: []gsm_map.OBcsmCamelTDPCriteria{{OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo, DestinationNumberCriteria: d}}}
				var r *DestinationNumberCriteria
				var senderErr error
				if where == "ISD" {
					w := &gsm_map.InsertSubscriberDataArg{VlrCamelSubscriptionInfo: &gsm_map.VlrCamelSubscriptionInfo{OBcsmCamelTDPCriteriaList: criteria}}
					got, err := ParseInsertSubscriberDataArg(strictBER(t, w))
					if err != nil {
						t.Fatal(err)
					}
					r = got.VlrCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList[0].DestinationNumberCriteria
					_, senderErr = got.Marshal()
				} else {
					w := &gsm_map.SendRoutingInfoRes{ExtendedRoutingInfo: new(gsm_map.NewExtendedRoutingInfoCamelRoutingInfo(gsm_map.CamelRoutingInfo{GmscCamelSubscriptionInfo: gsm_map.GmscCamelSubscriptionInfo{OBcsmCamelTDPCriteriaList: criteria}}))}
					got, err := ParseSriRes(strictBER(t, w))
					if err != nil {
						t.Fatal(err)
					}
					r = got.ExtendedRoutingInfo.CamelRoutingInfo.GmscCamelSubscriptionInfo.OBcsmCamelTDPCriteriaList[0].DestinationNumberCriteria
					_, senderErr = got.Marshal()
				}
				if r == nil || r.MatchType != MatchTypeInhibiting || (r.DestinationNumberList != nil) != (lists&1 != 0) || (r.DestinationNumberLengthList != nil) != (lists&2 != 0) {
					t.Fatalf("criteria changed: %+v", r)
				}
				if lists == 0 {
					if !errors.Is(senderErr, ErrCamelMissingDestinationNumberCriteria) {
						t.Fatalf("Marshal: %v", senderErr)
					}
				} else if senderErr != nil {
					t.Fatal(senderErr)
				}
			})
		}
	}
}

// Repeated triggers identify separate received descriptors; each descriptor's
// service key and position remain available after Parse (TS 29.002 §17.7.1).
func TestReceiverRepeatedTDPDescriptors(t *testing.T) {
	for _, family := range []string{"O", "T", "SMS", "GPRS"} {
		t.Run(family, func(t *testing.T) {
			w := isdWire(t, camelISD())
			switch family {
			case "O":
				l := w.VlrCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList
				l.Values = wireTDPList(l.Values[0], []int64{2, 2}, setOWire)
			case "T":
				l := w.VlrCamelSubscriptionInfo.VtCSI.TBcsmCamelTDPDataList
				l.Values = wireTDPList(l.Values[0], []int64{12, 12}, setTWire)
			case "SMS":
				l := w.VlrCamelSubscriptionInfo.MoSmsCSI.SmsCAMELTDPDataList
				l.Values = wireTDPList(l.Values[0], []int64{1, 1}, setSMSWire)
			case "GPRS":
				l := w.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList
				l.Values = wireTDPList(l.Values[0], []int64{1, 1}, setGPRSWire)
			}
			got, err := ParseInsertSubscriberDataArg(strictBER(t, w))
			if err != nil {
				t.Fatal(err)
			}
			var keys []int64
			switch family {
			case "O":
				for _, e := range got.VlrCamelSubscriptionInfo.OCSI.OBcsmCamelTDPDataList {
					keys = append(keys, e.ServiceKey)
				}
			case "T":
				for _, e := range got.VlrCamelSubscriptionInfo.VtCSI.TBcsmCamelTDPDataList {
					keys = append(keys, e.ServiceKey)
				}
			case "SMS":
				for _, e := range got.VlrCamelSubscriptionInfo.MoSmsCSI.SmsCAMELTDPDataList {
					keys = append(keys, e.ServiceKey)
				}
			case "GPRS":
				for _, e := range got.SgsnCAMELSubscriptionInfo.GprsCSI.GprsCamelTDPDataList {
					keys = append(keys, e.ServiceKey)
				}
			}
			wantEqual(t, "descriptor service keys in received order", []int64{1, 2}, keys)
			if _, err := got.Marshal(); !errors.Is(err, ErrCamelDuplicateTriggerDetectionPoint) {
				t.Fatalf("Marshal: %v", err)
			}
		})
	}
}

// A present empty list violates ASN SIZE, unlike an absent list; these
// constraints still belong to the codec (TS 29.002 V19.1.0 §17.7.1).
func TestReceiverDestinationNumberCriteriaEmptyPresentLists(t *testing.T) {
	for _, where := range []string{"ISD", "SRI"} {
		for _, numbers := range []bool{true, false} {
			t.Run(fmt.Sprintf("%s/numbers=%t", where, numbers), func(t *testing.T) {
				d := &gsm_map.DestinationNumberCriteria{MatchType: MatchTypeInhibiting}
				if numbers {
					d.DestinationNumberList = &gsm_map.DestinationNumberList{}
				} else {
					d.DestinationNumberLengthList = &gsm_map.DestinationNumberLengthList{}
				}
				l := &gsm_map.OBcsmCamelTDPCriteriaList{Values: []gsm_map.OBcsmCamelTDPCriteria{{OBcsmTriggerDetectionPoint: OBcsmTriggerCollectedInfo, DestinationNumberCriteria: d}}}
				var err error
				if where == "ISD" {
					_, err = ParseInsertSubscriberDataArg(tolerantBER(t, &gsm_map.InsertSubscriberDataArg{VlrCamelSubscriptionInfo: &gsm_map.VlrCamelSubscriptionInfo{OBcsmCamelTDPCriteriaList: l}}))
				} else {
					_, err = ParseSriRes(tolerantBER(t, &gsm_map.SendRoutingInfoRes{ExtendedRoutingInfo: new(gsm_map.NewExtendedRoutingInfoCamelRoutingInfo(gsm_map.CamelRoutingInfo{GmscCamelSubscriptionInfo: gsm_map.GmscCamelSubscriptionInfo{OBcsmCamelTDPCriteriaList: l}}))}))
				}
				var constraint *ber.ValidationError
				if !errors.As(err, &constraint) {
					t.Fatalf("Parse: %v, want codec SIZE error", err)
				}
			})
		}
	}
}
