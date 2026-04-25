package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// GPRSCamelTDPData / GPRSCamelTDPDataList
// — TS 29.002 MAP-MS-DataTypes.asn:1620-1635
// ============================================================================

func convertGPRSCamelTDPDataToWire(d *GPRSCamelTDPData) (*gsm_map.GPRSCamelTDPData, error) {
	if d == nil {
		return nil, nil
	}
	if d.GsmSCFAddress == "" {
		return nil, fmt.Errorf("GPRSCamelTDPData.GsmSCFAddress: mandatory field must not be empty on encode")
	}
	if d.ServiceKey < 0 || d.ServiceKey > 2147483647 {
		return nil, fmt.Errorf("GPRSCamelTDPData.ServiceKey: %w (got %d)", ErrCamelInvalidServiceKey, d.ServiceKey)
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
		GprsTriggerDetectionPoint: gsm_map.GPRSTriggerDetectionPoint(d.GprsTriggerDetectionPoint),
		ServiceKey:                gsm_map.ServiceKey(d.ServiceKey),
		GsmSCFAddress:             gsm_map.ISDNAddressString(addr),
		DefaultSessionHandling:    gsm_map.DefaultGPRSHandling(d.DefaultSessionHandling),
	}, nil
}

func convertWireToGPRSCamelTDPData(w *gsm_map.GPRSCamelTDPData) (*GPRSCamelTDPData, error) {
	if w == nil {
		return nil, nil
	}
	if len(w.GsmSCFAddress) == 0 {
		return nil, fmt.Errorf("GPRSCamelTDPData.GsmSCFAddress: mandatory field must be present on the wire")
	}
	addr, nature, plan, err := decodeAddressField([]byte(w.GsmSCFAddress))
	if err != nil {
		return nil, fmt.Errorf("decoding GPRSCamelTDPData.GsmSCFAddress: %w", err)
	}
	if addr == "" {
		return nil, fmt.Errorf("decoding GPRSCamelTDPData.GsmSCFAddress: empty digits in mandatory ISDN-AddressString")
	}
	if int64(w.ServiceKey) < 0 || int64(w.ServiceKey) > 2147483647 {
		return nil, fmt.Errorf("GPRSCamelTDPData.ServiceKey: %w (got %d)", ErrCamelInvalidServiceKey, w.ServiceKey)
	}
	sk := int64(w.ServiceKey)
	// DefaultGPRSHandling: spec exception clause (TS 29.002
	// MAP-MS-DataTypes.asn:1638-1640) says decoders MUST treat
	//   - values 2..31  as continueTransaction (0)
	//   - values >  31  as releaseTransaction (1)
	dgh := DefaultGPRSHandling(w.DefaultSessionHandling)
	switch {
	case dgh < 0:
		return nil, fmt.Errorf("%w (got %d)", ErrDefaultGPRSHandlingInvalid, dgh)
	case dgh >= 2 && dgh <= 31:
		dgh = DefaultGPRSContinueTransaction
	case dgh > 31:
		dgh = DefaultGPRSReleaseTransaction
	}
	return &GPRSCamelTDPData{
		GprsTriggerDetectionPoint: GPRSTriggerDetectionPoint(w.GprsTriggerDetectionPoint),
		ServiceKey:                sk,
		GsmSCFAddress:             addr,
		GsmSCFAddressNature:       nature,
		GsmSCFAddressPlan:         plan,
		DefaultSessionHandling:    dgh,
	}, nil
}

