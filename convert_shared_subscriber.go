package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// --- SubscriberInfo helpers (shared between ATI and SRI response) ---

func convertSubscriberInfoToWire(s *SubscriberInfo) (*gsm_map.SubscriberInfo, error) {
	si := &gsm_map.SubscriberInfo{}

	if s.LocationInformation != nil {
		locInfo, err := convertCSLocationToAsn1(s.LocationInformation)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformation: %w", err)
		}
		si.LocationInformation = locInfo
	}

	if s.SubscriberState != nil {
		wireSs, err := convertSubscriberStateToAsn1(s.SubscriberState)
		if err != nil {
			return nil, fmt.Errorf("converting SubscriberState: %w", err)
		}
		si.SubscriberState = wireSs
	}

	if s.LocationInformationGPRS != nil {
		locGPRS, err := convertGPRSLocationToAsn1(s.LocationInformationGPRS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationGPRS: %w", err)
		}
		si.LocationInformationGPRS = locGPRS
	}

	if s.PsSubscriberState != nil {
		ps, err := convertPsSubscriberStateToWire(s.PsSubscriberState)
		if err != nil {
			return nil, fmt.Errorf("converting PsSubscriberState: %w", err)
		}
		si.PsSubscriberState = ps
	}

	if s.IMEI != "" {
		imeiBytes, err := encodeIdentityDigits(identityIMEI, s.IMEI)
		if err != nil {
			return nil, fmt.Errorf("encoding IMEI: %w", err)
		}
		imei := imeiBytes
		si.Imei = &imei
	}

	if s.MsClassmark2 != nil {
		mc := gsm_map.MSClassmark2(s.MsClassmark2)
		si.MsClassmark2 = &mc
	}

	if s.GprsMSClass != nil {
		si.GprsMSClass = convertGprsMSClassToWire(s.GprsMSClass)
	}

	if s.MnpInfoRes != nil {
		mnp, err := convertMnpInfoResToWire(s.MnpInfoRes)
		if err != nil {
			return nil, fmt.Errorf("converting MnpInfoRes: %w", err)
		}
		si.MnpInfoRes = mnp
	}

	// ImsVoiceOverPSSessionsIndication — 0..2 per TS 29.002.
	if s.ImsVoiceOverPSSessionsIndication != nil {
		// go-asn1 does not enforce ENUMERATED membership: https://github.com/gomaja/go-asn1/issues/81.
		if *s.ImsVoiceOverPSSessionsIndication < 0 || *s.ImsVoiceOverPSSessionsIndication > 2 {
			return nil, fmt.Errorf("ImsVoiceOverPSSessionsIndication out of range 0..2: %d", *s.ImsVoiceOverPSSessionsIndication)
		}
		v := *s.ImsVoiceOverPSSessionsIndication
		si.ImsVoiceOverPSSessionsIndication = &v
	}

	if s.LastUEActivityTime != nil {
		t := gsm_map.Time(s.LastUEActivityTime)
		si.LastUEActivityTime = &t
	}

	// LastRATType — Used-RAT-Type per TS 29.002 MAP-MS-DataTypes.asn:582.
	// Spec marks the enum extensible (`...`): only a listed value is sent.
	if s.LastRATType != nil {
		v := *s.LastRATType
		if !isListedUsedRATType(v) {
			return nil, fmt.Errorf("LastRATType=%d: %w", v, ErrUsedRATTypeInvalid)
		}
		si.LastRATType = &v
	}

	if s.EpsSubscriberState != nil {
		ps, err := convertPsSubscriberStateToWire(s.EpsSubscriberState)
		if err != nil {
			return nil, fmt.Errorf("converting EpsSubscriberState: %w", err)
		}
		si.EpsSubscriberState = ps
	}

	if s.LocationInformationEPS != nil {
		locEPS, err := convertEPSLocationToAsn1(s.LocationInformationEPS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationEPS: %w", err)
		}
		si.LocationInformationEPS = locEPS
	}

	if s.TimeZone != nil {
		tz := gsm_map.TimeZone(s.TimeZone)
		si.TimeZone = &tz
	}

	// DaylightSavingTime — 0..2 per TS 29.002.
	if s.DaylightSavingTime != nil {
		// go-asn1 does not enforce ENUMERATED membership: https://github.com/gomaja/go-asn1/issues/81.
		if *s.DaylightSavingTime < 0 || *s.DaylightSavingTime > 2 {
			return nil, fmt.Errorf("DaylightSavingTime out of range 0..2: %d", *s.DaylightSavingTime)
		}
		dst := gsm_map.DaylightSavingTime(*s.DaylightSavingTime)
		si.DaylightSavingTime = &dst
	}

	if s.LocationInformation5GS != nil {
		loc5gs, err := convertLocationInformation5GSToWire(s.LocationInformation5GS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformation5GS: %w", err)
		}
		si.LocationInformation5GS = loc5gs
	}

	return si, nil
}

