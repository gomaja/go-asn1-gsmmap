// SubscriberData sub-struct converters for InsertSubscriberData (opCode 7).
//
// This file covers the small, self-contained SubscriberData sub-types:
// ODB-Data, ZoneCode(List), VBS/VGCS data entries + lists. Ext-SS-Info is
// in convert_extssinfo.go and CAMEL subscription info in convert_camel.go.
//
// All converters follow the established *ToWire / *ToDomain naming
// and propagate errors with a typed prefix so callers can match them
// via errors.Is against the package-level sentinels.

package gsmmap

import (
	"bytes"
	"fmt"

	"github.com/gomaja/go-asn1-gsmmap/tbcd"
	"github.com/gomaja/go-asn1/runtime"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
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

func convertZoneCodeListToWire(z ZoneCodeList) (*gsm_map.ZoneCodeList, error) {
	out := gsm_map.ZoneCodeList{Values: make([]gsm_map.ZoneCode, 0, len(z))}
	for i, zc := range z {
		if len(zc) != 2 {
			return nil, fmt.Errorf("ZoneCodeList[%d]: %w", i, ErrZoneCodeInvalidSize)
		}
		out.Values = append(out.Values, gsm_map.ZoneCode(zc))
	}
	return &out, nil
}

func convertWireToZoneCodeList(w *gsm_map.ZoneCodeList) (ZoneCodeList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(ZoneCodeList, 0, len(w.Values))
	for i, zc := range w.Values {
		if len(zc) != 2 {
			return nil, fmt.Errorf("ZoneCodeList[%d]: %w", i, ErrZoneCodeInvalidSize)
		}
		out = append(out, ZoneCode(zc))
	}
	return out, nil
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
		lg, err := encodeLongGroupID(v.LongGroupId)
		if err != nil {
			return nil, fmt.Errorf("VoiceBroadcastData.LongGroupId: %w", err)
		}
		out.LongGroupId = &lg
	}
	return out, nil
}

func convertWireToVoiceBroadcastData(w *gsm_map.VoiceBroadcastData) (*VoiceBroadcastData, error) {
	gid, err := decodeGroupID(w.Groupid, w.LongGroupId != nil)
	if err != nil {
		return nil, fmt.Errorf("VoiceBroadcastData.GroupId: %w", err)
	}
	out := &VoiceBroadcastData{
		GroupId:                  gid,
		BroadcastInitEntitlement: w.BroadcastInitEntitlement != nil,
	}
	if w.LongGroupId != nil {
		lg, err := decodeLongGroupID(*w.LongGroupId)
		if err != nil {
			return nil, fmt.Errorf("VoiceBroadcastData.LongGroupId: %w", err)
		}
		out.LongGroupId = lg
	}
	return out, nil
}

func convertVBSDataListToWire(list VBSDataList) (*gsm_map.VBSDataList, error) {
	out := gsm_map.VBSDataList{Values: make([]gsm_map.VoiceBroadcastData, 0, len(list))}
	for i := range list {
		w, err := convertVoiceBroadcastDataToWire(&list[i])
		if err != nil {
			return nil, fmt.Errorf("VBSDataList[%d]: %w", i, err)
		}
		out.Values = append(out.Values, *w)
	}
	return &out, nil
}

func convertWireToVBSDataList(w *gsm_map.VBSDataList) (VBSDataList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(VBSDataList, 0, len(w.Values))
	for i := range w.Values {
		v, err := convertWireToVoiceBroadcastData(&w.Values[i])
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
		// AdditionalInfo is modeled as HexBytes per the public type's
		// godoc — byte-aligned only. Set BitLength to len(bytes)*8;
		// non-byte-aligned peer values are lossy on decode.
		bs := runtime.BitString{Bytes: []byte(v.AdditionalInfo), BitLength: len(v.AdditionalInfo) * 8}
		out.AdditionalInfo = &bs
	}
	if v.LongGroupId != "" {
		lg, err := encodeLongGroupID(v.LongGroupId)
		if err != nil {
			return nil, fmt.Errorf("VoiceGroupCallData.LongGroupId: %w", err)
		}
		out.LongGroupId = &lg
	}
	return out, nil
}

func convertWireToVoiceGroupCallData(w *gsm_map.VoiceGroupCallData) (*VoiceGroupCallData, error) {
	gid, err := decodeGroupID(w.GroupId, w.LongGroupId != nil)
	if err != nil {
		return nil, fmt.Errorf("VoiceGroupCallData.GroupId: %w", err)
	}
	out := &VoiceGroupCallData{GroupId: gid}
	if w.AdditionalSubscriptions != nil {
		out.AdditionalSubscriptions = convertBitStringToAdditionalSubscriptions(*w.AdditionalSubscriptions)
	}
	if w.AdditionalInfo != nil {
		// Byte-aligned-only public type per the VoiceGroupCallData.Additional-
		// Info godoc: take full octets only (BitLength / 8, floor),
		// discarding any sub-byte trailing bits. A BitLength of 7 surfaces
		// zero bytes; callers who need sub-byte handling should read the
		// underlying BIT STRING directly.
		byteLen := w.AdditionalInfo.BitLength / 8
		if byteLen > 0 {
			out.AdditionalInfo = HexBytes(w.AdditionalInfo.Bytes[:byteLen])
		}
	}
	if w.LongGroupId != nil {
		lg, err := decodeLongGroupID(*w.LongGroupId)
		if err != nil {
			return nil, fmt.Errorf("VoiceGroupCallData.LongGroupId: %w", err)
		}
		out.LongGroupId = lg
	}
	return out, nil
}

