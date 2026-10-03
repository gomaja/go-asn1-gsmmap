package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// GPRSCamelTDPData / GPRSCamelTDPDataList
// — 3GPP TS 29.002 V19.1.0 §17.7.1
// ============================================================================

// isValidGPRSTriggerDetectionPoint reports whether v is a listed
// GPRS-TriggerDetectionPoint (3GPP TS 29.002 V19.1.0 §17.7.1).
func isValidGPRSTriggerDetectionPoint(v GPRSTriggerDetectionPoint) bool {
	switch v {
	case GPRSTDPAttach,
		GPRSTDPAttachChangeOfPosition,
		GPRSTDPPdpContextEstablishment,
		GPRSTDPPdpContextEstablishmentAcknowledgement,
		GPRSTDPPdpContextChangeOfPosition:
		return true
	}
	return false
}

func convertGPRSCamelTDPDataToWire(d *GPRSCamelTDPData) (*gsm_map.GPRSCamelTDPData, error) {
	if d == nil {
		return nil, nil
	}
	if !isValidGPRSTriggerDetectionPoint(d.GprsTriggerDetectionPoint) {
		return nil, fmt.Errorf("%w (got %d)", ErrGPRSTriggerDetectionPointInvalid, d.GprsTriggerDetectionPoint)
	}
	if d.GsmSCFAddress == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}

	addr, err := encodeAddressField(d.GsmSCFAddress, d.GsmSCFAddressNature, d.GsmSCFAddressPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding GPRSCamelTDPData.GsmSCFAddress: %w", err)
	}
	if d.DefaultSessionHandling < 0 || d.DefaultSessionHandling > 1 {
		// Encoder is strict: caller must use a defined enum value.
		// Decoder applies the spec lenient remap (>1 → releaseTransaction).
		return nil, fmt.Errorf("%w (got %d)", ErrDefaultGPRSHandlingInvalid, d.DefaultSessionHandling)
	}
	return &gsm_map.GPRSCamelTDPData{
		GprsTriggerDetectionPoint: d.GprsTriggerDetectionPoint,
		ServiceKey:                d.ServiceKey,
		GsmSCFAddress:             addr,
		DefaultSessionHandling:    d.DefaultSessionHandling,
	}, nil
}

// convertWireToGPRSCamelTDPData decodes one GPRS-CamelTDPData. It returns
// nil, without examining the other fields, for an entry the receiver
// ignores: 3GPP TS 29.002 V19.1.0 §17.7.1 GPRS-TriggerDetectionPoint, "For
// GPRS-CamelTDPData sequences containing this parameter with any other value
// than the ones listed the receiver shall ignore the whole GPRS-CamelTDPData
// sequence."
func convertWireToGPRSCamelTDPData(w *gsm_map.GPRSCamelTDPData) (*GPRSCamelTDPData, error) {
	if !isValidGPRSTriggerDetectionPoint(w.GprsTriggerDetectionPoint) {
		return nil, nil
	}

	addr, nature, plan, err := decodeAddressField(w.GsmSCFAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding GPRSCamelTDPData.GsmSCFAddress: %w", err)
	}
	if addr == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}

	sk := w.ServiceKey
	// DefaultGPRSHandling: spec exception clause (TS 29.002
	// 3GPP TS 29.002 V19.1.0 §17.7.1) says decoders MUST treat
	//   - values 2..31  as continueTransaction (0)
	//   - values >  31  as releaseTransaction (1)
	// A negative value lies outside both ranges; the type is extensible,
	// so it is kept (3GPP TS 29.002 V19.1.0 §17.1.4) and Marshal rejects it.
	dgh := w.DefaultSessionHandling
	switch {
	case dgh >= 2 && dgh <= 31:
		dgh = DefaultGPRSContinueTransaction
	case dgh > 31:
		dgh = DefaultGPRSReleaseTransaction
	}
	return &GPRSCamelTDPData{
		GprsTriggerDetectionPoint: w.GprsTriggerDetectionPoint,
		ServiceKey:                sk,
		GsmSCFAddress:             addr,
		GsmSCFAddressNature:       nature,
		GsmSCFAddressPlan:         plan,
		DefaultSessionHandling:    dgh,
	}, nil
}