func convertWireToSubscriberInfo(si *gsm_map.SubscriberInfo) (*SubscriberInfo, error) {
	out := &SubscriberInfo{}

	if si.LocationInformation != nil {
		locInfo, err := convertAsn1ToCSLocation(si.LocationInformation)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformation: %w", err)
		}
		out.LocationInformation = locInfo
	}

	if si.SubscriberState != nil {
		pubSs, err := convertAsn1ToSubscriberState(si.SubscriberState)
		if err != nil {
			return nil, fmt.Errorf("decoding SubscriberState: %w", err)
		}
		out.SubscriberState = pubSs
	}

	if si.LocationInformationGPRS != nil {
		locGPRS, err := convertAsn1ToGPRSLocation(si.LocationInformationGPRS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationGPRS: %w", err)
		}
		out.LocationInformationGPRS = locGPRS
	}

	if si.PsSubscriberState != nil {
		ps, err := convertWireToPsSubscriberState(si.PsSubscriberState)
		if err != nil {
			return nil, fmt.Errorf("converting PsSubscriberState: %w", err)
		}
		out.PsSubscriberState = ps
	}

	// IMEI is TBCD-STRING (SIZE(8)) per 3GPP TS 29.002. When present on
	// the wire it must be exactly 8 octets — empty/non-8-octet IMEI is
	// a spec violation, not "absent".
	if si.Imei != nil {
		imei, err := decodeIdentityDigits(identityIMEI, *si.Imei)
		if err != nil {
			return nil, fmt.Errorf("decoding IMEI: %w", err)
		}
		out.IMEI = imei
	}

	if si.MsClassmark2 != nil {
		out.MsClassmark2 = *si.MsClassmark2
	}

	if si.GprsMSClass != nil {
		out.GprsMSClass = convertWireToGprsMSClass(si.GprsMSClass)
	}

	if si.MnpInfoRes != nil {
		mnp, err := convertWireToMnpInfoRes(si.MnpInfoRes)
		if err != nil {
			return nil, fmt.Errorf("converting MnpInfoRes: %w", err)
		}
		out.MnpInfoRes = mnp
	}

	// ImsVoiceOverPSSessionsIndication — 0..2 per TS 29.002.
	if si.ImsVoiceOverPSSessionsIndication != nil {
		// go-asn1 does not enforce ENUMERATED membership: https://github.com/gomaja/go-asn1/issues/81.
		v, err := narrowInt64Range(int64(*si.ImsVoiceOverPSSessionsIndication), 0, 2, "ImsVoiceOverPSSessionsIndication")
		if err != nil {
			return nil, err
		}
		iv := ImsVoiceOverPSSessionsIndication(v)
		out.ImsVoiceOverPSSessionsIndication = &iv
	}

	if si.LastUEActivityTime != nil {
		out.LastUEActivityTime = *si.LastUEActivityTime
	}

	// LastRATType — Used-RAT-Type per TS 29.002 (extensible enum): an
	// unknown value is kept (3GPP TS 29.002 V19.1.0 §17.1.4) and Marshal
	// refuses it.
	if si.LastRATType != nil {
		v := *si.LastRATType
		out.LastRATType = &v
	}

	if si.EpsSubscriberState != nil {
		ps, err := convertWireToPsSubscriberState(si.EpsSubscriberState)
		if err != nil {
			return nil, fmt.Errorf("converting EpsSubscriberState: %w", err)
		}
		out.EpsSubscriberState = ps
	}

	if si.LocationInformationEPS != nil {
		locEPS, err := convertAsn1ToEPSLocation(si.LocationInformationEPS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformationEPS: %w", err)
		}
		out.LocationInformationEPS = locEPS
	}

	if si.TimeZone != nil {
		out.TimeZone = *si.TimeZone
	}

	// DaylightSavingTime — 0..2 per TS 29.002.
	if si.DaylightSavingTime != nil {
		// go-asn1 does not enforce ENUMERATED membership: https://github.com/gomaja/go-asn1/issues/81.
		v, err := narrowInt64Range(int64(*si.DaylightSavingTime), 0, 2, "DaylightSavingTime")
		if err != nil {
			return nil, err
		}
		out.DaylightSavingTime = &v
	}

	if si.LocationInformation5GS != nil {
		loc5gs, err := convertWireToLocationInformation5GS(si.LocationInformation5GS)
		if err != nil {
			return nil, fmt.Errorf("converting LocationInformation5GS: %w", err)
		}
		out.LocationInformation5GS = loc5gs
	}

	return out, nil
}

