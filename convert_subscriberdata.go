// SubscriberData sub-struct converters for InsertSubscriberData (opCode 7).
//
// This file covers the small, self-contained SubscriberData sub-types:
// ODB-Data, ZoneCode(List), VBS/VGCS data entries + lists. Deeper
// CHOICEs (Ext-SS-Info) and CAMEL subscription info are addressed in
// follow-up PRs.
//
// All converters follow the established *ToWire / *ToDomain naming
// and propagate errors with a typed prefix so callers can match them
// via errors.Is against the package-level sentinels.

package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/runtime"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// --- ODB-Data (MAP-MS-DataTypes.asn:1770) ---

func convertODBDataToWire(o *ODBData) (*gsm_map.ODBData, error) {
	if o.OdbGeneralData == nil {
		return nil, ErrODBDataMissingGeneralData
	}
	out := &gsm_map.ODBData{
		OdbGeneralData: convertODBGeneralDataToBitString(o.OdbGeneralData),
	}
	if o.OdbHPLMNData != nil {
		bs := convertODBHPLMNDataToBitString(o.OdbHPLMNData)
		out.OdbHPLMNData = &bs
	}
	return out, nil
}

func convertWireToODBData(w *gsm_map.ODBData) *ODBData {
	out := &ODBData{
		OdbGeneralData: convertBitStringToODBGeneralData(w.OdbGeneralData),
	}
	if w.OdbHPLMNData != nil {
		out.OdbHPLMNData = convertBitStringToODBHPLMNData(*w.OdbHPLMNData)
	}
	return out
}

// --- ZoneCode / ZoneCodeList (MAP-MS-DataTypes.asn:2070) ---

func convertZoneCodeListToWire(z ZoneCodeList) (gsm_map.ZoneCodeList, error) {
	if len(z) < 1 || len(z) > MaxNumOfZoneCodes {
		return nil, ErrZoneCodeListInvalidSize
	}
	out := make(gsm_map.ZoneCodeList, 0, len(z))
	for i, zc := range z {
		if len(zc) != 2 {
			return nil, fmt.Errorf("ZoneCodeList[%d]: %w", i, ErrZoneCodeInvalidSize)
		}
		out = append(out, gsm_map.ZoneCode(zc))
	}
	return out, nil
}

func convertWireToZoneCodeList(w gsm_map.ZoneCodeList) ZoneCodeList {
	if len(w) == 0 {
		return nil
	}
	out := make(ZoneCodeList, 0, len(w))
	for _, zc := range w {
		out = append(out, ZoneCode(zc))
	}
	return out
}

// --- VoiceBroadcastData / VBSDataList (MAP-MS-DataTypes.asn:2685, 2717) ---

func convertVoiceBroadcastDataToWire(v *VoiceBroadcastData) (*gsm_map.VoiceBroadcastData, error) {
	gid, err := encodeGroupID(v.GroupId, v.LongGroupId != "")
	if err != nil {
		return nil, fmt.Errorf("VoiceBroadcastData.GroupId: %w", err)
	}
	out := &gsm_map.VoiceBroadcastData{Groupid: gid}
	if v.BroadcastInitEntitlement {
		out.BroadcastInitEntitlement = &struct{}{}
	}
	if v.LongGroupId != "" {
		lgid, err := tbcd.Encode(v.LongGroupId)
		if err != nil {
			return nil, fmt.Errorf("VoiceBroadcastData.LongGroupId: %w", err)
		}
		lg := gsm_map.LongGroupId(lgid)
		out.LongGroupId = &lg
	}
	return out, nil
}

func convertWireToVoiceBroadcastData(w *gsm_map.VoiceBroadcastData) (*VoiceBroadcastData, error) {
	out := &VoiceBroadcastData{
		GroupId:                  decodeGroupID(w.Groupid, w.LongGroupId != nil),
		BroadcastInitEntitlement: w.BroadcastInitEntitlement != nil,
	}
	if w.LongGroupId != nil {
		lgid, err := tbcd.Decode(*w.LongGroupId)
		if err != nil {
			return nil, fmt.Errorf("VoiceBroadcastData.LongGroupId: %w", err)
		}
		out.LongGroupId = lgid
	}
	return out, nil
}

func convertVBSDataListToWire(list VBSDataList) (gsm_map.VBSDataList, error) {
	if len(list) < 1 || len(list) > MaxNumOfVBSGroupIds {
		return nil, ErrVBSDataListInvalidSize
	}
	out := make(gsm_map.VBSDataList, 0, len(list))
	for i := range list {
		w, err := convertVoiceBroadcastDataToWire(&list[i])
		if err != nil {
			return nil, fmt.Errorf("VBSDataList[%d]: %w", i, err)
		}
		out = append(out, *w)
	}
	return out, nil
}

func convertWireToVBSDataList(w gsm_map.VBSDataList) (VBSDataList, error) {
	if len(w) == 0 {
		return nil, nil
	}
	out := make(VBSDataList, 0, len(w))
	for i := range w {
		v, err := convertWireToVoiceBroadcastData(&w[i])
		if err != nil {
			return nil, fmt.Errorf("VBSDataList[%d]: %w", i, err)
		}
		out = append(out, *v)
	}
	return out, nil
}

