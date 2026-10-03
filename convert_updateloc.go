package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1-gsmmap/gsn"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// --- UpdateLocation ---

func convertUpdateLocationToArg(u *UpdateLocation) (*gsm_map.UpdateLocationArg, error) {
	// imsi, msc-Number and vlr-Number are mandatory in UpdateLocationArg
	// (3GPP TS 29.002 V19.1.0 §17.7.1). An empty IMSI fails in
	// encodeIdentityDigits with ErrIdentityEmpty.
	if u.MscNumber == "" {
		return nil, ErrUpdateLocationMissingMscNumber
	}
	if u.VlrNumber == "" {
		return nil, ErrUpdateLocationMissingVlrNumber
	}

	imsiBytes, err := encodeIdentityDigits(identityIMSI, u.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	mscNumber, err := encodeAddressField(u.MscNumber, u.MscNumberNature, u.MscNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding MscNumber: %w", err)
	}

	vlrNumber, err := encodeAddressField(u.VlrNumber, u.VlrNumberNature, u.VlrNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding VlrNumber: %w", err)
	}

	arg := &gsm_map.UpdateLocationArg{
		Imsi:      imsiBytes,
		MscNumber: mscNumber,
		VlrNumber: vlrNumber,
	}

	if u.VlrCapability != nil {
		vlrCap := &gsm_map.VLRCapability{}

		if u.VlrCapability.SupportedCamelPhases != nil {
			bs := convertCamelPhasesToBitString(u.VlrCapability.SupportedCamelPhases)
			vlrCap.SupportedCamelPhases = &bs
		}

		if u.VlrCapability.SupportedLCSCapabilitySets != nil {
			bs := convertLCSCapsToBitString(u.VlrCapability.SupportedLCSCapabilitySets)
			vlrCap.SupportedLCSCapabilitySets = &bs
		}

		vlrCap.SolsaSupportIndicator = boolToNullPtr(u.VlrCapability.SolsaSupportIndicator)

		if u.VlrCapability.IstSupportIndicator != nil {
			// Sender accepts defined values; receivers map values above 1 to istCommandSupported (3GPP TS 29.002 V19.1.0 §17.7.1).
			if *u.VlrCapability.IstSupportIndicator < 0 || *u.VlrCapability.IstSupportIndicator > 1 {
				return nil, fmt.Errorf("VlrCapability.IstSupportIndicator: %w (got %d)", ErrISTSupportIndicatorInvalid, *u.VlrCapability.IstSupportIndicator)
			}
			v := gsm_map.ISTSupportIndicator(int64(*u.VlrCapability.IstSupportIndicator))
			vlrCap.IstSupportIndicator = &v
		}

		if u.VlrCapability.SuperChargerSupportedInServingNetworkEntity != nil {
			sc, err := convertSuperChargerInfoToWire(u.VlrCapability.SuperChargerSupportedInServingNetworkEntity)
			if err != nil {
				return nil, fmt.Errorf("SuperChargerInfo: %w", err)
			}
			vlrCap.SuperChargerSupportedInServingNetworkEntity = sc
		}

		vlrCap.LongFTNSupported = boolToNullPtr(u.VlrCapability.LongFTNSupported)

		if u.VlrCapability.OfferedCamel4CSIs != nil {
			bs := convertOfferedCamel4CSIsToBitString(u.VlrCapability.OfferedCamel4CSIs)
			vlrCap.OfferedCamel4CSIs = &bs
		}

		if u.VlrCapability.SupportedRATTypesIndicator != nil {
			bs := convertSupportedRATTypesToBitString(u.VlrCapability.SupportedRATTypesIndicator)
			vlrCap.SupportedRATTypesIndicator = &bs
		}

		vlrCap.LongGroupIDSupported = boolToNullPtr(u.VlrCapability.LongGroupIDSupported)
		vlrCap.MtRoamingForwardingSupported = boolToNullPtr(u.VlrCapability.MtRoamingForwardingSupported)
		vlrCap.MsisdnLessOperationSupported = boolToNullPtr(u.VlrCapability.MsisdnLessOperationSupported)
		vlrCap.ResetIdsSupported = boolToNullPtr(u.VlrCapability.ResetIdsSupported)

		arg.VlrCapability = vlrCap
	}

	// Optional fields.
	if len(u.LMSI) > 0 {
		v := gsm_map.LMSI(u.LMSI)
		arg.Lmsi = &v
	}

	arg.InformPreviousNetworkEntity = boolToNullPtr(u.InformPreviousNetworkEntity)
	arg.CsLCSNotSupportedByUE = boolToNullPtr(u.CsLCSNotSupportedByUE)

	if u.VGmlcAddress != "" {
		gsnAddr, err := gsn.Build(u.VGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("encoding VGmlcAddress: %w", err)
		}
		v := gsnAddr
		arg.VGmlcAddress = &v
	}

	if u.AddInfo != nil {
		ai, err := convertAddInfoToWire(u.AddInfo)
		if err != nil {
			return nil, fmt.Errorf("AddInfo: %w", err)
		}
		arg.AddInfo = ai
	}

	if len(u.PagingArea) > 0 {
		pa := gsm_map.PagingArea{Values: make([]gsm_map.LocationArea, len(u.PagingArea))}
		for i, raw := range u.PagingArea {
			// Each raw HexBytes is BER-encoded LocationArea CHOICE.
			var la gsm_map.LocationArea
			if err := la.UnmarshalBER(raw); err != nil {
				return nil, fmt.Errorf("PagingArea[%d]: %w", i, err)
			}
			pa.Values[i] = la
		}
		arg.PagingArea = &pa
	}

	arg.SkipSubscriberDataUpdate = boolToNullPtr(u.SkipSubscriberDataUpdate)
	arg.RestorationIndicator = boolToNullPtr(u.RestorationIndicator)

	if len(u.EplmnList) > 0 {
		list := gsm_map.EPLMNList{Values: make([]gsm_map.PLMNId, len(u.EplmnList))}
		for i, raw := range u.EplmnList {
			// go-asn1 does not enforce SEQUENCE OF element SIZE: https://github.com/gomaja/go-asn1/issues/79.
			if len(raw) != 3 {
				return nil, fmt.Errorf("UpdateLocation: EplmnList[%d] length %d: %w", i, len(raw), ErrPLMNIdInvalidLength)
			}
			list.Values[i] = gsm_map.PLMNId(raw)
		}
		arg.EplmnList = &list
	}

	if u.MmeDiameterAddress != nil {
		arg.MmeDiameterAddress = convertNetworkNodeDiameterAddressToWire(u.MmeDiameterAddress)
	}

	return arg, nil
}

func convertArgToUpdateLocation(arg *gsm_map.UpdateLocationArg) (*UpdateLocation, error) {
	imsi, err := decodeIdentityDigits(identityIMSI, arg.Imsi)
	if err != nil {
		return nil, fmt.Errorf("decoding IMSI: %w", err)
	}

	msc, mscNature, mscPlan, err := decodeAddressWithDigits(arg.MscNumber, ErrUpdateLocationMissingMscNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding MscNumber: %w", err)
	}

	vlr, vlrNature, vlrPlan, err := decodeAddressWithDigits(arg.VlrNumber, ErrUpdateLocationMissingVlrNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding VlrNumber: %w", err)
	}

	u := &UpdateLocation{
		IMSI:            imsi,
		MscNumber:       msc,
		MscNumberNature: mscNature,
		MscNumberPlan:   mscPlan,
		VlrNumber:       vlr,
		VlrNumberNature: vlrNature,
		VlrNumberPlan:   vlrPlan,
	}

	if arg.VlrCapability != nil {
		vlrCap := &VlrCapability{}

		if arg.VlrCapability.SupportedCamelPhases != nil {
			vlrCap.SupportedCamelPhases = convertBitStringToCamelPhases(*arg.VlrCapability.SupportedCamelPhases)
		}

		if arg.VlrCapability.SupportedLCSCapabilitySets != nil {
			vlrCap.SupportedLCSCapabilitySets = convertBitStringToLCSCaps(*arg.VlrCapability.SupportedLCSCapabilitySets)
		}

		vlrCap.SolsaSupportIndicator = nullPtrToBool(arg.VlrCapability.SolsaSupportIndicator)

		// IstSupportIndicator — ENUMERATED { basicISTSupported(0),
		// istCommandSupported(1), ... } per TS 29.002. Spec exception:
		// "reception of values > 1 shall be mapped to 'istCommandSupported'".
		// Apply the mapping in int64 space first so wire values that exceed
		// platform int satisfy the spec mandate on 32-bit builds.
		if arg.VlrCapability.IstSupportIndicator != nil {
			v, err := istSupportIndicatorFromWire(*arg.VlrCapability.IstSupportIndicator)
			if err != nil {
				return nil, fmt.Errorf("VlrCapability.IstSupportIndicator: %w", err)
			}
			vlrCap.IstSupportIndicator = &v
		}

		if arg.VlrCapability.SuperChargerSupportedInServingNetworkEntity != nil {
			sc, err := convertWireToSuperChargerInfo(arg.VlrCapability.SuperChargerSupportedInServingNetworkEntity)
			if err != nil {
				return nil, fmt.Errorf("SuperChargerInfo: %w", err)
			}
			vlrCap.SuperChargerSupportedInServingNetworkEntity = sc
		}

		vlrCap.LongFTNSupported = nullPtrToBool(arg.VlrCapability.LongFTNSupported)

		if arg.VlrCapability.OfferedCamel4CSIs != nil {
			vlrCap.OfferedCamel4CSIs = convertBitStringToOfferedCamel4CSIs(*arg.VlrCapability.OfferedCamel4CSIs)
		}

		if arg.VlrCapability.SupportedRATTypesIndicator != nil {
			vlrCap.SupportedRATTypesIndicator = convertBitStringToSupportedRATTypes(*arg.VlrCapability.SupportedRATTypesIndicator)
		}

		vlrCap.LongGroupIDSupported = nullPtrToBool(arg.VlrCapability.LongGroupIDSupported)
		vlrCap.MtRoamingForwardingSupported = nullPtrToBool(arg.VlrCapability.MtRoamingForwardingSupported)
		vlrCap.MsisdnLessOperationSupported = nullPtrToBool(arg.VlrCapability.MsisdnLessOperationSupported)
		vlrCap.ResetIdsSupported = nullPtrToBool(arg.VlrCapability.ResetIdsSupported)

		u.VlrCapability = vlrCap
	}

	// Optional fields.
	if arg.Lmsi != nil {
		u.LMSI = HexBytes(*arg.Lmsi)
	}

	u.InformPreviousNetworkEntity = nullPtrToBool(arg.InformPreviousNetworkEntity)
	u.CsLCSNotSupportedByUE = nullPtrToBool(arg.CsLCSNotSupportedByUE)

	if arg.VGmlcAddress != nil {
		addr, err := gsn.Parse(*arg.VGmlcAddress)
		if err != nil {
			return nil, fmt.Errorf("decoding VGmlcAddress: %w", err)
		}
		u.VGmlcAddress = addr
	}

	if arg.AddInfo != nil {
		ai, err := convertWireToAddInfo(arg.AddInfo)
		if err != nil {
			return nil, fmt.Errorf("AddInfo: %w", err)
		}
		u.AddInfo = ai
	}

	if arg.PagingArea != nil && len(arg.PagingArea.Values) > 0 {
		pa := make([]HexBytes, len(arg.PagingArea.Values))
		for i, la := range arg.PagingArea.Values {
			encoded, err := la.MarshalDER()
			if err != nil {
				return nil, fmt.Errorf("PagingArea[%d]: %w", i, err)
			}
			pa[i] = HexBytes(encoded)
		}
		u.PagingArea = pa
	}

	u.SkipSubscriberDataUpdate = nullPtrToBool(arg.SkipSubscriberDataUpdate)
	u.RestorationIndicator = nullPtrToBool(arg.RestorationIndicator)

	if arg.EplmnList != nil && len(arg.EplmnList.Values) > 0 {
		list := make([]HexBytes, len(arg.EplmnList.Values))
		for i, plmn := range arg.EplmnList.Values {
			// go-asn1 does not enforce SEQUENCE OF element SIZE: https://github.com/gomaja/go-asn1/issues/79.
			if len(plmn) != 3 {
				return nil, fmt.Errorf("UpdateLocation: EplmnList[%d] length %d: %w", i, len(plmn), ErrPLMNIdInvalidLength)
			}
			list[i] = HexBytes(plmn)
		}
		u.EplmnList = list
	}

	if arg.MmeDiameterAddress != nil {
		u.MmeDiameterAddress = convertWireToNetworkNodeDiameterAddress(arg.MmeDiameterAddress)
	}

	return u, nil
}

// --- UpdateLocationRes ---

func convertUpdateLocationResToRes(u *UpdateLocationRes) (*gsm_map.UpdateLocationRes, error) {
	if u.HLRNumber == "" {
		return nil, ErrUpdateLocationResMissingHLRNumber
	}
	hlr, err := encodeAddressField(u.HLRNumber, u.HLRNumberNature, u.HLRNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding HLRNumber: %w", err)
	}

	res := &gsm_map.UpdateLocationRes{
		HlrNumber:            hlr,
		AddCapability:        boolToNullPtr(u.AddCapability),
		PagingAreaCapability: boolToNullPtr(u.PagingAreaCapability),
	}
	return res, nil
}

func convertResToUpdateLocationRes(res *gsm_map.UpdateLocationRes) (*UpdateLocationRes, error) {
	hlr, nature, plan, err := decodeAddressWithDigits(res.HlrNumber, ErrUpdateLocationResMissingHLRNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding HLRNumber: %w", err)
	}

	return &UpdateLocationRes{
		HLRNumber:            hlr,
		HLRNumberNature:      nature,
		HLRNumberPlan:        plan,
		AddCapability:        nullPtrToBool(res.AddCapability),
		PagingAreaCapability: nullPtrToBool(res.PagingAreaCapability),
	}, nil
}