// --- PS-SubscriberState (opCode 71) ---

// psSubscriberStateCount returns the number of alternatives set in p.
func psSubscriberStateCount(p *PsSubscriberState) int {
	c := 0
	if p.NotProvidedFromSGSNorMME {
		c++
	}
	if p.PsDetached {
		c++
	}
	if p.PsAttachedNotReachableForPaging {
		c++
	}
	if p.PsAttachedReachableForPaging {
		c++
	}
	if len(p.PsPDPActiveNotReachableForPaging) > 0 {
		c++
	}
	if len(p.PsPDPActiveReachableForPaging) > 0 {
		c++
	}
	if p.NetDetNotReachable != nil {
		c++
	}
	return c
}

func convertPsSubscriberStateToWire(p *PsSubscriberState) (*gsm_map.PSSubscriberState, error) {
	n := psSubscriberStateCount(p)
	if n == 0 {
		return nil, ErrAtiPsSubscriberStateNoAlternative
	}
	if n > 1 {
		return nil, ErrAtiPsSubscriberStateMultipleAlternatives
	}

	switch {
	case p.NotProvidedFromSGSNorMME:
		v := gsm_map.NewPSSubscriberStateNotProvidedFromSGSNorMME(struct{}{})
		return &v, nil
	case p.PsDetached:
		v := gsm_map.NewPSSubscriberStatePsDetached(struct{}{})
		return &v, nil
	case p.PsAttachedNotReachableForPaging:
		v := gsm_map.NewPSSubscriberStatePsAttachedNotReachableForPaging(struct{}{})
		return &v, nil
	case p.PsAttachedReachableForPaging:
		v := gsm_map.NewPSSubscriberStatePsAttachedReachableForPaging(struct{}{})
		return &v, nil
	case len(p.PsPDPActiveNotReachableForPaging) > 0:
		list, err := decodePDPContextInfoList(p.PsPDPActiveNotReachableForPaging)
		if err != nil {
			return nil, fmt.Errorf("decoding PsPDPActiveNotReachableForPaging: %w", err)
		}
		v := gsm_map.NewPSSubscriberStatePsPDPActiveNotReachableForPaging(list)
		return &v, nil
	case len(p.PsPDPActiveReachableForPaging) > 0:
		list, err := decodePDPContextInfoList(p.PsPDPActiveReachableForPaging)
		if err != nil {
			return nil, fmt.Errorf("decoding PsPDPActiveReachableForPaging: %w", err)
		}
		v := gsm_map.NewPSSubscriberStatePsPDPActiveReachableForPaging(list)
		return &v, nil
	case p.NetDetNotReachable != nil:
		// NotReachableReason — 0..3 per TS 29.002.
		// go-asn1 does not enforce ENUMERATED membership: https://github.com/gomaja/go-asn1/issues/81.
		if *p.NetDetNotReachable < 0 || *p.NetDetNotReachable > 3 {
			return nil, fmt.Errorf("PsSubscriberState.NetDetNotReachable out of range 0..3: %d", *p.NetDetNotReachable)
		}
		v := gsm_map.NewPSSubscriberStateNetDetNotReachable(gsm_map.NotReachableReason(int64(*p.NetDetNotReachable)))
		return &v, nil
	}
	return nil, ErrAtiPsSubscriberStateNoAlternative
}

