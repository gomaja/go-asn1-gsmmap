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
	"strings"

	"github.com/gomaja/go-asn1/runtime"
	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1-gsmmap/tbcd"
)

// groupIdFiller is the six-TBCD-nibble placeholder GroupId must carry
// whenever the LongGroupId field is populated, per TS 29.002.
const groupIdFiller = "ffffff"

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

func convertWireToZoneCodeList(w gsm_map.ZoneCodeList) (ZoneCodeList, error) {
	if w == nil {
		return nil, nil
	}
	if len(w) < 1 || len(w) > MaxNumOfZoneCodes {
		return nil, ErrZoneCodeListInvalidSize
	}
	out := make(ZoneCodeList, 0, len(w))
	for i, zc := range w {
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
	out := &VoiceBroadcastData{
		GroupId:                  decodeGroupID(w.Groupid),
		BroadcastInitEntitlement: w.BroadcastInitEntitlement != nil,
	}
	if w.LongGroupId != nil {
		out.LongGroupId = decodeLongGroupID(*w.LongGroupId)
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
	if w == nil {
		return nil, nil
	}
	if len(w) < 1 || len(w) > MaxNumOfVBSGroupIds {
		return nil, ErrVBSDataListInvalidSize
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
		if len(v.AdditionalInfo) > MaxAdditionalInfoOctets {
			return nil, fmt.Errorf("VoiceGroupCallData.AdditionalInfo: %w", ErrAdditionalInfoTooLong)
		}
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
	out := &VoiceGroupCallData{GroupId: decodeGroupID(w.GroupId)}
	if w.AdditionalSubscriptions != nil {
		out.AdditionalSubscriptions = convertBitStringToAdditionalSubscriptions(*w.AdditionalSubscriptions)
	}
	if w.AdditionalInfo != nil && w.AdditionalInfo.BitLength > 0 {
		// Byte-aligned-only public type per the VoiceGroupCallData.Additional-
		// Info godoc: take full octets only (BitLength / 8, floor),
		// discarding any sub-byte trailing bits. A BitLength of 7 surfaces
		// zero bytes; callers who need sub-byte handling should read the
		// underlying BIT STRING directly.
		byteLen := w.AdditionalInfo.BitLength / 8
		// Spec max is 136 bits = 17 octets; reject larger inputs. Use the
		// ceiling of BitLength to catch over-spec encodings that also
		// carry sub-byte trailing bits.
		if (w.AdditionalInfo.BitLength+7)/8 > MaxAdditionalInfoOctets {
			return nil, fmt.Errorf("VoiceGroupCallData.AdditionalInfo: %w", ErrAdditionalInfoTooLong)
		}
		if byteLen > len(w.AdditionalInfo.Bytes) {
			byteLen = len(w.AdditionalInfo.Bytes)
		}
		if byteLen > 0 {
			out.AdditionalInfo = HexBytes(w.AdditionalInfo.Bytes[:byteLen])
		}
	}
	if w.LongGroupId != nil {
		out.LongGroupId = decodeLongGroupID(*w.LongGroupId)
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
	if w == nil {
		return nil, nil
	}
	if len(w) < 1 || len(w) > MaxNumOfVGCSGroupIds {
		return nil, ErrVGCSDataListInvalidSize
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

// encodeGroupID enforces the TBCD GroupId invariants per TS 29.002:
//   - GroupId is mandatory when no LongGroupId is present;
//   - when LongGroupId IS present, GroupId must be the six TBCD fillers
//     "ffffff" (case-insensitive);
//   - the encoded value must be exactly 3 octets (6 hex nibbles).
func encodeGroupID(gid string, hasLong bool) (gsm_map.GroupId, error) {
	if hasLong {
		if !strings.EqualFold(gid, groupIdFiller) {
			return nil, ErrGroupIdFillerRequired
		}
	} else if gid == "" {
		return nil, ErrGroupIdMissingWithoutLong
	}
	enc, err := tbcd.Encode(gid)
	if err != nil {
		return nil, err
	}
	if len(enc) != GroupIdOctets {
		return nil, fmt.Errorf("%w: got %d octets from %q", ErrGroupIdInvalidEncodedLength, len(enc), gid)
	}
	return gsm_map.GroupId(enc), nil
}

// encodeLongGroupID enforces the 4-octet SIZE constraint on LongGroupId
// per TS 29.002 MAP-MS-DataTypes.asn:2735.
func encodeLongGroupID(s string) (gsm_map.LongGroupId, error) {
	enc, err := tbcd.Encode(s)
	if err != nil {
		return nil, err
	}
	if len(enc) != LongGroupIdOctets {
		return nil, fmt.Errorf("%w: got %d octets from %q", ErrLongGroupIdInvalidEncodedLength, len(enc), s)
	}
	return gsm_map.LongGroupId(enc), nil
}

// decodeGroupID returns the raw nibble-swapped hex of a TBCD GroupId
// without tbcd.Decode's trailing-'f' filler strip. Group IDs are
// TBCD-encoded hex identifiers per TS 23.003 — not phone numbers — so
// trailing 'f' nibbles can be legitimate data (or six-filler padding
// when LongGroupId is present). Either way the caller sees the exact
// nibble sequence the wire carried.
func decodeGroupID(raw []byte) string {
	return rawTBCDHex(raw)
}

// decodeLongGroupID mirrors decodeGroupID for the 4-octet LongGroupId
// field; tbcd.Decode would otherwise strip legitimate trailing 'f'
// nibbles in the identifier.
func decodeLongGroupID(raw []byte) string {
	return rawTBCDHex(raw)
}

// rawTBCDHex performs a pure nibble swap on TBCD bytes, returning a
// lowercase hex string. No filler stripping — the caller sees exactly
// what the wire carried.
func rawTBCDHex(raw []byte) string {
	out := make([]byte, len(raw)*2)
	const hexDigits = "0123456789abcdef"
	for i, b := range raw {
		out[i*2] = hexDigits[b&0x0f]
		out[i*2+1] = hexDigits[(b>>4)&0x0f]
	}
	return string(out)
}
