package gsmmap

import (
	"fmt"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
)

func convertCheckIMEIArgToWire(v *CheckIMEIArg) (*gsm_map.CheckIMEIArg, error) {
	if v == nil {
		return nil, ErrCheckIMEIArgNil
	}
	imei, err := encodeIdentityDigits(identityIMEI, v.IMEI)
	if err != nil {
		return nil, fmt.Errorf("IMEI: %w", err)
	}
	// 3GPP TS 29.002 V19.1.0 §8.7.1.3: IMEISV shall be present if BMUEF is requested.
	if v.RequestedEquipmentInfo.BMUEF && len(v.IMEI) < 16 {
		return nil, ErrCheckIMEIArgBMUEFRequiresIMEISV
	}
	// 3GPP TS 29.002 V19.1.0 §8.7.1.3: Equipment Status or BMUEF or both is requested.
	if !v.RequestedEquipmentInfo.EquipmentStatus && !v.RequestedEquipmentInfo.BMUEF {
		return nil, ErrCheckIMEIArgNoEquipmentInfoRequested
	}
	info, err := convertRequestedEquipmentInfoToBitString(&v.RequestedEquipmentInfo)
	if err != nil {
		return nil, err
	}
	return &gsm_map.CheckIMEIArg{Imei: imei, RequestedEquipmentInfo: info}, nil
}

func convertWireToCheckIMEIArg(w *gsm_map.CheckIMEIArg) (*CheckIMEIArg, error) {
	imei, err := decodeIdentityDigits(identityIMEI, w.Imei)
	if err != nil {
		return nil, fmt.Errorf("IMEI: %w", err)
	}
	return &CheckIMEIArg{IMEI: imei, RequestedEquipmentInfo: *convertBitStringToRequestedEquipmentInfo(w.RequestedEquipmentInfo)}, nil
}

func convertRequestedEquipmentInfoToBitString(v *RequestedEquipmentInfo) (runtime.BitString, error) {
	bits := []bool{gsm_map.RequestedEquipmentInfoEquipmentStatus: v.EquipmentStatus, gsm_map.RequestedEquipmentInfoBmuef: v.BMUEF}
	return gsm_map.NewRequestedEquipmentInfo(namedBitPositions(bits)...)
}

func convertBitStringToRequestedEquipmentInfo(w runtime.BitString) *RequestedEquipmentInfo {
	// 3GPP TS 29.002 V19.1.0 §17.7.1: discard unknown bit assignments.
	return &RequestedEquipmentInfo{EquipmentStatus: w.Has(gsm_map.RequestedEquipmentInfoEquipmentStatus), BMUEF: w.Has(gsm_map.RequestedEquipmentInfoBmuef)}
}

func convertCheckIMEIResToWire(v *CheckIMEIRes) (*gsm_map.CheckIMEIRes, error) {
	if v == nil {
		return nil, ErrCheckIMEIResNil
	}
	w := &gsm_map.CheckIMEIRes{EquipmentStatus: v.EquipmentStatus}
	if v.BMUEF != nil {
		w.Bmuef = &gsm_map.UESBIIu{}
		if len(v.BMUEF.UESBIIuA) > 0 || v.BMUEF.UESBIIuABits != 0 {
			b := bitStringToWire(v.BMUEF.UESBIIuA, v.BMUEF.UESBIIuABits)
			w.Bmuef.UesbiIuA = &b
		}
		if len(v.BMUEF.UESBIIuB) > 0 || v.BMUEF.UESBIIuBBits != 0 {
			b := bitStringToWire(v.BMUEF.UESBIIuB, v.BMUEF.UESBIIuBBits)
			w.Bmuef.UesbiIuB = &b
		}
	}
	return w, nil
}

func convertWireToCheckIMEIRes(w *gsm_map.CheckIMEIRes) (*CheckIMEIRes, error) {
	v := &CheckIMEIRes{EquipmentStatus: w.EquipmentStatus}
	if w.Bmuef != nil {
		v.BMUEF = &UESBIIu{}
		if w.Bmuef.UesbiIuA != nil {
			v.BMUEF.UESBIIuA = bitStringFromWire(*w.Bmuef.UesbiIuA)
			v.BMUEF.UESBIIuABits = w.Bmuef.UesbiIuA.BitLength
		}
		if w.Bmuef.UesbiIuB != nil {
			v.BMUEF.UESBIIuB = bitStringFromWire(*w.Bmuef.UesbiIuB)
			v.BMUEF.UESBIIuBBits = w.Bmuef.UesbiIuB.BitLength
		}
	}
	return v, nil
}