func convertWireToPsSubscriberState(w *gsm_map.PSSubscriberState) (*PsSubscriberState, error) {
	out := &PsSubscriberState{}
	switch w.Choice {
	case gsm_map.PSSubscriberStateChoiceNotProvidedFromSGSNorMME:
		out.NotProvidedFromSGSNorMME = true
	case gsm_map.PSSubscriberStateChoicePsDetached:
		out.PsDetached = true
	case gsm_map.PSSubscriberStateChoicePsAttachedNotReachableForPaging:
		out.PsAttachedNotReachableForPaging = true
	case gsm_map.PSSubscriberStateChoicePsAttachedReachableForPaging:
		out.PsAttachedReachableForPaging = true
	case gsm_map.PSSubscriberStateChoicePsPDPActiveNotReachableForPaging:
		enc, err := encodePDPContextInfoList(w.PsPDPActiveNotReachableForPaging)
		if err != nil {
			return nil, fmt.Errorf("encoding PsPDPActiveNotReachableForPaging: %w", err)
		}
		out.PsPDPActiveNotReachableForPaging = enc
	case gsm_map.PSSubscriberStateChoicePsPDPActiveReachableForPaging:
		enc, err := encodePDPContextInfoList(w.PsPDPActiveReachableForPaging)
		if err != nil {
			return nil, fmt.Errorf("encoding PsPDPActiveReachableForPaging: %w", err)
		}
		out.PsPDPActiveReachableForPaging = enc
	case gsm_map.PSSubscriberStateChoiceNetDetNotReachable:
		if w.NetDetNotReachable == nil {
			return nil, fmt.Errorf("PsSubscriberState: NetDetNotReachable alternative selected but reason is nil")
		}
		// NotReachableReason — 0..3 per TS 29.002 (msPurged / imsiDetached /
		// restrictedArea / notRegistered).
		// go-asn1 does not enforce ENUMERATED membership: https://github.com/gomaja/go-asn1/issues/81.
		v, err := narrowInt64Range(int64(*w.NetDetNotReachable), 0, 3, "PsSubscriberState.NetDetNotReachable")
		if err != nil {
			return nil, err
		}
		out.NetDetNotReachable = &v
	default:
		return nil, fmt.Errorf("PsSubscriberState: unknown CHOICE value %d", w.Choice)
	}
	return out, nil
}

// encodePDPContextInfoList serializes each gsm_map.PDPContextInfo entry to
// its BER-encoded bytes, keeping them opaque from the caller's perspective.
// Each entry is checked by the strict BER codec when it is marshalled.
func encodePDPContextInfoList(list *gsm_map.PDPContextInfoList) ([]HexBytes, error) {
	if list == nil {
		list = &gsm_map.PDPContextInfoList{}
	}

	out := make([]HexBytes, len(list.Values))
	for i := range list.Values {
		ctx := list.Values[i]
		enc, err := ctx.MarshalBER()
		if err != nil {
			return nil, fmt.Errorf("PDPContextInfo[%d]: %w", i, err)
		}
		out[i] = enc
	}
	return out, nil
}

// decodePDPContextInfoList deserializes each opaque PDPContextInfo entry
// back into its gsm_map.PDPContextInfo struct. The list size is checked
// when its containing value is marshalled.
func decodePDPContextInfoList(list []HexBytes) (*gsm_map.PDPContextInfoList, error) {
	out := gsm_map.PDPContextInfoList{Values: make([]gsm_map.PDPContextInfo, len(list))}
	for i, b := range list {
		var ctx gsm_map.PDPContextInfo
		if err := ctx.UnmarshalBER(b); err != nil {
			return nil, fmt.Errorf("PDPContextInfo[%d]: %w", i, err)
		}
		out.Values[i] = ctx
	}
	return &out, nil
}

// --- MNPInfoRes (opCode 71) ---

func convertMnpInfoResToWire(m *MnpInfoRes) (*gsm_map.MNPInfoRes, error) {
	out := &gsm_map.MNPInfoRes{}

	if m.RouteingNumber != nil {
		rn := gsm_map.RouteingNumber(m.RouteingNumber)
		out.RouteingNumber = &rn
	}

	if m.IMSI != "" {
		b, err := encodeIdentityDigits(identityIMSI, m.IMSI)
		if err != nil {
			return nil, fmt.Errorf(errEncodingIMSI, err)
		}
		imsi := b
		out.Imsi = &imsi
	}

	if m.MSISDN != "" {
		enc, err := encodeAddressField(m.MSISDN, m.MSISDNNature, m.MSISDNPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MSISDN: %w", err)
		}
		as := enc
		out.Msisdn = &as
	}

	// NumberPortabilityStatus — defined values 0,1,2,4,5 per TS 29.002.
	if m.NumberPortabilityStatus != nil {
		// Sender accepts defined values; receivers ignore unknown values (3GPP TS 29.002 V19.1.0 §17.7.1).
		switch *m.NumberPortabilityStatus {
		case MnpNotKnownToBePorted, MnpOwnNumberPortedOut, MnpForeignNumberPortedToForeignNetwork,
			MnpOwnNumberNotPortedOut, MnpForeignNumberPortedIn:
		default:
			return nil, fmt.Errorf("MnpInfoRes: NumberPortabilityStatus has undefined value %d", *m.NumberPortabilityStatus)
		}
		v := *m.NumberPortabilityStatus
		out.NumberPortabilityStatus = &v
	}

	return out, nil
}

