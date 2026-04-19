package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
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
// into the wire-level gsm_map.ReSynchronisationInfo. Validates RAND (16 octets)
// and AUTS (14 octets) per 3GPP TS 29.002.
func convertReSynchronisationInfoToWire(r *ReSynchronisationInfo) (*gsm_map.ReSynchronisationInfo, error) {
	if r == nil {
		return nil, nil
	}
	if len(r.RAND) != 16 {
		return nil, fmt.Errorf("ReSynchronisationInfo: RAND must be exactly 16 octets, got %d", len(r.RAND))
	}
	if len(r.AUTS) != 14 {
		return nil, fmt.Errorf("ReSynchronisationInfo: AUTS must be exactly 14 octets, got %d", len(r.AUTS))
	}
	return &gsm_map.ReSynchronisationInfo{
		Rand: gsm_map.RAND(r.RAND),
		Auts: gsm_map.AUTS(r.AUTS),
	}, nil
}

// convertWireToReSynchronisationInfo converts a wire-level
// gsm_map.ReSynchronisationInfo into the public ReSynchronisationInfo,
// enforcing the spec-mandated 16-octet RAND and 14-octet AUTS lengths on
// decode (symmetric with the encoder).
func convertWireToReSynchronisationInfo(w *gsm_map.ReSynchronisationInfo) (*ReSynchronisationInfo, error) {
	if w == nil {
		return nil, nil
	}
	if len(w.Rand) != 16 {
		return nil, fmt.Errorf("ReSynchronisationInfo: RAND must be exactly 16 octets, got %d", len(w.Rand))
	}
	if len(w.Auts) != 14 {
		return nil, fmt.Errorf("ReSynchronisationInfo: AUTS must be exactly 14 octets, got %d", len(w.Auts))
	}
	return &ReSynchronisationInfo{
		RAND: HexBytes(w.Rand),
		AUTS: HexBytes(w.Auts),
	}, nil
}

// validateTriplet enforces the fixed-length requirements on a GSM triplet
// per 3GPP TS 29.002: RAND 16 octets, SRES 4 octets, Kc 8 octets.
func validateTriplet(t *AuthenticationTriplet, idx int) error {
	if len(t.RAND) != 16 {
		return fmt.Errorf("sai: triplet[%d] RAND must be exactly 16 octets, got %d", idx, len(t.RAND))
	}
	if len(t.SRES) != 4 {
		return fmt.Errorf("sai: triplet[%d] SRES must be exactly 4 octets, got %d", idx, len(t.SRES))
	}
	if len(t.Kc) != 8 {
		return fmt.Errorf("sai: triplet[%d] Kc must be exactly 8 octets, got %d", idx, len(t.Kc))
	}
	return nil
}

// validateQuintuplet enforces the fixed-length requirements on a UMTS
// quintuplet per 3GPP TS 29.002: RAND 16, XRES 4..16, CK 16, IK 16,
// AUTN 16.
func validateQuintuplet(q *AuthenticationQuintuplet, idx int) error {
	if len(q.RAND) != 16 {
		return fmt.Errorf("sai: quintuplet[%d] RAND must be exactly 16 octets, got %d", idx, len(q.RAND))
	}
	if len(q.XRES) < 4 || len(q.XRES) > 16 {
		return fmt.Errorf("sai: quintuplet[%d] XRES must be 4..16 octets, got %d", idx, len(q.XRES))
	}
	if len(q.CK) != 16 {
		return fmt.Errorf("sai: quintuplet[%d] CK must be exactly 16 octets, got %d", idx, len(q.CK))
	}
	if len(q.IK) != 16 {
		return fmt.Errorf("sai: quintuplet[%d] IK must be exactly 16 octets, got %d", idx, len(q.IK))
	}
	if len(q.AUTN) != 16 {
		return fmt.Errorf("sai: quintuplet[%d] AUTN must be exactly 16 octets, got %d", idx, len(q.AUTN))
	}
	return nil
}

