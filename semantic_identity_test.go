// semantic_identity_test.go
//
// Digit counts of IMSI, IMEI and IMEISV. 3GPP TS 23.003 V20.1.0 fixes the
// composition of each identity (§2.2, §2.3, §6.2.1, §6.2.2); TS 29.002
// V19.1.0 §17.7.8 carries them as TBCD-STRING (IMSI SIZE (3..8), IMEI
// SIZE (8)), so the octet SIZE the codec checks admits digit counts the
// identities do not have. Shared identity rules apply to every operation;
// Parse also accepts five-digit MCC+MNC values and any 15th IMEI digit.
package gsmmap

import (
	"encoding/hex"
	"errors"
	"testing"

	identity "github.com/gomaja/go-3gpp-identity"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
	"github.com/gomaja/go-tbcd"

	"github.com/gomaja/go-asn1-gsmmap/address"
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
				return (&UpdateLocation{IMSI: id, MscNumber: msisdn, VlrNumber: msisdn}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseUpdateLocation(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"UpdateGprsLocation",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&UpdateGprsLocation{IMSI: id, SgsnNumber: msisdn, SGSNAddress: "192.168.31.1"}).Marshal()
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
		{"SriSm HLR-Id",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&SriSm{
					MSISDN: msisdn, ServiceCentreAddress: sca, SmRpPri: true,
					CorrelationID: &SriSmCorrelationID{HlrID: id, SipUriB: HexBytes("sip:b@example.com")},
				}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseSriSm(d)
				return semField(v, err, func() string { return v.CorrelationID.HlrID })
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
		{"MoFsm imsi",
			func(t *testing.T, id string) ([]byte, error) {
				m := knownMoFsm(t)
				m.IMSI = id
				return m.Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseMoFsm(d)
				return semField(v, err, func() string { return v.IMSI })
			}},
		{"InsertSubscriberData",
			func(_ *testing.T, id string) ([]byte, error) { return (&InsertSubscriberDataArg{IMSI: id}).Marshal() },
			func(d []byte) (string, error) {
				v, err := ParseInsertSubscriberData(d)
				return semField(v, err, func() string { return v.IMSI })
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
				return (&UpdateLocation{IMSI: "001010123456789", MscNumber: "31612345678", VlrNumber: "31612345678", AddInfo: &AddInfo{IMEISV: id}}).Marshal()
			},
			func(d []byte) (string, error) {
				v, err := ParseUpdateLocation(d)
				return semField(v, err, func() string { return v.AddInfo.IMEISV })
			}},
		{"UpdateGprsLocation add-Info",
			func(_ *testing.T, id string) ([]byte, error) {
				return (&UpdateGprsLocation{IMSI: "001010123456789", SgsnNumber: "31612345678", SGSNAddress: "192.168.31.1", AddInfo: &AddInfo{IMEISV: id}}).Marshal()
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
		IMSI:            "001010123456789",
		LocationType:    LocationType{LocationEstimateType: LocationEstimateCurrentLocation},
		MlcNumber:       "31612345678",
		MlcNumberNature: address.NatureInternational,
		MlcNumberPlan:   address.PlanISDN,
	}
}

func semSLRArg() *SubscriberLocationReportArg {
	return &SubscriberLocationReportArg{
		IMSI:            "001010123456789",
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
		{"IMSI", semIMSICases(), []int{6, 15}, []int{5, 16}, identity.ErrIMSILength},
		{"IMEI", semIMEICases(), []int{15, 16}, []int{14, 17}, identity.ErrIMEILength},
		{"IMEISV", semIMEISVCases(), []int{16}, []int{15, 17}, identity.ErrIMEISVLength},
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
	if _, err := ParseSendAuthenticationInfo(data); !errors.Is(err, identity.ErrIMSILength) {
		t.Errorf("ParseSendAuthenticationInfo: err = %v, want identity.ErrIMSILength", err)
	}
}

// Parse accepts five-digit MCC+MNC values from peers omitting the dummy MSIN
// (TS 29.002 V19.1.0 §12.1.3; TS 23.040 V20.0.0 §8.1.4), preserving the
// digits. Other IMSI counts follow TS 23.003 V20.1.0 §§2.2, 2.3. All raw
// identities below fit the codec's octet SIZE, including rejected counts.
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
		parse func([]byte) (string, error)
	}{
		{"SendAuthenticationInfo", func(_ *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.SendAuthenticationInfoArg{Imsi: imsi, NumberOfRequestedVectors: 1}
		}, func(d []byte) (string, error) {
			v, err := ParseSendAuthenticationInfo(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"UpdateLocation", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.UpdateLocationArg{Imsi: imsi, MscNumber: isdn(t), VlrNumber: isdn(t)}
		}, func(d []byte) (string, error) {
			v, err := ParseUpdateLocation(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"PurgeMS", func(_ *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.PurgeMSArg{Imsi: imsi}
		}, func(d []byte) (string, error) {
			v, err := ParsePurgeMS(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"ReportSMDeliveryStatus", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.ReportSMDeliveryStatusArg{Msisdn: isdn(t), ServiceCentreAddress: isdn(t), Imsi: &imsi}
		}, func(d []byte) (string, error) {
			v, err := ParseReportSMDeliveryStatus(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"SriSmResp", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.RoutingInfoForSMRes{Imsi: imsi, LocationInfoWithLMSI: gsm_map.LocationInfoWithLMSI{NetworkNodeNumber: isdn(t)}}
		}, func(d []byte) (string, error) {
			v, err := ParseSriSmResp(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"AnyTimeInterrogation", func(t *testing.T, imsi gsm_map.IMSI) wire {
			v := struct{}{}
			return &gsm_map.AnyTimeInterrogationArg{
				SubscriberIdentity: gsm_map.NewSubscriberIdentityImsi(imsi),
				RequestedInfo:      gsm_map.RequestedInfo{SubscriberState: &v},
				GsmSCFAddress:      isdn(t),
			}
		}, func(d []byte) (string, error) {
			v, err := ParseAnyTimeInterrogation(d)
			return semField(v, err, func() string { return v.SubscriberIdentity.IMSI })
		}},
		{"AnyTimeInterrogationRes mnpInfoRes", func(_ *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.AnyTimeInterrogationRes{SubscriberInfo: gsm_map.SubscriberInfo{MnpInfoRes: &gsm_map.MNPInfoRes{Imsi: &imsi}}}
		}, func(d []byte) (string, error) {
			v, err := ParseAnyTimeInterrogationRes(d)
			return semField(v, err, func() string { return v.SubscriberInfo.MnpInfoRes.IMSI })
		}},
		{"ProvideSubscriberLocation", func(t *testing.T, imsi gsm_map.IMSI) wire {
			w, err := convertProvideSubscriberLocationArgToWire(semPSLArg())
			if err != nil {
				t.Fatalf("convertProvideSubscriberLocationArgToWire: %v", err)
			}
			w.Imsi = &imsi
			return w
		}, func(d []byte) (string, error) {
			v, err := ParseProvideSubscriberLocation(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"SubscriberLocationReport", func(t *testing.T, imsi gsm_map.IMSI) wire {
			w, err := convertSubscriberLocationReportArgToWire(semSLRArg())
			if err != nil {
				t.Fatalf("convertSubscriberLocationReportArgToWire: %v", err)
			}
			w.Imsi = &imsi
			return w
		}, func(d []byte) (string, error) {
			v, err := ParseSubscriberLocationReport(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"absentSubscriberSM imsi", func(_ *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.AbsentSubscriberSMParam{Imsi: &imsi}
		}, func(d []byte) (string, error) {
			v, err := ParseReturnErrorParameter(MapErrorAbsentSubscriberSM, d)
			if err != nil {
				return "", err
			}
			return v.(*AbsentSubscriberSMParam).IMSI, nil
		}},
		{"absentSubscriberSM userIdentifierAlert", func(_ *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.AbsentSubscriberSMParam{UserIdentifierAlert: &imsi}
		}, func(d []byte) (string, error) {
			v, err := ParseReturnErrorParameter(MapErrorAbsentSubscriberSM, d)
			if err != nil {
				return "", err
			}
			return v.(*AbsentSubscriberSMParam).UserIdentifierAlert, nil
		}},
		{"UpdateGprsLocation", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.UpdateGprsLocationArg{Imsi: imsi, SgsnNumber: isdn(t), SgsnAddress: gsm_map.GSNAddress{0x04, 192, 168, 31, 1}}
		}, func(d []byte) (string, error) {
			v, err := ParseUpdateGprsLocation(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"ProvideSubscriberInfo", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.ProvideSubscriberInfoArg{Imsi: imsi, RequestedInfo: gsm_map.RequestedInfo{LocationInformation: &struct{}{}}}
		}, func(d []byte) (string, error) {
			v, err := ParseProvideSubscriberInfo(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"CancelLocation imsi", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.CancelLocationArg{Identity: gsm_map.NewIdentityImsi(imsi)}
		}, func(d []byte) (string, error) {
			v, err := ParseCancelLocation(d)
			return semField(v, err, func() string { return v.Identity.IMSI })
		}},
		{"CancelLocation imsi-WithLMSI", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.CancelLocationArg{Identity: gsm_map.NewIdentityImsiWithLMSI(gsm_map.IMSIWithLMSI{Imsi: imsi, Lmsi: gsm_map.LMSI{1, 2, 3, 4}})}
		}, func(d []byte) (string, error) {
			v, err := ParseCancelLocation(d)
			return semField(v, err, func() string { return v.Identity.IMSIWithLMSI.IMSI })
		}},
		{"AlertServiceCentre", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.AlertServiceCentreArg{Msisdn: isdn(t), ServiceCentreAddress: isdn(t), Imsi: &imsi}
		}, func(d []byte) (string, error) {
			v, err := ParseAlertServiceCentre(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"SriSm", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.RoutingInfoForSMArg{Msisdn: isdn(t), ServiceCentreAddress: isdn(t), Imsi: &imsi}
		}, func(d []byte) (string, error) {
			v, err := ParseSriSm(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"SriResp", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.SendRoutingInfoRes{Imsi: &imsi}
		}, func(d []byte) (string, error) {
			v, err := ParseSriResp(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"MtFsm sm-RP-DA", func(t *testing.T, imsi gsm_map.IMSI) wire {
			w := semMust(convertMtFsmToArg(semMtFsm(t)))
			w.SmRPDA = gsm_map.NewSMRPDAImsi(imsi)
			return w
		}, func(d []byte) (string, error) {
			v, err := ParseMtFsm(d)
			return semField(v, err, func() string { return v.SmRpDa.IMSI })
		}},
		{"MoFsm imsi", func(t *testing.T, imsi gsm_map.IMSI) wire {
			w := semMust(convertMoFsmToArg(knownMoFsm(t)))
			w.Imsi = &imsi
			return w
		}, func(d []byte) (string, error) {
			v, err := ParseMoFsm(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"InsertSubscriberData", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.InsertSubscriberDataArg{Imsi: &imsi}
		}, func(d []byte) (string, error) {
			v, err := ParseInsertSubscriberData(d)
			return semField(v, err, func() string { return v.IMSI })
		}},
		{"SriSm HLR-Id", func(t *testing.T, imsi gsm_map.IMSI) wire {
			return &gsm_map.RoutingInfoForSMArg{Msisdn: isdn(t), ServiceCentreAddress: isdn(t), CorrelationID: &gsm_map.CorrelationID{HlrId: &imsi, SipUriB: gsm_map.SIPURI("sip:b@example.com")}}
		}, func(d []byte) (string, error) {
			v, err := ParseSriSm(d)
			return semField(v, err, func() string { return v.CorrelationID.HlrID })
		}},
	}
	for _, c := range imsiCases {
		for _, tc := range []struct {
			digits string
			raw    gsm_map.IMSI
			want   error
		}{
			{semDigits(5), nil, nil},
			{semDigits(6), nil, nil},
			{semDigits(15), nil, nil},
			{semDigits(16), nil, identity.ErrIMSILength},
			{"nondecimal MCC", gsm_map.IMSI{0xa0, 0x01, 0xf1}, identity.ErrNotDigits},
			{"nondecimal MNC", gsm_map.IMSI{0x00, 0xa1, 0xf1}, identity.ErrNotDigits},
			{"three digits", gsm_map.IMSI{0x00, 0xf1, 0xff}, identity.ErrIMSILength},
		} {
			t.Run("IMSI/"+c.name+"/"+tc.digits, func(t *testing.T) {
				raw := tc.raw
				if raw == nil {
					raw = semTBCD(t, tc.digits)
				}
				data, err := c.build(t, raw).MarshalBER()
				if err != nil {
					t.Fatalf("MarshalBER: %v", err)
				}
				got, err := c.parse(data)
				if !errors.Is(err, tc.want) {
					t.Fatalf("Parse: err = %v, want %v", err, tc.want)
				}
				if err == nil && got != tc.digits {
					t.Errorf("parsed digits = %q, want %q", got, tc.digits)
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
			map[int]error{14: identity.ErrIMEILength, 15: nil, 16: nil}},
		{"ProvideSubscriberLocation imei", func(t *testing.T, imei gsm_map.IMEI) wire {
			w, err := convertProvideSubscriberLocationArgToWire(semPSLArg())
			if err != nil {
				t.Fatalf("convertProvideSubscriberLocationArgToWire: %v", err)
			}
			w.Imei = &imei
			return w
		}, func(d []byte) error { _, err := ParseProvideSubscriberLocation(d); return err },
			map[int]error{14: identity.ErrIMEILength, 15: nil, 16: nil}},
		{"SubscriberLocationReport imei", func(t *testing.T, imei gsm_map.IMEI) wire {
			w, err := convertSubscriberLocationReportArgToWire(semSLRArg())
			if err != nil {
				t.Fatalf("convertSubscriberLocationReportArgToWire: %v", err)
			}
			w.Imei = &imei
			return w
		}, func(d []byte) error { _, err := ParseSubscriberLocationReport(d); return err },
			map[int]error{14: identity.ErrIMEILength, 15: nil, 16: nil}},
		{"UpdateLocation add-Info imeisv", func(t *testing.T, imei gsm_map.IMEI) wire {
			return &gsm_map.UpdateLocationArg{
				Imsi: semTBCD(t, "001010123456789"), MscNumber: isdn(t), VlrNumber: isdn(t),
				AddInfo: &gsm_map.ADDInfo{Imeisv: imei},
			}
		}, func(d []byte) error { _, err := ParseUpdateLocation(d); return err },
			map[int]error{14: identity.ErrIMEISVLength, 15: identity.ErrIMEISVLength, 16: nil}},
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

// TS 29.002 V19.1.0 §12.1.3 permits MCC+MNC+dummy MSIN for only MCC+MNC
// requested; peers omitting the dummy MSIN still yield the received result.
func TestParseSriSmRespMCCMNC(t *testing.T) {
	for _, tc := range []struct {
		wire, imsi string
		marshalErr error
	}{
		{"301004030001f1a0098107915155000000f1", "00101", identity.ErrIMSILength},
		{"30100403000101a0098107915155000000f1", "001010", nil},
	} {
		t.Run(tc.imsi, func(t *testing.T) {
			data, err := hex.DecodeString(tc.wire)
			if err != nil {
				t.Fatal(err)
			}
			got, err := ParseSriSmResp(data)
			if err != nil {
				t.Fatalf("ParseSriSmResp: %v", err)
			}
			if got.IMSI != tc.imsi {
				t.Errorf("IMSI = %q, want %q", got.IMSI, tc.imsi)
			}
			if got.LocationInfoWithLMSI.NetworkNodeNumber != "15550000001" {
				t.Errorf("NetworkNodeNumber = %q, want 15550000001", got.LocationInfoWithLMSI.NetworkNodeNumber)
			}
			if _, err := got.Marshal(); !errors.Is(err, tc.marshalErr) {
				t.Errorf("Marshal: err = %v, want %v", err, tc.marshalErr)
			}
		})
	}
}