func convertWireToMnpInfoRes(w *gsm_map.MNPInfoRes) (*MnpInfoRes, error) {
	out := &MnpInfoRes{}

	if w.RouteingNumber != nil {
		out.RouteingNumber = *w.RouteingNumber
	}

	if w.Imsi != nil && len(*w.Imsi) > 0 {
		imsi, err := decodeIdentityDigits(identityIMSI, *w.Imsi)
		if err != nil {
			return nil, fmt.Errorf("decoding IMSI: %w", err)
		}
		out.IMSI = imsi
	}

	if w.Msisdn != nil {
		digits, nat, pl, err := decodeAddressWithDigits(*w.Msisdn, ErrMnpInfoResMSISDNDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding MSISDN: %w", err)
		}
		out.MSISDN = digits
		out.MSISDNNature = nat
		out.MSISDNPlan = pl
	}

	if w.NumberPortabilityStatus != nil {
		// NumberPortabilityStatus — ENUMERATED { 0, 1, 2, 4, 5 } per TS 29.002.
		// Spec exception: "reception of other values than the ones listed the
		// receiver shall ignore the whole NumberPortabilityStatus;".
		// Unknown extensions are ignored per 3GPP TS 29.002 V19.1.0 §17.7.1.
		switch *w.NumberPortabilityStatus {
		case MnpNotKnownToBePorted, MnpOwnNumberPortedOut,
			MnpForeignNumberPortedToForeignNetwork,
			MnpOwnNumberNotPortedOut, MnpForeignNumberPortedIn:
			v := *w.NumberPortabilityStatus
			out.NumberPortabilityStatus = &v
		}
		// Unknown value: leave field nil per spec.
	}

	return out, nil
}

// --- GprsMSClass (opCode 71) ---

func convertGprsMSClassToWire(g *GprsMSClass) *gsm_map.GPRSMSClass {
	out := &gsm_map.GPRSMSClass{
		MSNetworkCapability: gsm_map.MSNetworkCapability(g.MSNetworkCapability),
	}
	if g.MSRadioAccessCapability != nil {
		rac := gsm_map.MSRadioAccessCapability(g.MSRadioAccessCapability)
		out.MSRadioAccessCapability = &rac
	}
	return out
}

func convertWireToGprsMSClass(w *gsm_map.GPRSMSClass) *GprsMSClass {
	out := &GprsMSClass{
		MSNetworkCapability: w.MSNetworkCapability,
	}
	if w.MSRadioAccessCapability != nil {
		out.MSRadioAccessCapability = *w.MSRadioAccessCapability
	}
	return out
}

// --- UserCSGInformation (opCode 71) ---

func convertUserCSGInformationToWire(u *UserCSGInformation) (*gsm_map.UserCSGInformation, error) {
	csgID, err := bitStringToWire("UserCSGInformation.CsgID", u.CsgID, u.CsgIDBits)
	if err != nil {
		return nil, err
	}
	out := &gsm_map.UserCSGInformation{CsgId: csgID}
	if u.AccessMode != nil {
		out.AccessMode = []byte(u.AccessMode)
	}
	if u.CMI != nil {
		out.Cmi = []byte(u.CMI)
	}
	return out, nil
}

func convertWireToUserCSGInformation(w *gsm_map.UserCSGInformation) *UserCSGInformation {
	out := &UserCSGInformation{
		CsgID:     append([]byte(nil), w.CsgId.Bytes...),
		CsgIDBits: w.CsgId.BitLength,
	}
	if w.AccessMode != nil {
		out.AccessMode = w.AccessMode
	}
	if w.Cmi != nil {
		out.CMI = w.Cmi
	}
	return out
}

// --- LocationInformation5GS (opCode 71) ---

// isListedUsedRATType reports whether v is one of the Used-RAT-Type values
// of 3GPP TS 29.002 V19.1.0 §17.7.1, utran (0) to nb-iot (5). The type is
// extensible: the decoders keep any other value (§17.1.4).
func isListedUsedRATType(v UsedRatType) bool {
	return v >= UsedRatUTRAN && v <= UsedRatNBIOT
}

