package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// --- SendAuthenticationInfo (opCode 56) converters ---

// isValidRequestingNodeType reports whether v is one of the RequestingNodeType
// values defined in 3GPP TS 29.002 (vlr=0, sgsn=1, s-cscf=2, bsf=3,
// gan-aaa-server=4, wlan-aaa-server=5, mme=16, mme-sgsn=17).
func isValidRequestingNodeType(v RequestingNodeType) bool {
	switch v {
	case RequestingNodeVlr,
		RequestingNodeSgsn,
		RequestingNodeSCscf,
		RequestingNodeBsf,
		RequestingNodeGanAAAServer,
		RequestingNodeWlanAAAServer,
		RequestingNodeMme,
		RequestingNodeMmeSgsn:
		return true
	}
	return false
}

// convertReSynchronisationInfoToWire converts the public ReSynchronisationInfo
// into the wire-level gsm_map.ReSynchronisationInfo.
func convertReSynchronisationInfoToWire(r *ReSynchronisationInfo) (*gsm_map.ReSynchronisationInfo, error) {
	if r == nil {
		return nil, nil
	}

	return &gsm_map.ReSynchronisationInfo{
		Rand: gsm_map.RAND(r.RAND),
		Auts: gsm_map.AUTS(r.AUTS),
	}, nil
}

// convertWireToReSynchronisationInfo converts a wire-level
// gsm_map.ReSynchronisationInfo into the public ReSynchronisationInfo.
func convertWireToReSynchronisationInfo(w *gsm_map.ReSynchronisationInfo) (*ReSynchronisationInfo, error) {
	if w == nil {
		return nil, nil
	}

	return &ReSynchronisationInfo{
		RAND: HexBytes(w.Rand),
		AUTS: HexBytes(w.Auts),
	}, nil
}

// convertAuthenticationSetListToWire converts the public CHOICE
// AuthenticationSetList into the wire-level gsm_map.AuthenticationSetList.
func convertAuthenticationSetListToWire(a *AuthenticationSetList) (*gsm_map.AuthenticationSetList, error) {
	if a == nil {
		return nil, nil
	}
	hasTriplets := len(a.Triplets) > 0
	hasQuintuplets := len(a.Quintuplets) > 0
	if hasTriplets && hasQuintuplets {
		return nil, ErrAuthenticationSetListMultipleAlternatives
	}
	if !hasTriplets && !hasQuintuplets {
		return nil, ErrAuthenticationSetListNoAlternative
	}
	if hasTriplets {
		list := gsm_map.TripletList{Values: make([]gsm_map.AuthenticationTriplet, len(a.Triplets))}
		for i := range a.Triplets {
			list.Values[i] = gsm_map.AuthenticationTriplet{
				Rand: gsm_map.RAND(a.Triplets[i].RAND),
				Sres: gsm_map.SRES(a.Triplets[i].SRES),
				Kc:   gsm_map.Kc(a.Triplets[i].Kc),
			}
		}
		v := gsm_map.NewAuthenticationSetListTripletList(&list)
		return &v, nil
	}
	list := gsm_map.QuintupletList{Values: make([]gsm_map.AuthenticationQuintuplet, len(a.Quintuplets))}
	for i := range a.Quintuplets {
		list.Values[i] = gsm_map.AuthenticationQuintuplet{
			Rand: gsm_map.RAND(a.Quintuplets[i].RAND),
			Xres: gsm_map.XRES(a.Quintuplets[i].XRES),
			Ck:   gsm_map.CK(a.Quintuplets[i].CK),
			Ik:   gsm_map.IK(a.Quintuplets[i].IK),
			Autn: gsm_map.AUTN(a.Quintuplets[i].AUTN),
		}
	}
	v := gsm_map.NewAuthenticationSetListQuintupletList(&list)
	return &v, nil
}