// --- VoiceGroupCallData / VGCSDataList (MAP-MS-DataTypes.asn:2688, 2695) ---

func convertVoiceGroupCallDataToWire(v *VoiceGroupCallData) (*gsm_map.VoiceGroupCallData, error) {
	gid, err := encodeGroupID(v.GroupId, v.LongGroupId != "")
	if err != nil {
		return nil, fmt.Errorf("VoiceGroupCallData.GroupId: %w", err)
	}
	out := &gsm_map.VoiceGroupCallData{GroupId: gid}
	if v.AdditionalSubscriptions != nil {
		bs := convertAdditionalSubscriptionsToBitString(v.AdditionalSubscriptions)
		out.AdditionalSubscriptions = &bs
	}
	if len(v.AdditionalInfo) > 0 {
		// AdditionalInfo is an opaque BIT STRING per TS 43.068 —
		// surface the raw octets without reinterpreting them.
		bs := runtime.BitString{Bytes: []byte(v.AdditionalInfo), BitLength: len(v.AdditionalInfo) * 8}
		out.AdditionalInfo = &bs
	}
	if v.LongGroupId != "" {
		lgid, err := tbcd.Encode(v.LongGroupId)
		if err != nil {
			return nil, fmt.Errorf("VoiceGroupCallData.LongGroupId: %w", err)
		}
		lg := gsm_map.LongGroupId(lgid)
		out.LongGroupId = &lg
	}
	return out, nil
}

func convertWireToVoiceGroupCallData(w *gsm_map.VoiceGroupCallData) (*VoiceGroupCallData, error) {
	out := &VoiceGroupCallData{GroupId: decodeGroupID(w.GroupId, w.LongGroupId != nil)}
	if w.AdditionalSubscriptions != nil {
		out.AdditionalSubscriptions = convertBitStringToAdditionalSubscriptions(*w.AdditionalSubscriptions)
	}
	if w.AdditionalInfo != nil && w.AdditionalInfo.BitLength > 0 {
		out.AdditionalInfo = HexBytes(w.AdditionalInfo.Bytes)
	}
	if w.LongGroupId != nil {
		lgid, err := tbcd.Decode(*w.LongGroupId)
		if err != nil {
			return nil, fmt.Errorf("VoiceGroupCallData.LongGroupId: %w", err)
		}
		out.LongGroupId = lgid
	}
	return out, nil
}

func convertVGCSDataListToWire(list VGCSDataList) (gsm_map.VGCSDataList, error) {
	if len(list) < 1 || len(list) > MaxNumOfVGCSGroupIds {
		return nil, ErrVGCSDataListInvalidSize
	}
	out := make(gsm_map.VGCSDataList, 0, len(list))
	for i := range list {
		w, err := convertVoiceGroupCallDataToWire(&list[i])
		if err != nil {
			return nil, fmt.Errorf("VGCSDataList[%d]: %w", i, err)
		}
		out = append(out, *w)
	}
	return out, nil
}

func convertWireToVGCSDataList(w gsm_map.VGCSDataList) (VGCSDataList, error) {
	if len(w) == 0 {
		return nil, nil
	}
	out := make(VGCSDataList, 0, len(w))
	for i := range w {
		v, err := convertWireToVoiceGroupCallData(&w[i])
		if err != nil {
			return nil, fmt.Errorf("VGCSDataList[%d]: %w", i, err)
		}
		out = append(out, *v)
	}
	return out, nil
}

// encodeGroupID handles the TBCD filler rule per TS 29.002: when a
// LongGroupId is being emitted alongside the primary GroupId octet,
// the GroupId must be filled with six TBCD filler digits ("ffffff").
// We accept either the literal filler string or reject missing
// GroupId entirely so callers see the constraint.
func encodeGroupID(gid string, hasLong bool) (gsm_map.GroupId, error) {
	if gid == "" {
		if hasLong {
			return nil, ErrGroupIdFillerRequired
		}
		return nil, ErrGroupIdMissingWithoutLong
	}
	enc, err := tbcd.Encode(gid)
	if err != nil {
		return nil, err
	}
	return gsm_map.GroupId(enc), nil
}

// decodeGroupID decodes a 3-octet TBCD GroupId without tbcd.Decode's
// trailing-'f' filler strip. Group IDs are TBCD-encoded hex identifiers
// per TS 23.003, not phone numbers, so trailing 'f' nibbles can be
// legitimate data (or six-filler padding when LongGroupId is present).
// Either way, the caller sees the exact nibble sequence the wire carried.
// The hasLong parameter is kept for future use but both paths return
// the raw nibble-swapped hex today.
func decodeGroupID(raw []byte, _ bool) string {
	out := make([]byte, len(raw)*2)
	const hexDigits = "0123456789abcdef"
	for i, b := range raw {
		out[i*2] = hexDigits[b&0x0f]
		out[i*2+1] = hexDigits[(b>>4)&0x0f]
	}
	return string(out)
}
