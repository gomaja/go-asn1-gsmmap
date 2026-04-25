package gsmmap

import (
	"fmt"

	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// ============================================================================
// LCSClientExternalID — TS 29.002 MAP-CommonDataTypes.asn (gsm_map.LCSClientExternalID)
// ============================================================================

func convertLCSClientExternalIDToWire(c *LCSClientExternalID) (*gsm_map.LCSClientExternalID, error) {
	if c == nil {
		return nil, nil
	}
	out := &gsm_map.LCSClientExternalID{}
	if c.ExternalAddress != "" {
		isdn, err := encodeAddressField(c.ExternalAddress, c.ExternalAddressNature, c.ExternalAddressPlan)
		if err != nil {
			return nil, fmt.Errorf("encoding LCSClientExternalID.ExternalAddress: %w", err)
		}
		v := gsm_map.ISDNAddressString(isdn)
		out.ExternalAddress = &v
	}
	return out, nil
}

func convertWireToLCSClientExternalID(w *gsm_map.LCSClientExternalID) (*LCSClientExternalID, error) {
	if w == nil {
		return nil, nil
	}
	out := &LCSClientExternalID{}
	if w.ExternalAddress != nil {
		s, nature, plan, err := decodeAddressField([]byte(*w.ExternalAddress))
		if err != nil {
			return nil, fmt.Errorf("decoding LCSClientExternalID.ExternalAddress: %w", err)
		}
		if s == "" {
			return nil, fmt.Errorf("decoding LCSClientExternalID.ExternalAddress: present wire field decoded to empty digits")
		}
		out.ExternalAddress = s
		out.ExternalAddressNature = nature
		out.ExternalAddressPlan = plan
	}
	return out, nil
}

// ============================================================================
// ExternalClient / ExternalClientList / ExtExternalClientList
// — TS 29.002 MAP-MS-DataTypes.asn:2003-2018
// ============================================================================

func convertExternalClientToWire(c *ExternalClient) (*gsm_map.ExternalClient, error) {
	if c == nil {
		return nil, nil
	}
	id, err := convertLCSClientExternalIDToWire(&c.ClientIdentity)
	if err != nil {
		return nil, fmt.Errorf("ExternalClient.ClientIdentity: %w", err)
	}
	if c.GmlcRestriction != nil {
		if v := *c.GmlcRestriction; v < 0 || v > 1 {
			return nil, fmt.Errorf("ExternalClient.GmlcRestriction: %w (got %d)", ErrGMLCRestrictionInvalid, v)
		}
	}
	if c.NotificationToMSUser != nil {
		if v := *c.NotificationToMSUser; v < 0 || v > 3 {
			return nil, fmt.Errorf("ExternalClient.NotificationToMSUser: %w (got %d)", ErrNotificationToMSUserInvalid, v)
		}
	}
	out := &gsm_map.ExternalClient{ClientIdentity: *id}
	if c.GmlcRestriction != nil {
		v := gsm_map.GMLCRestriction(*c.GmlcRestriction)
		out.GmlcRestriction = &v
	}
	if c.NotificationToMSUser != nil {
		v := gsm_map.NotificationToMSUser(*c.NotificationToMSUser)
		out.NotificationToMSUser = &v
	}
	return out, nil
}

func convertWireToExternalClient(w *gsm_map.ExternalClient) (*ExternalClient, error) {
	if w == nil {
		return nil, nil
	}
	id, err := convertWireToLCSClientExternalID(&w.ClientIdentity)
	if err != nil {
		return nil, fmt.Errorf("ExternalClient.ClientIdentity: %w", err)
	}
	out := &ExternalClient{ClientIdentity: *id}
	if w.GmlcRestriction != nil {
		v := GMLCRestriction(*w.GmlcRestriction)
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("ExternalClient.GmlcRestriction: %w (got %d)", ErrGMLCRestrictionInvalid, v)
		}
		out.GmlcRestriction = &v
	}
	if w.NotificationToMSUser != nil {
		v := NotificationToMSUser(*w.NotificationToMSUser)
		if v < 0 || v > 3 {
			return nil, fmt.Errorf("ExternalClient.NotificationToMSUser: %w (got %d)", ErrNotificationToMSUserInvalid, v)
		}
		out.NotificationToMSUser = &v
	}
	return out, nil
}