func convertLocationInformation5GSToWire(l *LocationInformation5GS) (*gsm_map.LocationInformation5GS, error) {
	out := &gsm_map.LocationInformation5GS{}

	if l.NrCellGlobalIdentity != nil {
		cgi := gsm_map.NRCGI(l.NrCellGlobalIdentity)
		out.NrCellGlobalIdentity = &cgi
	}

	if l.EUtranCellGlobalIdentity != nil {
		cgi := gsm_map.EUTRANCGI(l.EUtranCellGlobalIdentity)
		out.EUtranCellGlobalIdentity = &cgi
	}

	if l.GeographicalInformation != nil {
		raw, err := l.GeographicalInformation.Encode()
		if err != nil {
			return nil, fmt.Errorf("encoding GeographicalInformation: %w", err)
		}
		gi := raw
		out.GeographicalInformation = &gi
	}

	if l.GeodeticInformation != nil {
		gd := gsm_map.GeodeticInformation(l.GeodeticInformation)
		out.GeodeticInformation = &gd
	}

	if l.AmfAddress != nil {
		amf := gsm_map.FQDN(l.AmfAddress)
		out.AmfAddress = &amf
	}

	if l.TrackingAreaIdentity != nil {
		ta := gsm_map.TAId(l.TrackingAreaIdentity)
		out.TrackingAreaIdentity = &ta
	}

	out.CurrentLocationRetrieved = boolToNullPtr(l.CurrentLocationRetrieved)

	if l.AgeOfLocationInformation != nil {
		age := gsm_map.AgeOfLocationInformation(*l.AgeOfLocationInformation)
		out.AgeOfLocationInformation = &age
	}

	if l.VplmnID != nil {
		p := gsm_map.PLMNId(l.VplmnID)
		out.VplmnId = &p
	}

	if l.LocalTimeZone != nil {
		tz := gsm_map.TimeZone(l.LocalTimeZone)
		out.LocaltimeZone = &tz
	}

	// RatType — Used-RAT-Type per TS 29.002 (extensible enum): only a
	// listed value is sent.
	if l.RatType != nil {
		v := *l.RatType
		if !isListedUsedRATType(v) {
			return nil, fmt.Errorf("LocationInformation5GS.RatType=%d: %w", v, ErrUsedRATTypeInvalid)
		}
		out.RatType = &v
	}

	if l.NrTrackingAreaIdentity != nil {
		ta := gsm_map.NRTAId(l.NrTrackingAreaIdentity)
		out.NrTrackingAreaIdentity = &ta
	}

	return out, nil
}

func convertWireToLocationInformation5GS(w *gsm_map.LocationInformation5GS) (*LocationInformation5GS, error) {
	out := &LocationInformation5GS{}

	if w.NrCellGlobalIdentity != nil {
		out.NrCellGlobalIdentity = *w.NrCellGlobalIdentity
	}

	if w.EUtranCellGlobalIdentity != nil {
		out.EUtranCellGlobalIdentity = *w.EUtranCellGlobalIdentity
	}

	if w.GeographicalInformation != nil {
		gi, err := DecodeGeographicalInfo(*w.GeographicalInformation)
		if err != nil {
			return nil, fmt.Errorf("decoding GeographicalInformation: %w", err)
		}
		out.GeographicalInformation = gi
	}

	if w.GeodeticInformation != nil {
		out.GeodeticInformation = *w.GeodeticInformation
	}

	if w.AmfAddress != nil {
		out.AmfAddress = *w.AmfAddress
	}

	if w.TrackingAreaIdentity != nil {
		out.TrackingAreaIdentity = *w.TrackingAreaIdentity
	}

	out.CurrentLocationRetrieved = nullPtrToBool(w.CurrentLocationRetrieved)

	if w.AgeOfLocationInformation != nil {
		v := int(*w.AgeOfLocationInformation)
		out.AgeOfLocationInformation = &v
	}

	if w.VplmnId != nil {
		p := *w.VplmnId

		out.VplmnID = p
	}

	if w.LocaltimeZone != nil {
		out.LocalTimeZone = *w.LocaltimeZone
	}

	// RatType — Used-RAT-Type per TS 29.002 (extensible enum): an unknown
	// value is kept (3GPP TS 29.002 V19.1.0 §17.1.4) and Marshal refuses it.
	if w.RatType != nil {
		v := *w.RatType
		out.RatType = &v
	}

	if w.NrTrackingAreaIdentity != nil {
		out.NrTrackingAreaIdentity = *w.NrTrackingAreaIdentity
	}

	return out, nil
}