func convertVGCSDataListToWire(list VGCSDataList) (*gsm_map.VGCSDataList, error) {
	out := gsm_map.VGCSDataList{Values: make([]gsm_map.VoiceGroupCallData, 0, len(list))}
	for i := range list {
		w, err := convertVoiceGroupCallDataToWire(&list[i])
		if err != nil {
			return nil, fmt.Errorf("VGCSDataList[%d]: %w", i, err)
		}
		out.Values = append(out.Values, *w)
	}
	return &out, nil
}

func convertWireToVGCSDataList(w *gsm_map.VGCSDataList) (VGCSDataList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(VGCSDataList, 0, len(w.Values))
	for i := range w.Values {
		v, err := convertWireToVoiceGroupCallData(&w.Values[i])
		if err != nil {
			return nil, fmt.Errorf("VGCSDataList[%d]: %w", i, err)
		}
		out = append(out, *v)
	}
	return out, nil
}

// encodeGroupID enforces the TBCD GroupId invariants per TS 29.002 V19.1.0
// (MAP-MS-DataTypes):
//
//	GroupId ::= TBCD-STRING (SIZE (3))
//	-- When Group-Id is less than six characters in length, the TBCD filler (1111)
//	-- is used to fill unused half octets.
//
// and, on VoiceGroupCallData / VoiceBroadcastData,
// "groupId shall be filled with six TBCD fillers (1111) if the longGroupId
// is present". gid is the digit string (TBCD alphabet, 1..6 characters):
//   - GroupId is mandatory when no LongGroupId is present;
//   - when LongGroupId IS present, gid must be empty and the wire carries
//     the six fillers (three 0xFF octets);
//   - the digits are padded with fillers to the 3-octet field size.
func encodeGroupID(gid string, hasLong bool) (gsm_map.GroupId, error) {
	if hasLong {
		if gid != "" {
			return nil, ErrGroupIdFillerRequired
		}
	} else if gid == "" {
		return nil, ErrGroupIdMissingWithoutLong
	}
	enc, err := encodeFixedTBCD(gid, groupIdOctets)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

// encodeLongGroupID encodes a Long-Group-Id per TS 29.002 V19.1.0:
//
//	Long-GroupId ::= TBCD-STRING (SIZE (4))
//	-- When Long-Group-Id is less than eight characters in length, the TBCD filler (1111)
//	-- is used to fill unused half octets.
func encodeLongGroupID(s string) (gsm_map.LongGroupId, error) {
	enc, err := encodeFixedTBCD(s, longGroupIdOctets)
	if err != nil {
		return nil, err
	}
	return enc, nil
}

// The fixed sizes of GroupId ::= TBCD-STRING (SIZE (3)) and
// Long-GroupId ::= TBCD-STRING (SIZE (4)), 3GPP TS 29.002 V19.1.0 §17.7.1.
const (
	groupIdOctets     = 3
	longGroupIdOctets = 4
)

// encodeFixedTBCD encodes s as a TBCD-STRING (SIZE (octets)): the digits,
// then TBCD filler (1111) in every unused half octet up to the field size.
// Digits that need more octets are returned as they are, for the BER
// codec's SIZE check to reject.
func encodeFixedTBCD(s string, octets int) ([]byte, error) {
	enc, err := tbcd.Encode(s)
	if err != nil || len(enc) >= octets {
		return enc, err
	}
	out := bytes.Repeat([]byte{0xFF}, octets)
	copy(out, enc)
	return out, nil
}

// decodeGroupID returns the digits of a TBCD GroupId, applying the rules of
// encodeGroupID. Filler padding is dropped, so the six fillers decode to "".
// With a LongGroupId present the GroupId "shall be filled with six TBCD
// fillers" (3GPP TS 29.002 V19.1.0 §17.7.1 VoiceBroadcastData,
// VoiceGroupCallData), so any digit is ErrGroupIdFillerRequired; without one
// the GroupId must carry digits (ErrGroupIdMissingWithoutLong).
func decodeGroupID(raw []byte, hasLong bool) (string, error) {
	s, err := tbcd.Decode(raw)
	if err != nil {
		return "", err
	}
	switch {
	case hasLong && s != "":
		return "", ErrGroupIdFillerRequired
	case !hasLong && s == "":
		return "", ErrGroupIdMissingWithoutLong
	}
	return s, nil
}

// decodeLongGroupID mirrors decodeGroupID for the 4-octet LongGroupId field.
func decodeLongGroupID(raw []byte) (string, error) {
	s, err := tbcd.Decode(raw)
	if err != nil {
		return "", err
	}
	if s == "" {
		return "", ErrLongGroupIdDecodedEmpty
	}
	return s, nil
}
