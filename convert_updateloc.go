package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/gsn"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// --- UpdateLocation ---

func convertUpdateLocationToArg(u *UpdateLocation) (*gsm_map.UpdateLocationArg, error) {
	imsiBytes, err := tbcd.Encode(u.IMSI)
	if err != nil {
		return nil, fmt.Errorf(errEncodingIMSI, err)
	}

	mscNumber, err := encodeAddressField(u.MSCNumber, u.MSCNature, u.MSCPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding MSCNumber: %w", err)
	}

	vlrNumber, err := encodeAddressField(u.VLRNumber, u.VLRNature, u.VLRPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding VLRNumber: %w", err)
	}

	arg := &gsm_map.UpdateLocationArg{
		Imsi:      gsm_map.IMSI(imsiBytes),
		MscNumber: gsm_map.ISDNAddressString(mscNumber),
		VlrNumber: gsm_map.ISDNAddressString(vlrNumber),
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
		if len(u.LMSI) != 4 {
			return nil, fmt.Errorf("UpdateLocation: LMSI must be exactly 4 octets, got %d", len(u.LMSI))
		}
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
		v := gsm_map.GSNAddress(gsnAddr)
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
		pa := make(gsm_map.PagingArea, len(u.PagingArea))
		for i, raw := range u.PagingArea {
			// Each raw HexBytes is BER-encoded LocationArea CHOICE.
			var la gsm_map.LocationArea
			if err := la.UnmarshalBER(raw); err != nil {
				return nil, fmt.Errorf("PagingArea[%d]: %w", i, err)
			}
			pa[i] = la
		}
		arg.PagingArea = pa
	}

	arg.SkipSubscriberDataUpdate = boolToNullPtr(u.SkipSubscriberDataUpdate)
	arg.RestorationIndicator = boolToNullPtr(u.RestorationIndicator)

	if len(u.EplmnList) > 0 {
		list := make(gsm_map.EPLMNList, len(u.EplmnList))
		for i, raw := range u.EplmnList {
			if len(raw) != 3 {
				return nil, fmt.Errorf("UpdateLocation: EplmnList[%d] PLMNId must be exactly 3 octets, got %d", i, len(raw))
			}
			list[i] = gsm_map.PLMNId(raw)
		}
		arg.EplmnList = list
	}

	if u.MmeDiameterAddress != nil {
		arg.MmeDiameterAddress = convertNetworkNodeDiameterAddressToWire(u.MmeDiameterAddress)
	}

	return arg, nil
}

func convertArgToUpdateLocation(arg *gsm_map.UpdateLocationArg) (*UpdateLocation, error) {
	imsi, err := tbcd.Decode(arg.Imsi)
	if err != nil {
		return nil, fmt.Errorf("decoding IMSI: %w", err)
	}

	msc, mscNature, mscPlan, err := decodeAddressField(arg.MscNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding MSCNumber: %w", err)
	}

	vlr, vlrNature, vlrPlan, err := decodeAddressField(arg.VlrNumber)
	if err != nil {
		return nil, fmt.Errorf("decoding VLRNumber: %w", err)
	}

	u := &UpdateLocation{
		IMSI:      imsi,
		MSCNumber: msc,
		MSCNature: mscNature,
		MSCPlan:   mscPlan,
		VLRNumber: vlr,
		VLRNature: vlrNature,
		VLRPlan:   vlrPlan,
	}

	if arg.VlrCapability != nil {
		vlrCap := &VlrCapability{}

		if arg.VlrCapability.SupportedCamelPhases != nil && arg.VlrCapability.SupportedCamelPhases.BitLength > 0 {
			vlrCap.SupportedCamelPhases = convertBitStringToCamelPhases(*arg.VlrCapability.SupportedCamelPhases)
		}

		if arg.VlrCapability.SupportedLCSCapabilitySets != nil && arg.VlrCapability.SupportedLCSCapabilitySets.BitLength > 0 {
			vlrCap.SupportedLCSCapabilitySets = convertBitStringToLCSCaps(*arg.VlrCapability.SupportedLCSCapabilitySets)
		}

		vlrCap.SolsaSupportIndicator = nullPtrToBool(arg.VlrCapability.SolsaSupportIndicator)

		if arg.VlrCapability.IstSupportIndicator != nil {
			v := int(int64(*arg.VlrCapability.IstSupportIndicator))
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

		if arg.VlrCapability.OfferedCamel4CSIs != nil && arg.VlrCapability.OfferedCamel4CSIs.BitLength > 0 {
			vlrCap.OfferedCamel4CSIs = convertBitStringToOfferedCamel4CSIs(*arg.VlrCapability.OfferedCamel4CSIs)
		}

		if arg.VlrCapability.SupportedRATTypesIndicator != nil && arg.VlrCapability.SupportedRATTypesIndicator.BitLength > 0 {
			if arg.VlrCapability.SupportedRATTypesIndicator.BitLength < 2 || arg.VlrCapability.SupportedRATTypesIndicator.BitLength > 8 {
				return nil, fmt.Errorf("UpdateLocation: SupportedRATTypes BitLength must be 2..8, got %d", arg.VlrCapability.SupportedRATTypesIndicator.BitLength)
			}
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
		if len(*arg.Lmsi) != 4 {
			return nil, fmt.Errorf("UpdateLocation: LMSI must be exactly 4 octets, got %d", len(*arg.Lmsi))
		}
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

	if len(arg.PagingArea) > 0 {
		pa := make([]HexBytes, len(arg.PagingArea))
		for i, la := range arg.PagingArea {
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

	if len(arg.EplmnList) > 0 {
		list := make([]HexBytes, len(arg.EplmnList))
		for i, plmn := range arg.EplmnList {
			if len(plmn) != 3 {
				return nil, fmt.Errorf("UpdateLocation: EplmnList[%d] PLMNId must be exactly 3 octets, got %d", i, len(plmn))
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
	hlr, err := encodeAddressField(u.HLRNumber, u.HLRNumberNature, u.HLRNumberPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding HLRNumber: %w", err)
	}

	res := &gsm_map.UpdateLocationRes{
		HlrNumber:            gsm_map.ISDNAddressString(hlr),
		AddCapability:        boolToNullPtr(u.AddCapability),
		PagingAreaCapability: boolToNullPtr(u.PagingAreaCapability),
	}
	return res, nil
}

func convertResToUpdateLocationRes(res *gsm_map.UpdateLocationRes) (*UpdateLocationRes, error) {
	hlr, nature, plan, err := decodeAddressField(res.HlrNumber)
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
