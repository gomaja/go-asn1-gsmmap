package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// --- CS Location conversion ---

func convertCSLocationToAsn1(loc *CSLocationInformation) (*gsm_map.LocationInformation, error) {
	li := &gsm_map.LocationInformation{}

	if loc.AgeOfLocationInformation != nil {
		age := gsm_map.AgeOfLocationInformation(*loc.AgeOfLocationInformation)
		li.AgeOfLocationInformation = &age
	}

	if loc.VlrNumber != "" {
		vlr, err := encodeAddressField(loc.VlrNumber, loc.VlrNumberNature, loc.VlrNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding VlrNumber: %w", err)
		}
		vlrAddr := gsm_map.ISDNAddressString(vlr)
		li.VlrNumber = &vlrAddr
	}

	if loc.MscNumber != "" {
		msc, err := encodeAddressField(loc.MscNumber, loc.MscNumberNature, loc.MscNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding MscNumber: %w", err)
		}
		mscAddr := gsm_map.ISDNAddressString(msc)
		li.MscNumber = &mscAddr
	}

	if loc.GeographicalInformation != nil {
		raw, err := loc.GeographicalInformation.Encode()
		if err != nil {
			return nil, fmt.Errorf("encoding GeographicalInformation: %w", err)
		}
		gi := gsm_map.GeographicalInformation(raw)
		li.GeographicalInformation = &gi
	}

	if loc.GeodeticInformation != nil {
		gd := gsm_map.GeodeticInformation(loc.GeodeticInformation)
		li.GeodeticInformation = &gd
	}

	if loc.CellGlobalId != nil {
		v := gsm_map.NewCellGlobalIdOrServiceAreaIdOrLAICellGlobalIdOrServiceAreaIdFixedLength(
			gsm_map.CellGlobalIdOrServiceAreaIdFixedLength(loc.CellGlobalId),
		)
		li.CellGlobalIdOrServiceAreaIdOrLAI = &v
	} else if loc.LAI != nil {
		v := gsm_map.NewCellGlobalIdOrServiceAreaIdOrLAILaiFixedLength(
			gsm_map.LAIFixedLength(loc.LAI),
		)
		li.CellGlobalIdOrServiceAreaIdOrLAI = &v
	}

	if loc.LocationNumber != nil {
		ln := gsm_map.LocationNumber(loc.LocationNumber)
		li.LocationNumber = &ln
	}

	if loc.SelectedLSAId != nil {
		lsa := gsm_map.LSAIdentity(loc.SelectedLSAId)
		li.SelectedLSAId = &lsa
	}

	if loc.UserCSGInformation != nil {
		csg, err := convertUserCSGInformationToWire(loc.UserCSGInformation)
		if err != nil {
			return nil, fmt.Errorf("UserCSGInformation: %w", err)
		}
		li.UserCSGInformation = csg
	}

	if loc.CurrentLocationRetrieved {
		li.CurrentLocationRetrieved = &struct{}{}
	}

	if loc.SAIPresent {
		li.SaiPresent = &struct{}{}
	}

	return li, nil
}