func convertExternalClientListToWire(list ExternalClientList) (gsm_map.ExternalClientList, error) {
	if list == nil {
		return nil, nil
	}
	if int64(len(list)) > gsm_map.MaxNumOfExternalClient {
		return nil, fmt.Errorf("%w (got %d)", ErrExternalClientListSize, len(list))
	}
	out := make(gsm_map.ExternalClientList, len(list))
	for i, c := range list {
		w, err := convertExternalClientToWire(&c)
		if err != nil {
			return nil, fmt.Errorf("ExternalClientList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToExternalClientList(w gsm_map.ExternalClientList) (ExternalClientList, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w)) > gsm_map.MaxNumOfExternalClient {
		return nil, fmt.Errorf("%w (got %d)", ErrExternalClientListSize, len(w))
	}
	out := make(ExternalClientList, len(w))
	for i, c := range w {
		v, err := convertWireToExternalClient(&c)
		if err != nil {
			return nil, fmt.Errorf("ExternalClientList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

func convertExtExternalClientListToWire(list ExtExternalClientList) (gsm_map.ExtExternalClientList, error) {
	if list == nil {
		return nil, nil
	}
	if int64(len(list)) < 1 || int64(len(list)) > gsm_map.MaxNumOfExtExternalClient {
		return nil, fmt.Errorf("%w (got %d)", ErrExtExternalClientListSize, len(list))
	}
	out := make(gsm_map.ExtExternalClientList, len(list))
	for i, c := range list {
		w, err := convertExternalClientToWire(&c)
		if err != nil {
			return nil, fmt.Errorf("ExtExternalClientList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToExtExternalClientList(w gsm_map.ExtExternalClientList) (ExtExternalClientList, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w)) < 1 || int64(len(w)) > gsm_map.MaxNumOfExtExternalClient {
		return nil, fmt.Errorf("%w (got %d)", ErrExtExternalClientListSize, len(w))
	}
	out := make(ExtExternalClientList, len(w))
	for i, c := range w {
		v, err := convertWireToExternalClient(&c)
		if err != nil {
			return nil, fmt.Errorf("ExtExternalClientList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// ============================================================================
// PLMNClientList — TS 29.002 MAP-MS-DataTypes.asn:2008
// ============================================================================

func convertPLMNClientListToWire(list PLMNClientList) (gsm_map.PLMNClientList, error) {
	if list == nil {
		return nil, nil
	}
	if int64(len(list)) < 1 || int64(len(list)) > gsm_map.MaxNumOfPLMNClient {
		return nil, fmt.Errorf("%w (got %d)", ErrPLMNClientListSize, len(list))
	}
	out := make(gsm_map.PLMNClientList, len(list))
	for i, v := range list {
		if v < LCSClientBroadcastService || v > LCSClientTargetMSsubscribedService {
			return nil, fmt.Errorf("PLMNClientList[%d]: %w (got %d)", i, ErrLCSClientInternalIDInvalid, v)
		}
		out[i] = gsm_map.LCSClientInternalID(v)
	}
	return out, nil
}

func convertWireToPLMNClientList(w gsm_map.PLMNClientList) (PLMNClientList, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w)) < 1 || int64(len(w)) > gsm_map.MaxNumOfPLMNClient {
		return nil, fmt.Errorf("%w (got %d)", ErrPLMNClientListSize, len(w))
	}
	out := make(PLMNClientList, len(w))
	for i, v := range w {
		lv := LCSClientInternalID(v)
		if lv < LCSClientBroadcastService || lv > LCSClientTargetMSsubscribedService {
			return nil, fmt.Errorf("PLMNClientList[%d]: %w (got %d)", i, ErrLCSClientInternalIDInvalid, lv)
		}
		out[i] = lv
	}
	return out, nil
}

// ============================================================================
// ServiceType / ServiceTypeList — TS 29.002 MAP-MS-DataTypes.asn:2045-2056
// ============================================================================

func convertServiceTypeToWire(s *ServiceType) (*gsm_map.ServiceType, error) {
	if s == nil {
		return nil, nil
	}
	if s.GmlcRestriction != nil {
		if v := *s.GmlcRestriction; v < 0 || v > 1 {
			return nil, fmt.Errorf("ServiceType.GmlcRestriction: %w (got %d)", ErrGMLCRestrictionInvalid, v)
		}
	}
	if s.NotificationToMSUser != nil {
		if v := *s.NotificationToMSUser; v < 0 || v > 3 {
			return nil, fmt.Errorf("ServiceType.NotificationToMSUser: %w (got %d)", ErrNotificationToMSUserInvalid, v)
		}
	}
	out := &gsm_map.ServiceType{ServiceTypeIdentity: gsm_map.LCSServiceTypeID(s.ServiceTypeIdentity)}
	if s.GmlcRestriction != nil {
		v := gsm_map.GMLCRestriction(*s.GmlcRestriction)
		out.GmlcRestriction = &v
	}
	if s.NotificationToMSUser != nil {
		v := gsm_map.NotificationToMSUser(*s.NotificationToMSUser)
		out.NotificationToMSUser = &v
	}
	return out, nil
}

func convertWireToServiceType(w *gsm_map.ServiceType) (*ServiceType, error) {
	if w == nil {
		return nil, nil
	}
	out := &ServiceType{ServiceTypeIdentity: int64(w.ServiceTypeIdentity)}
	if w.GmlcRestriction != nil {
		v := GMLCRestriction(*w.GmlcRestriction)
		if v < 0 || v > 1 {
			return nil, fmt.Errorf("ServiceType.GmlcRestriction: %w (got %d)", ErrGMLCRestrictionInvalid, v)
		}
		out.GmlcRestriction = &v
	}
	if w.NotificationToMSUser != nil {
		v := NotificationToMSUser(*w.NotificationToMSUser)
		if v < 0 || v > 3 {
			return nil, fmt.Errorf("ServiceType.NotificationToMSUser: %w (got %d)", ErrNotificationToMSUserInvalid, v)
		}
		out.NotificationToMSUser = &v
	}
	return out, nil
}

func convertServiceTypeListToWire(list ServiceTypeList) (gsm_map.ServiceTypeList, error) {
	if list == nil {
		return nil, nil
	}
	if int64(len(list)) < 1 || int64(len(list)) > gsm_map.MaxNumOfServiceType {
		return nil, fmt.Errorf("%w (got %d)", ErrServiceTypeListSize, len(list))
	}
	out := make(gsm_map.ServiceTypeList, len(list))
	for i, s := range list {
		w, err := convertServiceTypeToWire(&s)
		if err != nil {
			return nil, fmt.Errorf("ServiceTypeList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToServiceTypeList(w gsm_map.ServiceTypeList) (ServiceTypeList, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w)) < 1 || int64(len(w)) > gsm_map.MaxNumOfServiceType {
		return nil, fmt.Errorf("%w (got %d)", ErrServiceTypeListSize, len(w))
	}
	out := make(ServiceTypeList, len(w))
	for i, s := range w {
		v, err := convertWireToServiceType(&s)
		if err != nil {
			return nil, fmt.Errorf("ServiceTypeList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// ============================================================================
// LCSPrivacyClass / LCSPrivacyExceptionList
// — TS 29.002 MAP-MS-DataTypes.asn:1971-1996
// ============================================================================

func convertLCSPrivacyClassToWire(c *LCSPrivacyClass) (*gsm_map.LCSPrivacyClass, error) {
	if c == nil {
		return nil, nil
	}
	if err := validateExtSSStatus(c.SsStatus, "LCSPrivacyClass.SsStatus"); err != nil {
		return nil, err
	}
	if c.NotificationToMSUser != nil {
		if v := *c.NotificationToMSUser; v < 0 || v > 3 {
			return nil, fmt.Errorf("LCSPrivacyClass.NotificationToMSUser: %w (got %d)", ErrNotificationToMSUserInvalid, v)
		}
	}
	out := &gsm_map.LCSPrivacyClass{
		SsCode:   gsm_map.SSCode{byte(c.SsCode)},
		SsStatus: gsm_map.ExtSSStatus(c.SsStatus),
	}
	if c.NotificationToMSUser != nil {
		v := gsm_map.NotificationToMSUser(*c.NotificationToMSUser)
		out.NotificationToMSUser = &v
	}
	if c.ExternalClientList != nil {
		l, err := convertExternalClientListToWire(c.ExternalClientList)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyClass.ExternalClientList: %w", err)
		}
		out.ExternalClientList = l
	}
	if c.PlmnClientList != nil {
		l, err := convertPLMNClientListToWire(c.PlmnClientList)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyClass.PlmnClientList: %w", err)
		}
		out.PlmnClientList = l
	}
	if c.ExtExternalClientList != nil {
		l, err := convertExtExternalClientListToWire(c.ExtExternalClientList)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyClass.ExtExternalClientList: %w", err)
		}
		out.ExtExternalClientList = l
	}
	if c.ServiceTypeList != nil {
		l, err := convertServiceTypeListToWire(c.ServiceTypeList)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyClass.ServiceTypeList: %w", err)
		}
		out.ServiceTypeList = l
	}
	return out, nil
}

func convertWireToLCSPrivacyClass(w *gsm_map.LCSPrivacyClass) (*LCSPrivacyClass, error) {
	if w == nil {
		return nil, nil
	}
	if err := validateExtSSStatus(HexBytes(w.SsStatus), "LCSPrivacyClass.SsStatus"); err != nil {
		return nil, err
	}
	if len(w.SsCode) != 1 {
		return nil, fmt.Errorf("%w (got %d)", ErrLCSPrivacyClassSsCodeInvalidSize, len(w.SsCode))
	}
	out := &LCSPrivacyClass{
		SsCode:   SsCode(w.SsCode[0]),
		SsStatus: HexBytes(w.SsStatus),
	}
	if w.NotificationToMSUser != nil {
		v := NotificationToMSUser(*w.NotificationToMSUser)
		if v < 0 || v > 3 {
			return nil, fmt.Errorf("LCSPrivacyClass.NotificationToMSUser: %w (got %d)", ErrNotificationToMSUserInvalid, v)
		}
		out.NotificationToMSUser = &v
	}
	if w.ExternalClientList != nil {
		l, err := convertWireToExternalClientList(w.ExternalClientList)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyClass.ExternalClientList: %w", err)
		}
		out.ExternalClientList = l
	}
	if w.PlmnClientList != nil {
		l, err := convertWireToPLMNClientList(w.PlmnClientList)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyClass.PlmnClientList: %w", err)
		}
		out.PlmnClientList = l
	}
	if w.ExtExternalClientList != nil {
		l, err := convertWireToExtExternalClientList(w.ExtExternalClientList)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyClass.ExtExternalClientList: %w", err)
		}
		out.ExtExternalClientList = l
	}
	if w.ServiceTypeList != nil {
		l, err := convertWireToServiceTypeList(w.ServiceTypeList)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyClass.ServiceTypeList: %w", err)
		}
		out.ServiceTypeList = l
	}
	return out, nil
}

func convertLCSPrivacyExceptionListToWire(list LCSPrivacyExceptionList) (gsm_map.LCSPrivacyExceptionList, error) {
	if list == nil {
		return nil, nil
	}
	if int64(len(list)) < 1 || int64(len(list)) > gsm_map.MaxNumOfPrivacyClass {
		return nil, fmt.Errorf("%w (got %d)", ErrLCSPrivacyExceptionListSize, len(list))
	}
	out := make(gsm_map.LCSPrivacyExceptionList, len(list))
	for i, c := range list {
		w, err := convertLCSPrivacyClassToWire(&c)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyExceptionList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToLCSPrivacyExceptionList(w gsm_map.LCSPrivacyExceptionList) (LCSPrivacyExceptionList, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w)) < 1 || int64(len(w)) > gsm_map.MaxNumOfPrivacyClass {
		return nil, fmt.Errorf("%w (got %d)", ErrLCSPrivacyExceptionListSize, len(w))
	}
	out := make(LCSPrivacyExceptionList, len(w))
	for i, c := range w {
		v, err := convertWireToLCSPrivacyClass(&c)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyExceptionList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// ============================================================================
// MOLRClass / MOLRList — TS 29.002 MAP-MS-DataTypes.asn:2059-2068
// ============================================================================

func convertMOLRClassToWire(c *MOLRClass) (*gsm_map.MOLRClass, error) {
	if c == nil {
		return nil, nil
	}
	if err := validateExtSSStatus(c.SsStatus, "MOLRClass.SsStatus"); err != nil {
		return nil, err
	}
	return &gsm_map.MOLRClass{
		SsCode:   gsm_map.SSCode{byte(c.SsCode)},
		SsStatus: gsm_map.ExtSSStatus(c.SsStatus),
	}, nil
}

func convertWireToMOLRClass(w *gsm_map.MOLRClass) (*MOLRClass, error) {
	if w == nil {
		return nil, nil
	}
	if err := validateExtSSStatus(HexBytes(w.SsStatus), "MOLRClass.SsStatus"); err != nil {
		return nil, err
	}
	if len(w.SsCode) != 1 {
		return nil, fmt.Errorf("%w (got %d)", ErrMOLRClassSsCodeInvalidSize, len(w.SsCode))
	}
	return &MOLRClass{
		SsCode:   SsCode(w.SsCode[0]),
		SsStatus: HexBytes(w.SsStatus),
	}, nil
}

func convertMOLRListToWire(list MOLRList) (gsm_map.MOLRList, error) {
	if list == nil {
		return nil, nil
	}
	if int64(len(list)) < 1 || int64(len(list)) > gsm_map.MaxNumOfMOLRClass {
		return nil, fmt.Errorf("%w (got %d)", ErrMOLRListSize, len(list))
	}
	out := make(gsm_map.MOLRList, len(list))
	for i, c := range list {
		w, err := convertMOLRClassToWire(&c)
		if err != nil {
			return nil, fmt.Errorf("MOLRList[%d]: %w", i, err)
		}
		out[i] = *w
	}
	return out, nil
}

func convertWireToMOLRList(w gsm_map.MOLRList) (MOLRList, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w)) < 1 || int64(len(w)) > gsm_map.MaxNumOfMOLRClass {
		return nil, fmt.Errorf("%w (got %d)", ErrMOLRListSize, len(w))
	}
	out := make(MOLRList, len(w))
	for i, c := range w {
		v, err := convertWireToMOLRClass(&c)
		if err != nil {
			return nil, fmt.Errorf("MOLRList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

// ============================================================================
// GMLCList — TS 29.002 MAP-MS-DataTypes.asn:1503
// ============================================================================

func convertGMLCListToWire(list GMLCList) (gsm_map.GMLCList, error) {
	if list == nil {
		return nil, nil
	}
	if int64(len(list)) < 1 || int64(len(list)) > gsm_map.MaxNumOfGMLC {
		return nil, fmt.Errorf("%w (got %d)", ErrGMLCListSize, len(list))
	}
	out := make(gsm_map.GMLCList, len(list))
	for i, a := range list {
		if a.Address == "" {
			return nil, fmt.Errorf("GMLCList[%d]: %w", i, ErrGMLCAddressEmpty)
		}
		isdn, err := encodeAddressField(a.Address, a.Nature, a.Plan)
		if err != nil {
			return nil, fmt.Errorf("GMLCList[%d]: %w", i, err)
		}
		out[i] = gsm_map.ISDNAddressString(isdn)
	}
	return out, nil
}

func convertWireToGMLCList(w gsm_map.GMLCList) (GMLCList, error) {
	if w == nil {
		return nil, nil
	}
	if int64(len(w)) < 1 || int64(len(w)) > gsm_map.MaxNumOfGMLC {
		return nil, fmt.Errorf("%w (got %d)", ErrGMLCListSize, len(w))
	}
	out := make(GMLCList, len(w))
	for i, a := range w {
		s, nature, plan, err := decodeAddressField([]byte(a))
		if err != nil {
			return nil, fmt.Errorf("GMLCList[%d]: %w", i, err)
		}
		if s == "" {
			return nil, fmt.Errorf("GMLCList[%d]: %w", i, ErrGMLCAddressEmpty)
		}
		out[i] = GMLCAddress{Address: s, Nature: nature, Plan: plan}
	}
	return out, nil
}

// ============================================================================
// LCSInformation — TS 29.002 MAP-MS-DataTypes.asn:1490
// ============================================================================

func convertLCSInformationToWire(l *LCSInformation) (*gsm_map.LCSInformation, error) {
	if l == nil {
		return nil, nil
	}
	out := &gsm_map.LCSInformation{}
	if l.GmlcList != nil {
		gl, err := convertGMLCListToWire(l.GmlcList)
		if err != nil {
			return nil, err
		}
		out.GmlcList = gl
	}
	if l.LcsPrivacyExceptionList != nil {
		pe, err := convertLCSPrivacyExceptionListToWire(l.LcsPrivacyExceptionList)
		if err != nil {
			return nil, err
		}
		out.LcsPrivacyExceptionList = pe
	}
	if l.MolrList != nil {
		ml, err := convertMOLRListToWire(l.MolrList)
		if err != nil {
			return nil, err
		}
		out.MolrList = ml
	}
	if l.AddLcsPrivacyExceptionList != nil {
		al, err := convertLCSPrivacyExceptionListToWire(l.AddLcsPrivacyExceptionList)
		if err != nil {
			return nil, fmt.Errorf("LCSInformation.AddLcsPrivacyExceptionList: %w", err)
		}
		out.AddLcsPrivacyExceptionList = al
	}
	return out, nil
}

func convertWireToLCSInformation(w *gsm_map.LCSInformation) (*LCSInformation, error) {
	if w == nil {
		return nil, nil
	}
	out := &LCSInformation{}
	if w.GmlcList != nil {
		gl, err := convertWireToGMLCList(w.GmlcList)
		if err != nil {
			return nil, err
		}
		out.GmlcList = gl
	}
	if w.LcsPrivacyExceptionList != nil {
		pe, err := convertWireToLCSPrivacyExceptionList(w.LcsPrivacyExceptionList)
		if err != nil {
			return nil, err
		}
		out.LcsPrivacyExceptionList = pe
	}
	if w.MolrList != nil {
		ml, err := convertWireToMOLRList(w.MolrList)
		if err != nil {
			return nil, err
		}
		out.MolrList = ml
	}
	if w.AddLcsPrivacyExceptionList != nil {
		al, err := convertWireToLCSPrivacyExceptionList(w.AddLcsPrivacyExceptionList)
		if err != nil {
			return nil, fmt.Errorf("LCSInformation.AddLcsPrivacyExceptionList: %w", err)
		}
		out.AddLcsPrivacyExceptionList = al
	}
	return out, nil
}