// validateEpcAV enforces the fixed-length requirements on an EPS
// authentication vector per 3GPP TS 29.272: RAND 16, XRES 4..16, AUTN 16,
// KASME 32.
func validateEpcAV(e *EpcAV, idx int) error {
	if len(e.RAND) != 16 {
		return fmt.Errorf("sai: epsAuthenticationSetList[%d] RAND must be exactly 16 octets, got %d", idx, len(e.RAND))
	}
	if len(e.XRES) < 4 || len(e.XRES) > 16 {
		return fmt.Errorf("sai: epsAuthenticationSetList[%d] XRES must be 4..16 octets, got %d", idx, len(e.XRES))
	}
	if len(e.AUTN) != 16 {
		return fmt.Errorf("sai: epsAuthenticationSetList[%d] AUTN must be exactly 16 octets, got %d", idx, len(e.AUTN))
	}
	if len(e.KASME) != 32 {
		return fmt.Errorf("sai: epsAuthenticationSetList[%d] KASME must be exactly 32 octets, got %d", idx, len(e.KASME))
	}
	return nil
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
		return nil, ErrSaiAuthSetListChoiceMultipleAlternatives
	}
	if !hasTriplets && !hasQuintuplets {
		return nil, ErrSaiAuthSetListChoiceNoAlternative
	}
	if hasTriplets {
		list := make(gsm_map.TripletList, len(a.Triplets))
		for i := range a.Triplets {
			if err := validateTriplet(&a.Triplets[i], i); err != nil {
				return nil, err
			}
			list[i] = gsm_map.AuthenticationTriplet{
				Rand: gsm_map.RAND(a.Triplets[i].RAND),
				Sres: gsm_map.SRES(a.Triplets[i].SRES),
				Kc:   gsm_map.Kc(a.Triplets[i].Kc),
			}
		}
		v := gsm_map.NewAuthenticationSetListTripletList(list)
		return &v, nil
	}
	list := make(gsm_map.QuintupletList, len(a.Quintuplets))
	for i := range a.Quintuplets {
		if err := validateQuintuplet(&a.Quintuplets[i], i); err != nil {
			return nil, err
		}
		list[i] = gsm_map.AuthenticationQuintuplet{
			Rand: gsm_map.RAND(a.Quintuplets[i].RAND),
			Xres: gsm_map.XRES(a.Quintuplets[i].XRES),
			Ck:   gsm_map.CK(a.Quintuplets[i].CK),
			Ik:   gsm_map.IK(a.Quintuplets[i].IK),
			Autn: gsm_map.AUTN(a.Quintuplets[i].AUTN),
		}
	}
	v := gsm_map.NewAuthenticationSetListQuintupletList(list)
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
		out := make([]AuthenticationTriplet, len(w.TripletList))
		for i, t := range w.TripletList {
			out[i] = AuthenticationTriplet{
				RAND: HexBytes(t.Rand),
				SRES: HexBytes(t.Sres),
				Kc:   HexBytes(t.Kc),
			}
			if err := validateTriplet(&out[i], i); err != nil {
				return nil, err
			}
		}
		return &AuthenticationSetList{Triplets: out}, nil
	case gsm_map.AuthenticationSetListChoiceQuintupletList:
		out := make([]AuthenticationQuintuplet, len(w.QuintupletList))
		for i, q := range w.QuintupletList {
			out[i] = AuthenticationQuintuplet{
				RAND: HexBytes(q.Rand),
				XRES: HexBytes(q.Xres),
				CK:   HexBytes(q.Ck),
				IK:   HexBytes(q.Ik),
				AUTN: HexBytes(q.Autn),
			}
			if err := validateQuintuplet(&out[i], i); err != nil {
				return nil, err
			}
		}
		return &AuthenticationSetList{Quintuplets: out}, nil
	default:
		return nil, fmt.Errorf("sai: unknown AuthenticationSetList CHOICE %d", w.Choice)
	}
}

// convertEpcAVToWire converts the public EpcAV into the wire-level gsm_map.EPCAV,
// enforcing RAND/XRES/AUTN/KASME size constraints per 3GPP TS 29.272.
func convertEpcAVToWire(e *EpcAV, idx int) (gsm_map.EPCAV, error) {
	if err := validateEpcAV(e, idx); err != nil {
		return gsm_map.EPCAV{}, err
	}
	return gsm_map.EPCAV{
		Rand:  gsm_map.RAND(e.RAND),
		Xres:  gsm_map.XRES(e.XRES),
		Autn:  gsm_map.AUTN(e.AUTN),
		Kasme: gsm_map.KASME(e.KASME),
	}, nil
}

// convertWireToEpcAV converts a wire-level gsm_map.EPCAV into the public
// EpcAV, enforcing the same size constraints symmetrically on decode.
func convertWireToEpcAV(w *gsm_map.EPCAV, idx int) (EpcAV, error) {
	out := EpcAV{
		RAND:  HexBytes(w.Rand),
		XRES:  HexBytes(w.Xres),
		AUTN:  HexBytes(w.Autn),
		KASME: HexBytes(w.Kasme),
	}
	if err := validateEpcAV(&out, idx); err != nil {
		return EpcAV{}, err
	}
	return out, nil
}