func convertAsn1ToCSLocation(li *gsm_map.LocationInformation) (*CSLocationInformation, error) {
	loc := &CSLocationInformation{}

	if li.AgeOfLocationInformation != nil {
		v := int(*li.AgeOfLocationInformation)
		loc.AgeOfLocationInformation = &v
	}

	if li.VlrNumber != nil {
		vlr, nature, plan, err := decodeAddressField(*li.VlrNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding VlrNumber: %w", err)
		}
		loc.VlrNumber = vlr
		loc.VlrNumberNature = nature
		loc.VlrNumberPlan = plan
	}

	if li.MscNumber != nil {
		msc, nature, plan, err := decodeAddressField(*li.MscNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding MscNumber: %w", err)
		}
		loc.MscNumber = msc
		loc.MscNumberNature = nature
		loc.MscNumberPlan = plan
	}

	if li.GeographicalInformation != nil {
		gi, err := DecodeGeographicalInfo([]byte(*li.GeographicalInformation))
		if err != nil {
			return nil, fmt.Errorf("decoding GeographicalInformation: %w", err)
		}
		loc.GeographicalInformation = gi
	}

	if li.GeodeticInformation != nil {
		loc.GeodeticInformation = []byte(*li.GeodeticInformation)
	}

	if li.CellGlobalIdOrServiceAreaIdOrLAI != nil {
		choice := li.CellGlobalIdOrServiceAreaIdOrLAI
		switch choice.Choice {
		case gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceCellGlobalIdOrServiceAreaIdFixedLength:
			if choice.CellGlobalIdOrServiceAreaIdFixedLength == nil {
				return nil, fmt.Errorf("CellGlobalIdOrServiceAreaIdOrLAI: cellGlobalId alternative selected but payload is nil")
			}
			b := []byte(*choice.CellGlobalIdOrServiceAreaIdFixedLength)
			if len(b) != 7 {
				return nil, fmt.Errorf("CellGlobalId must be exactly 7 octets, got %d", len(b))
			}
			loc.CellGlobalId = b
		case gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceLaiFixedLength:
			if choice.LaiFixedLength == nil {
				return nil, fmt.Errorf("CellGlobalIdOrServiceAreaIdOrLAI: LAI alternative selected but payload is nil")
			}
			b := []byte(*choice.LaiFixedLength)
			if len(b) != 5 {
				return nil, fmt.Errorf("LAI must be exactly 5 octets, got %d", len(b))
			}
			loc.LAI = b
		default:
			return nil, fmt.Errorf("CellGlobalIdOrServiceAreaIdOrLAI: unknown CHOICE %d", choice.Choice)
		}
	}

	if li.LocationNumber != nil {
		loc.LocationNumber = []byte(*li.LocationNumber)
	}

	if li.SelectedLSAId != nil {
		loc.SelectedLSAId = []byte(*li.SelectedLSAId)
	}

	if li.UserCSGInformation != nil {
		loc.UserCSGInformation = convertWireToUserCSGInformation(li.UserCSGInformation)
	}

	loc.CurrentLocationRetrieved = li.CurrentLocationRetrieved != nil
	loc.SAIPresent = li.SaiPresent != nil

	return loc, nil
}

// --- SubscriberState conversion ---

// convertSubscriberStateToAsn1 encodes the public SubscriberStateInfo into
// a wire SubscriberState CHOICE. Unknown State values return an error so
// an out-of-spec caller can't silently produce a zero-valued wire value
// that re-decodes as assumedIdle.
func convertSubscriberStateToAsn1(ss *SubscriberStateInfo) (*gsm_map.SubscriberState, error) {
	var s gsm_map.SubscriberState
	switch ss.State {
	case StateAssumedIdle:
		s = gsm_map.NewSubscriberStateAssumedIdle(struct{}{})
	case StateCamelBusy:
		s = gsm_map.NewSubscriberStateCamelBusy(struct{}{})
	case StateNetDetNotReachable:
		// NotReachableReason is mandatory when this alternative is chosen;
		// reject nil rather than silently encoding msPurged(0).
		if ss.NotReachableReason == nil {
			return nil, fmt.Errorf("SubscriberState: StateNetDetNotReachable requires a non-nil NotReachableReason")
		}
		if *ss.NotReachableReason < 0 || *ss.NotReachableReason > 3 {
			return nil, fmt.Errorf("SubscriberState.NotReachableReason out of range 0..3: %d", *ss.NotReachableReason)
		}
		s = gsm_map.NewSubscriberStateNetDetNotReachable(gsm_map.NotReachableReason(*ss.NotReachableReason))
	case StateNotProvidedFromVLR:
		s = gsm_map.NewSubscriberStateNotProvidedFromVLR(struct{}{})
	default:
		return nil, fmt.Errorf("SubscriberState: unknown State %d", ss.State)
	}
	return &s, nil
}

// convertAsn1ToSubscriberState decodes the wire SubscriberState CHOICE into
// the public type, rejecting unknown CHOICE values rather than silently
// emitting a zero-valued state.
func convertAsn1ToSubscriberState(ss *gsm_map.SubscriberState) (*SubscriberStateInfo, error) {
	info := &SubscriberStateInfo{}
	switch ss.Choice {
	case gsm_map.SubscriberStateChoiceAssumedIdle:
		info.State = StateAssumedIdle
	case gsm_map.SubscriberStateChoiceCamelBusy:
		info.State = StateCamelBusy
	case gsm_map.SubscriberStateChoiceNetDetNotReachable:
		info.State = StateNetDetNotReachable
		if ss.NetDetNotReachable == nil {
			return nil, fmt.Errorf("SubscriberState: netDetNotReachable alternative selected but reason is nil")
		}
		reason, err := narrowInt64Range(int64(*ss.NetDetNotReachable), 0, 3, "NotReachableReason")
		if err != nil {
			return nil, err
		}
		info.NotReachableReason = &reason
	case gsm_map.SubscriberStateChoiceNotProvidedFromVLR:
		info.State = StateNotProvidedFromVLR
	default:
		return nil, fmt.Errorf("SubscriberState: unknown CHOICE %d", ss.Choice)
	}
	return info, nil
}

// --- EPS Location conversion ---