func convertGPRSCamelTDPDataListToWire(list GPRSCamelTDPDataList) (gsm_map.GPRSCamelTDPDataList, error) {
	if list == nil {
		return nil, nil
	}
	if int64(len(list)) < 1 || int64(len(list)) > gsm_map.MaxNumOfCamelTDPData {
		return nil, fmt.Errorf("%w (got %d)", ErrGPRSCamelTDPDataListSize, len(list))
	}
	out := make(gsm_map.GPRSCamelTDPDataList, len(list))
	for i, d := range list {
		w, err := convertGPRSCamelTDPDataToWire(&d)
		if err != nil {
			return nil, fmt.Errorf("GPRSCamelTDPDataList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToGPRSCamelTDPDataList(w gsm_map.GPRSCamelTDPDataList) (GPRSCamelTDPDataList, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w)) < 1 || int64(len(w)) > gsm_map.MaxNumOfCamelTDPData {
		return nil, fmt.Errorf("%w (got %d)", ErrGPRSCamelTDPDataListSize, len(w))
	}
	out := make(GPRSCamelTDPDataList, len(w))
	for i, d := range w {
		v, err := convertWireToGPRSCamelTDPData(&d)
		if err != nil {
			return nil, fmt.Errorf("GPRSCamelTDPDataList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// ============================================================================
// GPRSCSI — TS 29.002 MAP-MS-DataTypes.asn:1606
// ============================================================================

func convertGPRSCSIToWire(g *GPRSCSI) (*gsm_map.GPRSCSI, error) {
	if g == nil {
		return nil, nil
	}
	// Per TS 29.002 MAP-MS-DataTypes.asn:1615-1616: when GPRS-CSI is
	// present, BOTH GprsCamelTDPDataList AND CamelCapabilityHandling
	// SHALL be present.
	if g.GprsCamelTDPDataList == nil || g.CamelCapabilityHandling == nil {
		return nil, ErrGPRSCSIRequiresTDPListAndPhase
	}
	if g.CamelCapabilityHandling != nil {
		if v := *g.CamelCapabilityHandling; v < 1 || v > 4 {
			return nil, fmt.Errorf("%w (got %d)", ErrCamelCapabilityHandlingOutOfRange, v)
		}
	}
	out := &gsm_map.GPRSCSI{
		NotificationToCSE: boolToNullPtr(g.NotificationToCSE),
		CsiActive:         boolToNullPtr(g.CsiActive),
	}
	if g.GprsCamelTDPDataList != nil {
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

func convertWireToGPRSCSI(w *gsm_map.GPRSCSI) (*GPRSCSI, error) {
	if w == nil {
		return nil, nil
	}
	if w.GprsCamelTDPDataList == nil || w.CamelCapabilityHandling == nil {
		return nil, ErrGPRSCSIRequiresTDPListAndPhase
	}
	out := &GPRSCSI{
		NotificationToCSE: nullPtrToBool(w.NotificationToCSE),
		CsiActive:         nullPtrToBool(w.CsiActive),
	}
	if w.GprsCamelTDPDataList != nil {
		dl, err := convertWireToGPRSCamelTDPDataList(w.GprsCamelTDPDataList)
		if err != nil {
			return nil, err
		}
		out.GprsCamelTDPDataList = dl
	}
	if w.CamelCapabilityHandling != nil {
		v, err := narrowInt64Range(int64(*w.CamelCapabilityHandling), 1, 4, "GPRSCSI.CamelCapabilityHandling")
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrCamelCapabilityHandlingOutOfRange, err)
		}
		out.CamelCapabilityHandling = &v
	}
	return out, nil
}

// ============================================================================
// MGCSI — TS 29.002 MAP-MS-DataTypes.asn:2528
// ============================================================================

func convertMGCSIToWire(m *MGCSI) (*gsm_map.MGCSI, error) {
	if m == nil {
		return nil, nil
	}
	if int64(len(m.MobilityTriggers)) < 1 || int64(len(m.MobilityTriggers)) > gsm_map.MaxNumOfMobilityTriggers {
		return nil, fmt.Errorf("%w (got %d)", ErrMobilityTriggersSize, len(m.MobilityTriggers))
	}
	mt := make(gsm_map.MobilityTriggers, len(m.MobilityTriggers))
	for i, c := range m.MobilityTriggers {
		if len(c) != 1 {
			return nil, fmt.Errorf("MobilityTriggers[%d]: %w (got %d)", i, ErrMMCodeInvalidSize, len(c))
		}
		mt[i] = gsm_map.MMCode(c)
	}
	if m.GsmSCFAddress == "" {
		return nil, fmt.Errorf("MGCSI.GsmSCFAddress: mandatory field must not be empty on encode")
	}
	if m.ServiceKey < 0 || m.ServiceKey > 2147483647 {
		return nil, fmt.Errorf("MGCSI.ServiceKey: %w (got %d)", ErrCamelInvalidServiceKey, m.ServiceKey)
	}
	addr, err := encodeAddressField(m.GsmSCFAddress, m.GsmSCFAddressNature, m.GsmSCFAddressPlan)
	if err != nil {
		return nil, fmt.Errorf("encoding MGCSI.GsmSCFAddress: %w", err)
	}
	return &gsm_map.MGCSI{
		MobilityTriggers:  mt,
		ServiceKey:        gsm_map.ServiceKey(m.ServiceKey),
		GsmSCFAddress:     gsm_map.ISDNAddressString(addr),
		NotificationToCSE: boolToNullPtr(m.NotificationToCSE),
		CsiActive:         boolToNullPtr(m.CsiActive),
	}, nil
}

func convertWireToMGCSI(w *gsm_map.MGCSI) (*MGCSI, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w.MobilityTriggers)) < 1 || int64(len(w.MobilityTriggers)) > gsm_map.MaxNumOfMobilityTriggers {
		return nil, fmt.Errorf("%w (got %d)", ErrMobilityTriggersSize, len(w.MobilityTriggers))
	}
	mt := make([]HexBytes, len(w.MobilityTriggers))
	for i, c := range w.MobilityTriggers {
		if len(c) != 1 {
			return nil, fmt.Errorf("MobilityTriggers[%d]: %w (got %d)", i, ErrMMCodeInvalidSize, len(c))
		}
		mt[i] = HexBytes(c)
	}
	if len(w.GsmSCFAddress) == 0 {
		return nil, fmt.Errorf("MGCSI.GsmSCFAddress: mandatory field must be present on the wire")
	}
	addr, nature, plan, err := decodeAddressField([]byte(w.GsmSCFAddress))
	if err != nil {
		return nil, fmt.Errorf("decoding MGCSI.GsmSCFAddress: %w", err)
	}
	if addr == "" {
		return nil, fmt.Errorf("decoding MGCSI.GsmSCFAddress: empty digits in mandatory ISDN-AddressString")
	}
	if int64(w.ServiceKey) < 0 || int64(w.ServiceKey) > 2147483647 {
		return nil, fmt.Errorf("MGCSI.ServiceKey: %w (got %d)", ErrCamelInvalidServiceKey, w.ServiceKey)
	}
	sk := int64(w.ServiceKey)
	return &MGCSI{
		MobilityTriggers:    mt,
		ServiceKey:          sk,
		GsmSCFAddress:       addr,
		GsmSCFAddressNature: nature,
		GsmSCFAddressPlan:   plan,
		NotificationToCSE:   nullPtrToBool(w.NotificationToCSE),
		CsiActive:           nullPtrToBool(w.CsiActive),
	}, nil
}

// ============================================================================
// SGSNCAMELSubscriptionInfo — TS 29.002 MAP-MS-DataTypes.asn:1596
// SMSCSI / MTSmsCAMELTDPCriteria converters reused from PR C
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
		v, err := convertSMSCSIToWire(s.MoSmsCSI)
		if err != nil {
			return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MoSmsCSI: %w", err)
		}
		out.MoSmsCSI = v
	}
	if s.MtSmsCSI != nil {
		v, err := convertSMSCSIToWire(s.MtSmsCSI)
		if err != nil {
			return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MtSmsCSI: %w", err)
		}
		out.MtSmsCSI = v
	}
	if s.MtSmsCAMELTDPCriteriaList != nil {
		// Reuse PR C per-element converter. List bound is enforced by
		// the SMS-CSI domain (1..10 entries per spec).
		if int64(len(s.MtSmsCAMELTDPCriteriaList)) < 1 || int64(len(s.MtSmsCAMELTDPCriteriaList)) > gsm_map.MaxNumOfCamelTDPData {
			return nil, fmt.Errorf("%w (got %d)", ErrSGSNMtSmsCAMELTDPCriteriaListSize, len(s.MtSmsCAMELTDPCriteriaList))
		}
		list := make(gsm_map.MTSmsCAMELTDPCriteriaList, len(s.MtSmsCAMELTDPCriteriaList))
		for i, c := range s.MtSmsCAMELTDPCriteriaList {
			w, err := convertMTSmsCAMELTDPCriteriaToWire(&c)
			if err != nil {
				return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MtSmsCAMELTDPCriteriaList[%d]: %w", i, err)
			}
			list[i] = w
		}
		out.MtSmsCAMELTDPCriteriaList = list
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
		v, err := convertWireToSMSCSI(w.MoSmsCSI)
		if err != nil {
			return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MoSmsCSI: %w", err)
		}
		out.MoSmsCSI = v
	}
	if w.MtSmsCSI != nil {
		v, err := convertWireToSMSCSI(w.MtSmsCSI)
		if err != nil {
			return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MtSmsCSI: %w", err)
		}
		out.MtSmsCSI = v
	}
	if w.MtSmsCAMELTDPCriteriaList != nil {
		if int64(len(w.MtSmsCAMELTDPCriteriaList)) < 1 || int64(len(w.MtSmsCAMELTDPCriteriaList)) > gsm_map.MaxNumOfCamelTDPData {
			return nil, fmt.Errorf("%w (got %d)", ErrSGSNMtSmsCAMELTDPCriteriaListSize, len(w.MtSmsCAMELTDPCriteriaList))
		}
		list := make([]MTSmsCAMELTDPCriteria, len(w.MtSmsCAMELTDPCriteriaList))
		for i, c := range w.MtSmsCAMELTDPCriteriaList {
			v, err := convertWireToMTSmsCAMELTDPCriteria(&c)
			if err != nil {
				return nil, fmt.Errorf("SGSNCAMELSubscriptionInfo.MtSmsCAMELTDPCriteriaList[%d]: %w", i, err)
			}
			list[i] = *v
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