func convertGPRSCamelTDPDataListToWire(list GPRSCamelTDPDataList) (*gsm_map.GPRSCamelTDPDataList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.GPRSCamelTDPDataList{Values: make([]gsm_map.GPRSCamelTDPData, len(list))}
	seen := map[GPRSTriggerDetectionPoint]bool{}
	for i, d := range list {
		w, err := convertGPRSCamelTDPDataToWire(&d)
		if err != nil {
			return nil, fmt.Errorf("GPRSCamelTDPDataList[%d]: %w", i, err)
		}
		if err := checkTDPOnce(seen, w.GprsTriggerDetectionPoint); err != nil {
			return nil, fmt.Errorf("GPRSCamelTDPDataList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return &out, nil
}

// convertWireToGPRSCamelTDPDataList decodes a GPRS-CamelTDPDataList. It
// returns nil when the receiver ignores every entry.
func convertWireToGPRSCamelTDPDataList(w *gsm_map.GPRSCamelTDPDataList) (GPRSCamelTDPDataList, error) {
	if w == nil {
		return nil, nil
	}
	return convertTDPDataWireList("GPRSCamelTDPDataList", w.Values, convertWireToGPRSCamelTDPData,
		func(d *GPRSCamelTDPData) GPRSTriggerDetectionPoint { return d.GprsTriggerDetectionPoint })
}

// ============================================================================
// GPRSCSI — 3GPP TS 29.002 V19.1.0 §17.7.1
// ============================================================================

// convertGPRSCSIToWire encodes a GPRS-CSI. The list and
// camelCapabilityHandling are encoded when set; their presence depends on
// the segment (convertDCSIToWire in convert_camel.go).
func convertGPRSCSIToWire(g *GPRSCSI) (*gsm_map.GPRSCSI, error) {
	if g == nil {
		return nil, nil
	}
	if err := validateCamelCapabilityHandling(g.CamelCapabilityHandling); err != nil {
		return nil, err
	}
	out := &gsm_map.GPRSCSI{}
	if len(g.GprsCamelTDPDataList) > 0 {
		dl, err := convertGPRSCamelTDPDataListToWire(g.GprsCamelTDPDataList)
		if err != nil {
			return nil, err
		}
		out.GprsCamelTDPDataList = dl
	}
	if g.CamelCapabilityHandling != nil {
		v := gsm_map.CamelCapabilityHandling(*g.CamelCapabilityHandling)
		out.CamelCapabilityHandling = &v
	}
	return out, nil
}

// convertWireToGPRSCSI decodes a wire GPRS-CSI. An absent list, or one
// whose every GPRS-CamelTDPData the receiver ignores
// (convertWireToGPRSCamelTDPData), decodes to a nil list; the rest of the
// CSI is kept, since another segment may carry the TDP data.
func convertWireToGPRSCSI(w *gsm_map.GPRSCSI) (*GPRSCSI, error) {
	if w == nil {
		return nil, nil
	}
	dl, err := convertWireToGPRSCamelTDPDataList(w.GprsCamelTDPDataList)
	if err != nil {
		return nil, err
	}
	out := &GPRSCSI{
		GprsCamelTDPDataList: dl,
	}
	if w.CamelCapabilityHandling != nil {
		out.CamelCapabilityHandling = camelCapabilityHandlingFromWire(*w.CamelCapabilityHandling)
	}
	return out, nil
}

// ============================================================================
// MGCSI — 3GPP TS 29.002 V19.1.0 §17.7.1
// ============================================================================

func convertMGCSIToWire(m *MGCSI) (*gsm_map.MGCSI, error) {
	if m == nil {
		return nil, nil
	}

	mt := gsm_map.MobilityTriggers{Values: make([]gsm_map.MMCode, len(m.MobilityTriggers))}
	for i, c := range m.MobilityTriggers {
		if !isPSMMCode(c) {
			return nil, fmt.Errorf("MGCSI.MobilityTriggers[%d]=0x%02x: %w", i, byte(c), ErrMGCSIMMCodeInvalid)
		}
		mt.Values[i] = gsm_map.MMCode{byte(c)}
	}
	if m.GsmSCFAddress == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}

	addr, err := encodeAddressField(m.GsmSCFAddress, m.GsmSCFAddressNature, m.GsmSCFAddressPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding MGCSI.GsmSCFAddress: %w", err)
	}
	return &gsm_map.MGCSI{
		MobilityTriggers: &mt,
		ServiceKey:       m.ServiceKey,
		GsmSCFAddress:    addr,
	}, nil
}

// convertWireToMGCSI decodes an MG-CSI received by the SGSN. It returns nil
// when the receiver ignores every MM-Code (mmCodesFromWire): an MG-CSI arms
// its events only through MobilityTriggers, SIZE (1..10), so with none left
// the receiver holds no MG-CSI.
func convertWireToMGCSI(w *gsm_map.MGCSI) (*MGCSI, error) {
	if w == nil {
		return nil, nil
	}
	mt, err := mmCodesFromWire("MGCSI.MobilityTriggers", w.MobilityTriggers, isPSMMCode)
	if err != nil {
		return nil, err
	}
	if mt == nil {
		return nil, nil
	}

	addr, nature, plan, err := decodeAddressField(w.GsmSCFAddress)
	if err != nil {
		return nil, fmt.Errorf("decoding MGCSI.GsmSCFAddress: %w", err)
	}
	if addr == "" {
		return nil, ErrCamelMissingGsmSCFAddress
	}

	sk := w.ServiceKey
	return &MGCSI{
		MobilityTriggers:    mt,
		ServiceKey:          sk,
		GsmSCFAddress:       addr,
		GsmSCFAddressNature: nature,
		GsmSCFAddressPlan:   plan,
	}, nil
}

// ============================================================================
// SGSNCAMELSubscriptionInfo — 3GPP TS 29.002 V19.1.0 §17.7.1
// SMSCSI / MTSmsCAMELTDPCriteria converters shared with VLR CAMEL data.
// (convert_camel.go: convertSMSCSIToWire/convertWireToSMSCSI,
// convertMTSmsCAMELTDPCriteriaToWire/convertWireToMTSmsCAMELTDPCriteria).
// ============================================================================

func convertSGSNCAMELSubscriptionInfoToWire(s *SGSNCAMELSubscriptionInfo) (*gsm_map.SGSNCAMELSubscriptionInfo, error) {
	if s == nil {
		return nil, nil
	}
	out := &gsm_map.SGSNCAMELSubscriptionInfo{}
	if s.GprsCSI != nil {
		v, err := convertGPRSCSIToWire(s.GprsCSI)
		if err != nil {
			return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.GprsCSI: %w", err)
		}
		out.GprsCSI = v
	}
	if s.MoSmsCSI != nil {
		v, err := convertSMSCSIToWire(s.MoSmsCSI, moSMSTriggerDetectionPoint)
		if err != nil {
			return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MoSmsCSI: %w", err)
		}
		out.MoSmsCSI = v
	}
	if s.MtSmsCSI != nil {
		v, err := convertSMSCSIToWire(s.MtSmsCSI, mtSMSTriggerDetectionPoint)
		if err != nil {
			return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MtSmsCSI: %w", err)
		}
		out.MtSmsCSI = v
	}
	if s.MtSmsCAMELTDPCriteriaList != nil {
		// Reuse the SMS-CSI entry converter.
		list := gsm_map.MTSmsCAMELTDPCriteriaList{Values: make([]gsm_map.MTSmsCAMELTDPCriteria, len(s.MtSmsCAMELTDPCriteriaList))}
		for i, c := range s.MtSmsCAMELTDPCriteriaList {
			w, err := convertMTSmsCAMELTDPCriteriaToWire(&c)
			if err != nil {
				return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MtSmsCAMELTDPCriteriaList[%d]: %w", i, err)
			}
			list.Values[i] = w
		}
		out.MtSmsCAMELTDPCriteriaList = &list
	}
	if s.MgCsi != nil {
		v, err := convertMGCSIToWire(s.MgCsi)
		if err != nil {
			return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MgCsi: %w", err)
		}
		out.MgCsi = v
	}
	return out, nil
}

func convertWireToSGSNCAMELSubscriptionInfo(w *gsm_map.SGSNCAMELSubscriptionInfo) (*SGSNCAMELSubscriptionInfo, error) {
	if w == nil {
		return nil, nil
	}
	out := &SGSNCAMELSubscriptionInfo{}
	if w.GprsCSI != nil {
		v, err := convertWireToGPRSCSI(w.GprsCSI)
		if err != nil {
			return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.GprsCSI: %w", err)
		}
		out.GprsCSI = v
	}
	if w.MoSmsCSI != nil {
		v, err := convertWireToSMSCSI(w.MoSmsCSI, moSMSTriggerDetectionPoint)
		if err != nil {
			return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MoSmsCSI: %w", err)
		}
		out.MoSmsCSI = v
	}
	if w.MtSmsCSI != nil {
		v, err := convertWireToSMSCSI(w.MtSmsCSI, mtSMSTriggerDetectionPoint)
		if err != nil {
			return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MtSmsCSI: %w", err)
		}
		out.MtSmsCSI = v
	}
	if w.MtSmsCAMELTDPCriteriaList != nil {
		// Absent when the receiver ignores every entry.
		list, err := convertIgnorableWireList("SGSNCAMELSubscriptionInfo.MtSmsCAMELTDPCriteriaList", w.MtSmsCAMELTDPCriteriaList.Values, convertWireToMTSmsCAMELTDPCriteria)
		if err != nil {
			return nil, err
		}
		out.MtSmsCAMELTDPCriteriaList = list
	}
	if w.MgCsi != nil {
		v, err := convertWireToMGCSI(w.MgCsi)
		if err != nil {
			return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MgCsi: %w", err)
		}
		out.MgCsi = v
	}
	return out, nil
}