func convertEPSLocationToAsn1(loc *EPSLocationInformation) (*gsm_map.LocationInformationEPS, error) {
	li := &gsm_map.LocationInformationEPS{}

	if loc.AgeOfLocationInformation != nil {
		age := gsm_map.AgeOfLocationInformation(*loc.AgeOfLocationInformation)
		li.AgeOfLocationInformation = &age
	}

	if loc.EUtranCellGlobalIdentity != nil {
		cgi := gsm_map.EUTRANCGI(loc.EUtranCellGlobalIdentity)
		li.EUtranCellGlobalIdentity = &cgi
	}

	if loc.TrackingAreaIdentity != nil {
		ta := gsm_map.TAId(loc.TrackingAreaIdentity)
		li.TrackingAreaIdentity = &ta
	}

	if loc.GeographicalInformation != nil {
		raw, err := loc.GeographicalInformation.Encode()
		if err != nil {
			return nil, fmt.Errorf("encoding GeographicalInformation: %w", err)
		}
		gi := gsm_map.GeographicalInformation(raw)
		li.GeographicalInformation = &gi
	}

	if loc.GeodeticInformation != nil {
		gd := gsm_map.GeodeticInformation(loc.GeodeticInformation)
		li.GeodeticInformation = &gd
	}

	if loc.CurrentLocationRetrieved {
		li.CurrentLocationRetrieved = &struct{}{}
	}

	if loc.MmeName != nil {
		mm := gsm_map.DiameterIdentity(loc.MmeName)
		li.MmeName = &mm
	}

	return li, nil
}

func convertAsn1ToEPSLocation(li *gsm_map.LocationInformationEPS) (*EPSLocationInformation, error) {
	loc := &EPSLocationInformation{}

	if li.AgeOfLocationInformation != nil {
		v := int(*li.AgeOfLocationInformation)
		loc.AgeOfLocationInformation = &v
	}

	if li.EUtranCellGlobalIdentity != nil {
		b := []byte(*li.EUtranCellGlobalIdentity)
		if len(b) != 7 {
			return nil, fmt.Errorf("EUtranCellGlobalIdentity must be exactly 7 octets, got %d", len(b))
		}
		loc.EUtranCellGlobalIdentity = b
	}

	if li.TrackingAreaIdentity != nil {
		b := []byte(*li.TrackingAreaIdentity)
		if len(b) != 5 {
			return nil, fmt.Errorf("TrackingAreaIdentity must be exactly 5 octets, got %d", len(b))
		}
		loc.TrackingAreaIdentity = b
	}

	if li.GeographicalInformation != nil {
		gi, err := DecodeGeographicalInfo([]byte(*li.GeographicalInformation))
		if err != nil {
			return nil, fmt.Errorf("decoding GeographicalInformation: %w", err)
		}
		loc.GeographicalInformation = gi
	}

	if li.GeodeticInformation != nil {
		loc.GeodeticInformation = []byte(*li.GeodeticInformation)
	}

	loc.CurrentLocationRetrieved = li.CurrentLocationRetrieved != nil

	if li.MmeName != nil {
		loc.MmeName = []byte(*li.MmeName)
	}

	return loc, nil
}

// --- GPRS Location conversion ---

func convertGPRSLocationToAsn1(loc *GPRSLocationInformation) (*gsm_map.LocationInformationGPRS, error) {
	li := &gsm_map.LocationInformationGPRS{}

	if loc.AgeOfLocationInformation != nil {
		age := gsm_map.AgeOfLocationInformation(*loc.AgeOfLocationInformation)
		li.AgeOfLocationInformation = &age
	}

	if loc.CellGlobalId != nil {
		v := gsm_map.NewCellGlobalIdOrServiceAreaIdOrLAICellGlobalIdOrServiceAreaIdFixedLength(
			gsm_map.CellGlobalIdOrServiceAreaIdFixedLength(loc.CellGlobalId),
		)
		li.CellGlobalIdOrServiceAreaIdOrLAI = &v
	} else if loc.LAI != nil {
		v := gsm_map.NewCellGlobalIdOrServiceAreaIdOrLAILaiFixedLength(
			gsm_map.LAIFixedLength(loc.LAI),
		)
		li.CellGlobalIdOrServiceAreaIdOrLAI = &v
	}

	if loc.RouteingAreaIdentity != nil {
		ra := gsm_map.RAIdentity(loc.RouteingAreaIdentity)
		li.RouteingAreaIdentity = &ra
	}

	if loc.GeographicalInformation != nil {
		raw, err := loc.GeographicalInformation.Encode()
		if err != nil {
			return nil, fmt.Errorf("encoding GeographicalInformation: %w", err)
		}
		gi := gsm_map.GeographicalInformation(raw)
		li.GeographicalInformation = &gi
	}

	if loc.GeodeticInformation != nil {
		gd := gsm_map.GeodeticInformation(loc.GeodeticInformation)
		li.GeodeticInformation = &gd
	}

	if loc.SgsnNumber != "" {
		sgsn, err := encodeAddressField(loc.SgsnNumber, loc.SgsnNumberNature, loc.SgsnNumberPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding SgsnNumber: %w", err)
		}
		sgsnAddr := gsm_map.ISDNAddressString(sgsn)
		li.SgsnNumber = &sgsnAddr
	}

	if loc.SelectedLSAIdentity != nil {
		lsa := gsm_map.LSAIdentity(loc.SelectedLSAIdentity)
		li.SelectedLSAIdentity = &lsa
	}

	if loc.UserCSGInformation != nil {
		csg, err := convertUserCSGInformationToWire(loc.UserCSGInformation)
		if err != nil {
			return nil, fmt.Errorf("UserCSGInformation: %w", err)
		}
		li.UserCSGInformation = csg
	}

	if loc.CurrentLocationRetrieved {
		li.CurrentLocationRetrieved = &struct{}{}
	}

	if loc.SAIPresent {
		li.SaiPresent = &struct{}{}
	}

	return li, nil
}