// convertSendAuthenticationInfoToArg converts the public SendAuthenticationInfo
// into the wire-level gsm_map.SendAuthenticationInfoArg.
func convertSendAuthenticationInfoToArg(s *SendAuthenticationInfo) (*gsm_map.SendAuthenticationInfoArg, error) {
	if s.IMSI == "" {
		return nil, ErrSaiMissingIMSI
	}
	if s.NumberOfRequestedVectors < 1 || s.NumberOfRequestedVectors > 5 {
		return nil, ErrSaiInvalidNumberOfRequestedVectors
	}
	if s.NumberOfRequestedAdditionalVectors != nil {
		v := *s.NumberOfRequestedAdditionalVectors
		if v < 1 || v > 5 {
			return nil, ErrSaiInvalidNumberOfRequestedAdditionalVectors
		}
	}
	if len(s.RequestingPLMNId) > 0 && len(s.RequestingPLMNId) != 3 {
		return nil, ErrSaiInvalidPLMNId
	}

	imsiBytes, err := tbcd.Encode(s.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	resync, err := convertReSynchronisationInfoToWire(s.ReSynchronisationInfo)
	if err != nil {
		return nil, err
	}

	arg := &gsm_map.SendAuthenticationInfoArg{
		Imsi:                         gsm_map.IMSI(imsiBytes),
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
		v := gsm_map.RequestingNodeType(*s.RequestingNodeType)
		arg.RequestingNodeType = &v
	}
	if len(s.RequestingPLMNId) > 0 {
		v := gsm_map.PLMNId(s.RequestingPLMNId)
		arg.RequestingPLMNId = &v
	}
	if s.NumberOfRequestedAdditionalVectors != nil {
		v := gsm_map.NumberOfRequestedVectors(int64(*s.NumberOfRequestedAdditionalVectors))
		arg.NumberOfRequestedAdditionalVectors = &v
	}

	return arg, nil
}

// convertArgToSendAuthenticationInfo converts a wire-level
// gsm_map.SendAuthenticationInfoArg back into the public SendAuthenticationInfo.
func convertArgToSendAuthenticationInfo(arg *gsm_map.SendAuthenticationInfoArg) (*SendAuthenticationInfo, error) {
	if len(arg.Imsi) == 0 {
		return nil, ErrSaiMissingIMSI
	}
	if arg.NumberOfRequestedVectors < 1 || arg.NumberOfRequestedVectors > 5 {
		return nil, ErrSaiInvalidNumberOfRequestedVectors
	}

	imsi, err := tbcd.Decode(arg.Imsi)
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
	//   "received values in the range (6-15) shall be treated as 'vlr'"
	//   "received values greater than 17 shall be treated as 'sgsn'"
	if arg.RequestingNodeType != nil {
		raw64 := int64(*arg.RequestingNodeType)
		if raw64 < 0 {
			return nil, fmt.Errorf("RequestingNodeType cannot be negative: %d", raw64)
		}
		raw, err := narrowInt64(raw64)
		if err != nil {
			return nil, fmt.Errorf("RequestingNodeType: %w", err)
		}
		v := RequestingNodeType(raw)
		switch {
		case raw >= 6 && raw <= 15:
			v = RequestingNodeVlr
		case raw > 17:
			v = RequestingNodeSgsn
		}
		out.RequestingNodeType = &v
	}
	if arg.RequestingPLMNId != nil {
		plmn := []byte(*arg.RequestingPLMNId)
		if len(plmn) != 3 {
			return nil, ErrSaiInvalidPLMNId
		}
		out.RequestingPLMNId = HexBytes(plmn)
	}
	if arg.NumberOfRequestedAdditionalVectors != nil {
		v := int64(*arg.NumberOfRequestedAdditionalVectors)
		if v < 1 || v > 5 {
			return nil, ErrSaiInvalidNumberOfRequestedAdditionalVectors
		}
		iv := int(v)
		out.NumberOfRequestedAdditionalVectors = &iv
	}

	return out, nil
}

// convertSendAuthenticationInfoResToRes converts the public
// SendAuthenticationInfoRes into the wire-level gsm_map.SendAuthenticationInfoRes.
func convertSendAuthenticationInfoResToRes(s *SendAuthenticationInfoRes) (*gsm_map.SendAuthenticationInfoRes, error) {
	if len(s.UeUsageType) > 0 && len(s.UeUsageType) != 4 {
		return nil, ErrSaiInvalidUeUsageType
	}

	res := &gsm_map.SendAuthenticationInfoRes{}

	if s.AuthenticationSetList != nil {
		asl, err := convertAuthenticationSetListToWire(s.AuthenticationSetList)
		if err != nil {
			return nil, err
		}
		res.AuthenticationSetList = asl
	}

	if len(s.EpsAuthenticationSetList) > 0 {
		if len(s.EpsAuthenticationSetList) > 5 {
			return nil, ErrSaiInvalidEpsAuthSetListSize
		}
		list := make(gsm_map.EPSAuthenticationSetList, len(s.EpsAuthenticationSetList))
		for i := range s.EpsAuthenticationSetList {
			av, err := convertEpcAVToWire(&s.EpsAuthenticationSetList[i], i)
			if err != nil {
				return nil, err
			}
			list[i] = av
		}
		res.EpsAuthenticationSetList = list
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

	if len(res.EpsAuthenticationSetList) > 0 {
		if len(res.EpsAuthenticationSetList) > 5 {
			return nil, ErrSaiInvalidEpsAuthSetListSize
		}
		list := make([]EpcAV, len(res.EpsAuthenticationSetList))
		for i := range res.EpsAuthenticationSetList {
			av, err := convertWireToEpcAV(&res.EpsAuthenticationSetList[i], i)
			if err != nil {
				return nil, err
			}
			list[i] = av
		}
		out.EpsAuthenticationSetList = list
	}

	if res.UeUsageType != nil {
		ue := []byte(*res.UeUsageType)
		if len(ue) != 4 {
			return nil, ErrSaiInvalidUeUsageType
		}
		out.UeUsageType = HexBytes(ue)
	}

	return out, nil
}