// convertWireToAuthenticationSetList converts a wire-level
// gsm_map.AuthenticationSetList back into the public CHOICE.
func convertWireToAuthenticationSetList(w *gsm_map.AuthenticationSetList) (*AuthenticationSetList, error) {
	if w == nil {
		return nil, nil
	}
	switch w.Choice {
	case gsm_map.AuthenticationSetListChoiceTripletList:
		triplets := w.TripletList
		out := make([]AuthenticationTriplet, len(triplets.Values))
		for i, t := range triplets.Values {
			out[i] = AuthenticationTriplet{
				RAND: HexBytes(t.Rand),
				SRES: HexBytes(t.Sres),
				Kc:   HexBytes(t.Kc),
			}
		}
		return &AuthenticationSetList{Triplets: out}, nil
	case gsm_map.AuthenticationSetListChoiceQuintupletList:
		quintuplets := w.QuintupletList
		out := make([]AuthenticationQuintuplet, len(quintuplets.Values))
		for i, q := range quintuplets.Values {
			out[i] = AuthenticationQuintuplet{
				RAND: HexBytes(q.Rand),
				XRES: HexBytes(q.Xres),
				CK:   HexBytes(q.Ck),
				IK:   HexBytes(q.Ik),
				AUTN: HexBytes(q.Autn),
			}
		}
		return &AuthenticationSetList{Quintuplets: out}, nil
	default:
		return nil, fmt.Errorf("%w: sai: unknown AuthenticationSetList CHOICE %d", ErrAuthenticationSetListUnknownAlternative, w.Choice)
	}
}

// convertEpcAVToWire converts the public EpcAV into the wire-level gsm_map.EPCAV.
func convertEpcAVToWire(e *EpcAV) gsm_map.EPCAV {
	return gsm_map.EPCAV{
		Rand:  gsm_map.RAND(e.RAND),
		Xres:  gsm_map.XRES(e.XRES),
		Autn:  gsm_map.AUTN(e.AUTN),
		Kasme: gsm_map.KASME(e.KASME),
	}
}

// convertWireToEpcAV converts a wire-level gsm_map.EPCAV into the public EpcAV.
func convertWireToEpcAV(w *gsm_map.EPCAV) EpcAV {
	return EpcAV{
		RAND:  HexBytes(w.Rand),
		XRES:  HexBytes(w.Xres),
		AUTN:  HexBytes(w.Autn),
		KASME: HexBytes(w.Kasme),
	}
}