func convertAsn1ToGPRSLocation(li *gsm_map.LocationInformationGPRS) (*GPRSLocationInformation, error) {
	loc := &GPRSLocationInformation{}

	if li.AgeOfLocationInformation != nil {
		v := int(*li.AgeOfLocationInformation)
		loc.AgeOfLocationInformation = &v
	}

	if li.CellGlobalIdOrServiceAreaIdOrLAI != nil {
		choice := li.CellGlobalIdOrServiceAreaIdOrLAI
		switch choice.Choice {
		case gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceCellGlobalIdOrServiceAreaIdFixedLength:
			if choice.CellGlobalIdOrServiceAreaIdFixedLength == nil {
				return nil, fmt.Errorf("CellGlobalIdOrServiceAreaIdOrLAI: cellGlobalId alternative selected but payload is nil")
			}
			b := []byte(*choice.CellGlobalIdOrServiceAreaIdFixedLength)
			if len(b) != 7 {
				return nil, fmt.Errorf("CellGlobalId must be exactly 7 octets, got %d", len(b))
			}
			loc.CellGlobalId = b
		case gsm_map.CellGlobalIdOrServiceAreaIdOrLAIChoiceLaiFixedLength:
			if choice.LaiFixedLength == nil {
				return nil, fmt.Errorf("CellGlobalIdOrServiceAreaIdOrLAI: LAI alternative selected but payload is nil")
			}
			b := []byte(*choice.LaiFixedLength)
			if len(b) != 5 {
				return nil, fmt.Errorf("LAI must be exactly 5 octets, got %d", len(b))
			}
			loc.LAI = b
		default:
			return nil, fmt.Errorf("CellGlobalIdOrServiceAreaIdOrLAI: unknown CHOICE %d", choice.Choice)
		}
	}

	if li.RouteingAreaIdentity != nil {
		b := []byte(*li.RouteingAreaIdentity)
		if len(b) != 6 {
			return nil, fmt.Errorf("RouteingAreaIdentity must be exactly 6 octets, got %d", len(b))
		}
		loc.RouteingAreaIdentity = b
	}

	if li.GeographicalInformation != nil {
		gi, err := DecodeGeographicalInfo([]byte(*li.GeographicalInformation))
		if err != nil {
			return nil, fmt.Errorf("decoding GeographicalInformation: %w", err)
		}
		loc.GeographicalInformation = gi
	}

	if li.GeodeticInformation != nil {
		loc.GeodeticInformation = []byte(*li.GeodeticInformation)
	}

	if li.SgsnNumber != nil {
		sgsn, nature, plan, err := decodeAddressField(*li.SgsnNumber)
		if err != nil {
			return nil, fmt.Errorf("decoding SgsnNumber: %w", err)
		}
		loc.SgsnNumber = sgsn
		loc.SgsnNumberNature = nature
		loc.SgsnNumberPlan = plan
	}

	if li.SelectedLSAIdentity != nil {
		loc.SelectedLSAIdentity = []byte(*li.SelectedLSAIdentity)
	}

	if li.UserCSGInformation != nil {
		loc.UserCSGInformation = convertWireToUserCSGInformation(li.UserCSGInformation)
	}

	loc.CurrentLocationRetrieved = li.CurrentLocationRetrieved != nil
	loc.SAIPresent = li.SaiPresent != nil

	return loc, nil
}
