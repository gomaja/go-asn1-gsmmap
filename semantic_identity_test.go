// semantic_identity_test.go
//
// Digit counts of IMSI, IMEI and IMEISV. 3GPP TS 23.003 V20.1.0 fixes the
// composition of each identity (§2.2, §2.3, §6.2.1, §6.2.2); TS 29.002
// V19.1.0 §17.7.8 carries them as TBCD-STRING (IMSI SIZE (3..8), IMEI
// SIZE (8)), so the octet SIZE the codec checks admits digit counts the
// identities do not have. One rule per identity applies to every operation,
// on Marshal and on Parse alike.
package gsmmap

import (
	"encoding/hex"
	"errors"
	"testing"

	"github.com/gomaja/go-asn1/runtime/ber"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"

	"github.com/gomaja/go-asn1-gsmmap/address"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// semDigits returns n digits. The 15th is 0, the spare digit a 15-digit
// IMEI ends in (semantic_imei_test.go).
func semDigits(n int) string {
	return "12345678901234067890"[:n]
}

func semTBCD(t *testing.T, digits string) []byte {
	t.Helper()
	b, err := tbcd.Encode(digits)
	if err != nil {
		t.Fatalf("tbcd.Encode(%q): %v", digits, err)
	}
	return b
}

// semMtFsmHex is an MT-ForwardSM-Arg whose SM-RP-DA is an IMSI.
const semMtFsmHex = "3077800832140080803138f684069169318488880463040b916971101174f40000422182612464805bd2e2b1252d467ff6de6c47efd96eb6a1d056cb0d69b49a10269c098537586e96931965b260d15613da72c29b91261bde72c6a1ad2623d682b5996d58331271375a0d1733eee4bd98ec768bd966b41c0d"

type semIdentityCase struct {
	name string
	// marshal builds a valid message carrying id in the field under test and
	// marshals it.
	marshal func(t *testing.T, id string) ([]byte, error)
	// parse parses data and returns the field under test.
	parse func(data []byte) (string, error)
}

func semIMSICases() []semIdentityCase {
	const msisdn, sca = "31612345678", "31600000001"
	return []semIdentityCase{
		{"SendAuthenticationInfo",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&SendAuthenticationInfo{IMSI: id, NumberOfRequestedVectors: 1}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseSendAuthenticationInfo(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"UpdateLocation",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&UpdateLocation{IMSI: id, MSCNumber: msisdn, VLRNumber: msisdn}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseUpdateLocation(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"UpdateGprsLocation",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&UpdateGprsLocation{IMSI: id, SGSNNumber: msisdn, SGSNAddress: "192.168.31.1"}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseUpdateGprsLocation(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"PurgeMS",
			func(_ *testing.T, id string) ([]byte, error) { return (&PurgeMS{IMSI: id}).Marshal() },
			func(d []byte) (string, error) {
				v, err := ParsePurgeMS(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"ProvideSubscriberInfo",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&ProvideSubscriberInfo{IMSI: id, RequestedInfo: RequestedInfo{LocationInformation: true}}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseProvideSubscriberInfo(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"CancelLocation imsi",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&CancelLocation{Identity: CancelLocationIdentity{IMSI: id}}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseCancelLocation(d)
				return semField(v, err, func() string { return v.Identity.IMSI })
			}},
		{"CancelLocation imsi-WithLMSI",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&CancelLocation{Identity: CancelLocationIdentity{IMSIWithLMSI: &CancelLocationIMSIWithLMSI{IMSI: id, LMSI: HexBytes{1, 2, 3, 4}}}}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseCancelLocation(d)
				return semField(v, err, func() string { return v.Identity.IMSIWithLMSI.IMSI })
			}},
		{"AlertServiceCentre",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&AlertServiceCentre{MSISDN: msisdn, ServiceCentreAddress: sca, IMSI: id}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseAlertServiceCentre(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"ReportSMDeliveryStatus",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&ReportSMDeliveryStatus{MSISDN: msisdn, ServiceCentreAddress: sca, IMSI: id}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseReportSMDeliveryStatus(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"SriSm",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&SriSm{MSISDN: msisdn, ServiceCentreAddress: sca, SmRpPri: true, IMSI: id}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseSriSm(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"SriSmResp",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&SriSmResp{IMSI: id, LocationInfoWithLMSI: LocationInfoWithLMSI{NetworkNodeNumber: msisdn}}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseSriSmResp(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"SriResp",
			func(_ *testing.T, id string) ([]byte, error) { return (&SriResp{IMSI: id}).Marshal() },
			func(d []byte) (string, error) {
				v, err := ParseSriResp(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"MtFsm sm-RP-DA",
			func(t *testing.T, id string) ([]byte, error) {
				m := semMtFsm(t)
				m.SmRpDa = SmRpDa{IMSI: id}
				return m.Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseMtFsm(d)
				return semField(v, err, func() string { return v.SmRpDa.IMSI })
			}},
		{"ProvideSubscriberLocation",
			func(_ *testing.T, id string) ([]byte, error) {
				a := semPSLArg()
				a.IMSI = id
				return a.Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseProvideSubscriberLocation(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"SubscriberLocationReport",
			func(_ *testing.T, id string) ([]byte, error) {
				a := semSLRArg()
				a.IMSI = id
				return a.Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseSubscriberLocationReport(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"AnyTimeInterrogation subscriberIdentity",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&AnyTimeInterrogation{
					SubscriberIdentity: SubscriberIdentity{IMSI: id},
					RequestedInfo:      RequestedInfo{SubscriberState: true},
					GsmSCFAddress:      msisdn,
				}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseAnyTimeInterrogation(d)
				return semField(v, err, func() string { return v.SubscriberIdentity.IMSI })
			}},
		{"AnyTimeInterrogationRes mnpInfoRes",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{MnpInfoRes: &MnpInfoRes{IMSI: id}}}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseAnyTimeInterrogationRes(d)
				return semField(v, err, func() string { return v.SubscriberInfo.MnpInfoRes.IMSI })
			}},
	}
}

func semIMEICases() []semIdentityCase {
	return []semIdentityCase{
		{"AnyTimeInterrogationRes imei",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&AnyTimeInterrogationRes{SubscriberInfo: SubscriberInfo{IMEI: id}}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseAnyTimeInterrogationRes(d)
				return semField(v, err, func() string { return v.SubscriberInfo.IMEI })
			}},
		{"ProvideSubscriberInfoRes imei",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&ProvideSubscriberInfoRes{SubscriberInfo: SubscriberInfo{IMEI: id}}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseProvideSubscriberInfoRes(d)
				return semField(v, err, func() string { return v.SubscriberInfo.IMEI })
			}},
		{"ProvideSubscriberLocation imei",
			func(_ *testing.T, id string) ([]byte, error) {
				a := semPSLArg()
				a.IMEI = id
				return a.Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseProvideSubscriberLocation(d)
				return semField(v, err, func() string { return v.IMEI })
			}},
		{"SubscriberLocationReport imei",
			func(_ *testing.T, id string) ([]byte, error) {
				a := semSLRArg()
				a.IMEI = id
				return a.Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseSubscriberLocationReport(d)
				return semField(v, err, func() string { return v.IMEI })
			}},
	}
}

func semIMEISVCases() []semIdentityCase {
	return []semIdentityCase{
		{"UpdateLocation add-Info",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&UpdateLocation{IMSI: "001010123456789", MSCNumber: "31612345678", VLRNumber: "31612345678", AddInfo: &AddInfo{IMEISV: id}}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseUpdateLocation(d)
				return semField(v, err, func() string { return v.AddInfo.IMEISV })
			}},
		{"UpdateGprsLocation add-Info",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&UpdateGprsLocation{IMSI: "001010123456789", SGSNNumber: "31612345678", SGSNAddress: "192.168.31.1", AddInfo: &AddInfo{IMEISV: id}}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseUpdateGprsLocation(d)
				return semField(v, err, func() string { return v.AddInfo.IMEISV })
			}},
	}
}

func semField[T any](v *T, err error, get func() string) (string, error) {
	if err != nil {
		return "", err
	}
	return get(), nil
}

func semMtFsm(t *testing.T) *MtFsm {
	t.Helper()
	b, err := hex.DecodeString(semMtFsmHex)
	if err != nil {
		t.Fatalf("hex: %v", err)
	}
	m, err := ParseMtFsm(b)
	if err != nil {
		t.Fatalf("ParseMtFsm: %v", err)
	}
	return m
}

func semPSLArg() *ProvideSubscriberLocationArg {
	return &ProvideSubscriberLocationArg{
		LocationType:    LocationType{LocationEstimateType: LocationEstimateCurrentLocation},
		MlcNumber:       "31612345678",
		MlcNumberNature: address.NatureInternational,
		MlcNumberPlan:   address.PlanISDN,
	}
}

func semSLRArg() *SubscriberLocationReportArg {
	return &SubscriberLocationReportArg{
		LcsEvent:        LCSEventEmergencyCallOrigination,
		LcsClientID:     LCSClientID{LcsClientType: LCSClientTypeEmergencyServices},
		LcsLocationInfo: LCSLocationInfo{NetworkNodeNumber: "31650000000"},
	}
}

// 3GPP TS 23.003 V20.1.0 §2.2: an IMSI is MCC (three digits), MNC (two or
// three digits) and MSIN, "Not more than 15 digits" (figure 1); §2.3: "The
// number of digits in IMSI shall not exceed 15." With a one-digit MSIN the
// shortest IMSI has 6 digits.
//
// §6.2.1: the IMEI is TAC (8 digits), SNR (6 digits) and CD/SD (1 digit), 15
// digits. §6.2.2: the IMEISV is TAC, SNR and SVN (2 digits), 16 digits. TS
// 29.002 V19.1.0 §17.7.8 IMEI "Refers to International Mobile Station
// Equipment Identity and Software Version Number (SVN)": an IMEI field holds
// the IMEI, or the IMEISV when the SVN is present.
func TestMarshalIdentityDigitCount(t *testing.T) {
	for _, kind := range []struct {
		name  string
		cases []semIdentityCase
		ok    []int
		bad   []int
		want  error
	}{
		{"IMSI", semIMSICases(), []int{6, 15}, []int{5, 16}, ErrIMSIInvalidLength},
		{"IMEI", semIMEICases(), []int{15, 16}, []int{14, 17}, ErrIMEIInvalidLength},
		{"IMEISV", semIMEISVCases(), []int{16}, []int{15, 17}, ErrIMEISVInvalidLength},
	} {
		for _, c := range kind.cases {
			for _, n := range kind.ok {
				id := semDigits(n)
				t.Run(kind.name+"/"+c.name+"/"+id, func(t *testing.T) {
					data, err := c.marshal(t, id)
					if err != nil {
						t.Fatalf("Marshal: %v", err)
					}
					got, err := c.parse(data)
					if err != nil {
						t.Fatalf("Parse: %v", err)
					}
					if got != id {
						t.Errorf("parsed %s = %q, want %q", kind.name, got, id)
					}
				})
			}
			for _, n := range kind.bad {
				id := semDigits(n)
				t.Run(kind.name+"/"+c.name+"/"+id, func(t *testing.T) {
					if _, err := c.marshal(t, id); !errors.Is(err, kind.want) {
						t.Errorf("Marshal: err = %v, want %v", err, kind.want)
					}
				})
			}
		}
	}
}

// The IMSI 00 f1 ff fits IMSI SIZE (3..8) but holds the three digits "001".
func TestParseSendAuthenticationInfoShortIMSI(t *testing.T) {
	data, _ := hex.DecodeString("3008800300f1ff020101")
	if _, err := ParseSendAuthenticationInfo(data); !errors.Is(err, ErrIMSIInvalidLength) {
		t.Errorf("ParseSendAuthenticationInfo: err = %v, want ErrIMSIInvalidLength", err)
	}
}

// Parse applies the same rules as Marshal. The octet strings below fit the
// codec's SIZE; the identity rule rejects their digit count.
func TestParseIdentityDigitCount(t *testing.T) {
	const scf = "31612345678"
	isdn := func(t *testing.T) []byte {
		b, err := encodeAddressField(scf, address.NatureInternational, address.PlanISDN)
		if err != nil {
			t.Fatalf("encodeAddressField: %v", err)
		}
		return b
	}
	type wire interface {
		MarshalBER(...ber.EncodeOption) ([]byte, error)
	}
	imsiCases := []struct {
		name  string
		build func(t *testing.T, imsi gsm_map.IMSI) wire
		parse func([]byte) error
	}{
		{"SendAuthenticationInfo", func(_ *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.SendAuthenticationInfoArg{Imsi: imsi, NumberOfRequestedVectors: 1}
		}, func(d []byte) error { _, err := ParseSendAuthenticationInfo(d); return err }},
		{"UpdateLocation", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.UpdateLocationArg{Imsi: imsi, MscNumber: isdn(t), VlrNumber: isdn(t)}
		}, func(d []byte) error { _, err := ParseUpdateLocation(d); return err }},
		{"PurgeMS", func(_ *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.PurgeMSArg{Imsi: imsi}
		}, func(d []byte) error { _, err := ParsePurgeMS(d); return err }},
		{"ReportSMDeliveryStatus", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.ReportSMDeliveryStatusArg{Msisdn: isdn(t), ServiceCentreAddress: isdn(t), Imsi: &imsi}
		}, func(d []byte) error { _, err := ParseReportSMDeliveryStatus(d); return err }},
		{"SriSmResp", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.RoutingInfoForSMRes{Imsi: imsi, LocationInfoWithLMSI: gsm_map.LocationInfoWithLMSI{NetworkNodeNumber: isdn(t)}}
		}, func(d []byte) error { _, err := ParseSriSmResp(d); return err }},
		{"AnyTimeInterrogation", func(t *testing.T, imsi gsm_map.IMSI) wire {
			v := struct{}{}
			return &gsm_map.AnyTimeInterrogationArg{
				SubscriberIdentity: gsm_map.NewSubscriberIdentityImsi(imsi),
				RequestedInfo:      gsm_map.RequestedInfo{SubscriberState: &v},
				GsmSCFAddress:      isdn(t),
			}
		}, func(d []byte) error { _, err := ParseAnyTimeInterrogation(d); return err }},
		{"AnyTimeInterrogationRes mnpInfoRes", func(_ *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{MnpInfoRes: &gsm_map.MNPInfoRes{Imsi: &imsi}}}
		}, func(d []byte) error { _, err := ParseAnyTimeInterrogationRes(d); return err }},
		{"ProvideSubscriberLocation", func(t *testing.T, imsi gsm_map.IMSI) wire {
			w, err := convertProvideSubscriberLocationArgToWire(semPSLArg())
			if err != nil {
				t.Fatalf("convertProvideSubscriberLocationArgToWire: %v", err)
			}
			w.Imsi = &imsi
			return w
		}, func(d []byte) error { _, err := ParseProvideSubscriberLocation(d); return err }},
		{"SubscriberLocationReport", func(t *testing.T, imsi gsm_map.IMSI) wire {
			w, err := convertSubscriberLocationReportArgToWire(semSLRArg())
			if err != nil {
				t.Fatalf("convertSubscriberLocationReportArgToWire: %v", err)
			}
			w.Imsi = &imsi
			return w
		}, func(d []byte) error { _, err := ParseSubscriberLocationReport(d); return err }},
		{"absentSubscriberSM imsi", func(_ *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.AbsentSubscriberSMParam{Imsi: &imsi}
		}, func(d []byte) error { _, err := ParseReturnErrorParameter(MapErrorAbsentSubscriberSM, d); return err }},
		{"absentSubscriberSM userIdentifierAlert", func(_ *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.AbsentSubscriberSMParam{UserIdentifierAlert: &imsi}
		}, func(d []byte) error { _, err := ParseReturnErrorParameter(MapErrorAbsentSubscriberSM, d); return err }},
	}
	for _, c := range imsiCases {
		for _, tc := range []struct {
			digits string
			want   error
		}{
			{semDigits(5), ErrIMSIInvalidLength},
			{semDigits(6), nil},
			{semDigits(15), nil},
			{semDigits(16), ErrIMSIInvalidLength},
		} {
			t.Run("IMSI/"+c.name+"/"+tc.digits, func(t *testing.T) {
				data, err := c.build(t, semTBCD(t, tc.digits)).MarshalBER()
				if err != nil {
					t.Fatalf("MarshalBER: %v", err)
				}
				if err := c.parse(data); !errors.Is(err, tc.want) {
					t.Errorf("Parse: err = %v, want %v", err, tc.want)
				}
			})
		}
	}

	// IMEI SIZE (8): 14 digits leave a filler octet; 16 digits fill it.
	imeiCases := []struct {
		name  string
		build func(t *testing.T, imei gsm_map.IMEI) wire
		parse func([]byte) error
		want  map[int]error
	}{
		{"AnyTimeInterrogationRes imei", func(_ *testing.T, imei gsm_map.IMEI) wire {
			return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{Imei: &imei}}
		}, func(d []byte) error { _, err := ParseAnyTimeInterrogationRes(d); return err },
			map[int]error{14: ErrIMEIInvalidLength, 15: nil, 16: nil}},
		{"ProvideSubscriberLocation imei", func(t *testing.T, imei gsm_map.IMEI) wire {
			w, err := convertProvideSubscriberLocationArgToWire(semPSLArg())
			if err != nil {
				t.Fatalf("convertProvideSubscriberLocationArgToWire: %v", err)
			}
			w.Imei = &imei
			return w
		}, func(d []byte) error { _, err := ParseProvideSubscriberLocation(d); return err },
			map[int]error{14: ErrIMEIInvalidLength, 15: nil, 16: nil}},
		{"SubscriberLocationReport imei", func(t *testing.T, imei gsm_map.IMEI) wire {
			w, err := convertSubscriberLocationReportArgToWire(semSLRArg())
			if err != nil {
				t.Fatalf("convertSubscriberLocationReportArgToWire: %v", err)
			}
			w.Imei = &imei
			return w
		}, func(d []byte) error { _, err := ParseSubscriberLocationReport(d); return err },
			map[int]error{14: ErrIMEIInvalidLength, 15: nil, 16: nil}},
		{"UpdateLocation add-Info imeisv", func(t *testing.T, imei gsm_map.IMEI) wire {
			return &gsm_map.UpdateLocationArg{
				Imsi: semTBCD(t, "001010123456789"), MscNumber: isdn(t), VlrNumber: isdn(t),
				AddInfo: &gsm_map.ADDInfo{Imeisv: imei},
			}
		}, func(d []byte) error { _, err := ParseUpdateLocation(d); return err },
			map[int]error{14: ErrIMEISVInvalidLength, 15: ErrIMEISVInvalidLength, 16: nil}},
	}
	for _, c := range imeiCases {
		for _, n := range []int{14, 15, 16} {
			digits := semDigits(n)
			t.Run("IMEI/"+c.name+"/"+digits, func(t *testing.T) {
				raw := semTBCD(t, digits)
				for len(raw) < 8 {
					raw = append(raw, 0xFF)
				}
				data, err := c.build(t, raw).MarshalBER()
				if err != nil {
					t.Fatalf("MarshalBER: %v", err)
				}
				if err := c.parse(data); !errors.Is(err, c.want[n]) {
					t.Errorf("Parse: err = %v, want %v", err, c.want[n])
				}
			})
		}
	}
}