// convertSendAuthenticationInfoToArg converts the public SendAuthenticationInfo
// into the wire-level gsm_map.SendAuthenticationInfoArg.
func convertSendAuthenticationInfoToArg(s *SendAuthenticationInfo) (*gsm_map.SendAuthenticationInfoArg, error) {
	imsiBytes, err := encodeIdentityDigits(identityIMSI, s.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	resync, err := convertReSynchronisationInfoToWire(s.ReSynchronisationInfo)
	if err != nil {
		return nil, err
	}

	arg := &gsm_map.SendAuthenticationInfoArg{
		Imsi:                         imsiBytes,
		NumberOfRequestedVectors:     int64(s.NumberOfRequestedVectors),
		SegmentationProhibited:       boolToNullPtr(s.SegmentationProhibited),
		ImmediateResponsePreferred:   boolToNullPtr(s.ImmediateResponsePreferred),
		ReSynchronisationInfo:        resync,
		AdditionalVectorsAreForEPS:   boolToNullPtr(s.AdditionalVectorsAreForEPS),
		UeUsageTypeRequestIndication: boolToNullPtr(s.UeUsageTypeRequestIndication),
	}

	if s.RequestingNodeType != nil {
		if !isValidRequestingNodeType(*s.RequestingNodeType) {
			return nil, fmt.Errorf("%w: got %d", ErrSaiInvalidRequestingNodeType, *s.RequestingNodeType)
		}
		v := *s.RequestingNodeType
		arg.RequestingNodeType = &v
	}
	if len(s.RequestingPLMNId) > 0 {
		v := gsm_map.PLMNId(s.RequestingPLMNId)
		arg.RequestingPLMNId = &v
	}
	if s.NumberOfRequestedAdditionalVectors != nil {
		v := int64(*s.NumberOfRequestedAdditionalVectors)
		arg.NumberOfRequestedAdditionalVectors = &v
	}

	return arg, nil
}

// convertArgToSendAuthenticationInfo converts a wire-level
// gsm_map.SendAuthenticationInfoArg back into the public SendAuthenticationInfo.
func convertArgToSendAuthenticationInfo(arg *gsm_map.SendAuthenticationInfoArg) (*SendAuthenticationInfo, error) {
	imsi, err := decodeIdentityDigits(identityIMSI, arg.Imsi)
	if err != nil {
		return nil, fmt.Errorf("decoding IMSI: %w", err)
	}

	resync, err := convertWireToReSynchronisationInfo(arg.ReSynchronisationInfo)
	if err != nil {
		return nil, err
	}

	out := &SendAuthenticationInfo{
		IMSI:                         imsi,
		NumberOfRequestedVectors:     int(arg.NumberOfRequestedVectors),
		SegmentationProhibited:       nullPtrToBool(arg.SegmentationProhibited),
		ImmediateResponsePreferred:   nullPtrToBool(arg.ImmediateResponsePreferred),
		ReSynchronisationInfo:        resync,
		AdditionalVectorsAreForEPS:   nullPtrToBool(arg.AdditionalVectorsAreForEPS),
		UeUsageTypeRequestIndication: nullPtrToBool(arg.UeUsageTypeRequestIndication),
	}

	// RequestingNodeType — ENUMERATED { vlr(0), sgsn(1), ..., s-cscf(2),
	// bsf(3), gan-aaa-server(4), wlan-aaa-server(5), mme(16), mme-sgsn(17) }
	// per TS 29.002. Spec exception handling:
	// received values in the range (6-15) shall be treated as "vlr"
	// received values greater than 17 shall be treated as "sgsn"
	// Apply the receiver mapping before exposing the value as int.
	if arg.RequestingNodeType != nil {
		// A negative value lies outside both rules; the type is extensible,
		// so it is kept (3GPP TS 29.002 V19.1.0 §17.1.4).
		raw64 := int64(*arg.RequestingNodeType)
		switch {
		case raw64 >= 6 && raw64 <= 15:
			raw64 = int64(RequestingNodeVlr)
		case raw64 > 17:
			raw64 = int64(RequestingNodeSgsn)
		}
		v := RequestingNodeType(raw64)
		out.RequestingNodeType = &v
	}
	if arg.RequestingPLMNId != nil {
		plmn := *arg.RequestingPLMNId

		out.RequestingPLMNId = HexBytes(plmn)
	}
	if arg.NumberOfRequestedAdditionalVectors != nil {
		v := *arg.NumberOfRequestedAdditionalVectors

		iv := int(v)
		out.NumberOfRequestedAdditionalVectors = &iv
	}

	return out, nil
}

// convertSendAuthenticationInfoResToRes converts the public
// SendAuthenticationInfoRes into the wire-level gsm_map.SendAuthenticationInfoRes.
func convertSendAuthenticationInfoResToRes(s *SendAuthenticationInfoRes) (*gsm_map.SendAuthenticationInfoRes, error) {
	res := &gsm_map.SendAuthenticationInfoRes{}

	if s.AuthenticationSetList != nil {
		asl, err := convertAuthenticationSetListToWire(s.AuthenticationSetList)
		if err != nil {
			return nil, err
		}
		res.AuthenticationSetList = asl
	}

	if len(s.EpsAuthenticationSetList) > 0 {
		list := gsm_map.EPSAuthenticationSetList{Values: make([]gsm_map.EPCAV, len(s.EpsAuthenticationSetList))}
		for i := range s.EpsAuthenticationSetList {
			list.Values[i] = convertEpcAVToWire(&s.EpsAuthenticationSetList[i])
		}
		res.EpsAuthenticationSetList = &list
	}

	if len(s.UeUsageType) > 0 {
		v := gsm_map.UEUsageType(s.UeUsageType)
		res.UeUsageType = &v
	}

	return res, nil
}

// convertResToSendAuthenticationInfoRes converts a wire-level
// gsm_map.SendAuthenticationInfoRes back into the public type.
func convertResToSendAuthenticationInfoRes(res *gsm_map.SendAuthenticationInfoRes) (*SendAuthenticationInfoRes, error) {
	out := &SendAuthenticationInfoRes{}

	if res.AuthenticationSetList != nil {
		asl, err := convertWireToAuthenticationSetList(res.AuthenticationSetList)
		if err != nil {
			return nil, err
		}
		out.AuthenticationSetList = asl
	}

	if res.EpsAuthenticationSetList != nil && len(res.EpsAuthenticationSetList.Values) > 0 {
		list := make([]EpcAV, len(res.EpsAuthenticationSetList.Values))
		for i := range res.EpsAuthenticationSetList.Values {
			list[i] = convertWireToEpcAV(&res.EpsAuthenticationSetList.Values[i])
		}
		out.EpsAuthenticationSetList = list
	}

	if res.UeUsageType != nil {
		ue := *res.UeUsageType

		out.UeUsageType = HexBytes(ue)
	}

	return out, nil
}
