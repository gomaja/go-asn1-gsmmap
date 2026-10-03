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
		v := isdn
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
		s, nature, plan, err := decodeAddressWithDigits(*w.ExternalAddress, ErrLCSClientExternalIDExternalAddressDecodedEmpty)
		if err != nil {
			return nil, fmt.Errorf("decoding LCSClientExternalID.ExternalAddress: %w", err)
		}
		out.ExternalAddress = s
		out.ExternalAddressNature = nature
		out.ExternalAddressPlan = plan
	}
	return out, nil
}

// ============================================================================
// GMLC-Restriction / NotificationToMSUser — 3GPP TS 29.002 V19.1.0 §17.7.1
// ============================================================================

// isValidGMLCRestriction reports whether v is a listed GMLC-Restriction.
func isValidGMLCRestriction(v GMLCRestriction) bool {
	return v == GMLCRestrictionGmlcList || v == GMLCRestrictionHomeCountry
}

// isValidNotificationToMSUser reports whether v is a listed
// NotificationToMSUser, locationNotAllowed included.
func isValidNotificationToMSUser(v NotificationToMSUser) bool {
	switch v {
	case NotifyLocationAllowed,
		NotifyAndVerifyLocationAllowedIfNoResponse,
		NotifyAndVerifyLocationNotAllowedIfNoResponse,
		NotificationLocationNotAllowed:
		return true
	}
	return false
}

// gmlcRestrictionFromWire applies 3GPP TS 29.002 V19.1.0 §17.7.1
// GMLC-Restriction: "At reception of any other value than the ones listed
// the receiver shall ignore GMLC-Restriction." The parameter is OPTIONAL in
// every container, so an ignored one decodes as absent (nil).
func gmlcRestrictionFromWire(w *gsm_map.GMLCRestriction) *GMLCRestriction {
	if w == nil || !isValidGMLCRestriction(*w) {
		return nil
	}
	v := *w
	return &v
}

// notificationToMSUserFromWire applies 3GPP TS 29.002 V19.1.0 §17.7.1
// NotificationToMSUser: "At reception of any other value than the ones
// listed the receiver shall ignore NotificationToMSUser." The parameter is
// OPTIONAL in every container, so an ignored one decodes as absent (nil),
// and the TS 23.271 default applies as when it is not received.
func notificationToMSUserFromWire(w *gsm_map.NotificationToMSUser) *NotificationToMSUser {
	if w == nil || !isValidNotificationToMSUser(*w) {
		return nil
	}
	v := *w
	return &v
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
	if c.GmlcRestriction != nil && !isValidGMLCRestriction(*c.GmlcRestriction) {
		return nil, fmt.Errorf("ExternalClient.GmlcRestriction: %w (got %d)", ErrGMLCRestrictionInvalid, *c.GmlcRestriction)
	}
	if c.NotificationToMSUser != nil && !isValidNotificationToMSUser(*c.NotificationToMSUser) {
		return nil, fmt.Errorf("ExternalClient.NotificationToMSUser: %w (got %d)", ErrNotificationToMSUserInvalid, *c.NotificationToMSUser)
	}
	out := &gsm_map.ExternalClient{ClientIdentity: *id}
	if c.GmlcRestriction != nil {
		v := *c.GmlcRestriction
		out.GmlcRestriction = &v
	}
	if c.NotificationToMSUser != nil {
		v := *c.NotificationToMSUser
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
	return &ExternalClient{
		ClientIdentity:       *id,
		GmlcRestriction:      gmlcRestrictionFromWire(w.GmlcRestriction),
		NotificationToMSUser: notificationToMSUserFromWire(w.NotificationToMSUser),
	}, nil
}

func convertExternalClientListToWire(list ExternalClientList) (*gsm_map.ExternalClientList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.ExternalClientList{Values: make([]gsm_map.ExternalClient, len(list))}
	for i, c := range list {
		w, err := convertExternalClientToWire(&c)
		if err != nil {
			return nil, fmt.Errorf("ExternalClientList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return &out, nil
}

func convertWireToExternalClientList(w *gsm_map.ExternalClientList) (ExternalClientList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(ExternalClientList, len(w.Values))
	for i, c := range w.Values {
		v, err := convertWireToExternalClient(&c)
		if err != nil {
			return nil, fmt.Errorf("ExternalClientList[%d]: %w", i, err)
		}
		out[i] = *v
	}
	return out, nil
}

func convertExtExternalClientListToWire(list ExtExternalClientList) (*gsm_map.ExtExternalClientList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.ExtExternalClientList{Values: make([]gsm_map.ExternalClient, len(list))}
	for i, c := range list {
		w, err := convertExternalClientToWire(&c)
		if err != nil {
			return nil, fmt.Errorf("ExtExternalClientList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return &out, nil
}

func convertWireToExtExternalClientList(w *gsm_map.ExtExternalClientList) (ExtExternalClientList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(ExtExternalClientList, len(w.Values))
	for i, c := range w.Values {
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

func convertPLMNClientListToWire(list PLMNClientList) (*gsm_map.PLMNClientList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.PLMNClientList{Values: make([]gsm_map.LCSClientInternalID, len(list))}
	for i, v := range list {
		if v < LCSClientBroadcastService || v > LCSClientTargetMSsubscribedService {
			return nil, fmt.Errorf("PLMNClientList[%d]: %w (got %d)", i, ErrLCSClientInternalIDInvalid, v)
		}
		out.Values[i] = v
	}
	return &out, nil
}

func convertWireToPLMNClientList(w *gsm_map.PLMNClientList) PLMNClientList {
	if w == nil {
		return nil
	}

	out := make(PLMNClientList, len(w.Values))
	// LCSClientInternalID is extensible (3GPP TS 29.002 V19.1.0 §17.7.8), so
	// an unlisted value is kept (§17.1.4); Marshal sends only listed values.
	copy(out, w.Values)
	return out
}

// ============================================================================
// ServiceType / ServiceTypeList — TS 29.002 MAP-MS-DataTypes.asn:2045-2056
// ============================================================================

func convertServiceTypeToWire(s *ServiceType) (*gsm_map.ServiceType, error) {
	if s == nil {
		return nil, nil
	}

	if s.GmlcRestriction != nil && !isValidGMLCRestriction(*s.GmlcRestriction) {
		return nil, fmt.Errorf("ServiceType.GmlcRestriction: %w (got %d)", ErrGMLCRestrictionInvalid, *s.GmlcRestriction)
	}
	if s.NotificationToMSUser != nil && !isValidNotificationToMSUser(*s.NotificationToMSUser) {
		return nil, fmt.Errorf("ServiceType.NotificationToMSUser: %w (got %d)", ErrNotificationToMSUserInvalid, *s.NotificationToMSUser)
	}
	out := &gsm_map.ServiceType{ServiceTypeIdentity: s.ServiceTypeIdentity}
	if s.GmlcRestriction != nil {
		v := *s.GmlcRestriction
		out.GmlcRestriction = &v
	}
	if s.NotificationToMSUser != nil {
		v := *s.NotificationToMSUser
		out.NotificationToMSUser = &v
	}
	return out, nil
}

func convertWireToServiceType(w *gsm_map.ServiceType) *ServiceType {
	if w == nil {
		return nil
	}

	return &ServiceType{
		ServiceTypeIdentity:  w.ServiceTypeIdentity,
		GmlcRestriction:      gmlcRestrictionFromWire(w.GmlcRestriction),
		NotificationToMSUser: notificationToMSUserFromWire(w.NotificationToMSUser),
	}
}

func convertServiceTypeListToWire(list ServiceTypeList) (*gsm_map.ServiceTypeList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.ServiceTypeList{Values: make([]gsm_map.ServiceType, len(list))}
	for i, s := range list {
		w, err := convertServiceTypeToWire(&s)
		if err != nil {
			return nil, fmt.Errorf("ServiceTypeList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return &out, nil
}

func convertWireToServiceTypeList(w *gsm_map.ServiceTypeList) ServiceTypeList {
	if w == nil {
		return nil
	}

	out := make(ServiceTypeList, len(w.Values))
	for i, s := range w.Values {
		out[i] = *convertWireToServiceType(&s)
	}
	return out
}

// ============================================================================
// LCSPrivacyClass / LCSPrivacyExceptionList
// — TS 29.002 MAP-MS-DataTypes.asn:1971-1996
// ============================================================================

func convertLCSPrivacyClassToWire(c *LCSPrivacyClass) (*gsm_map.LCSPrivacyClass, error) {
	if c == nil {
		return nil, nil
	}

	if c.NotificationToMSUser != nil && !isValidNotificationToMSUser(*c.NotificationToMSUser) {
		return nil, fmt.Errorf("LCSPrivacyClass.NotificationToMSUser: %w (got %d)", ErrNotificationToMSUserInvalid, *c.NotificationToMSUser)
	}
	out := &gsm_map.LCSPrivacyClass{
		SsCode:   gsm_map.SSCode{byte(c.SsCode)},
		SsStatus: gsm_map.ExtSSStatus(c.SsStatus),
	}
	if c.NotificationToMSUser != nil {
		v := *c.NotificationToMSUser
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

	out := &LCSPrivacyClass{
		SsCode:               SsCode(w.SsCode[0]),
		SsStatus:             HexBytes(w.SsStatus),
		NotificationToMSUser: notificationToMSUserFromWire(w.NotificationToMSUser),
	}
	if w.ExternalClientList != nil {
		l, err := convertWireToExternalClientList(w.ExternalClientList)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyClass.ExternalClientList: %w", err)
		}
		out.ExternalClientList = l
	}
	if w.PlmnClientList != nil {
		out.PlmnClientList = convertWireToPLMNClientList(w.PlmnClientList)
	}
	if w.ExtExternalClientList != nil {
		l, err := convertWireToExtExternalClientList(w.ExtExternalClientList)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyClass.ExtExternalClientList: %w", err)
		}
		out.ExtExternalClientList = l
	}
	if w.ServiceTypeList != nil {
		out.ServiceTypeList = convertWireToServiceTypeList(w.ServiceTypeList)
	}
	return out, nil
}

func convertLCSPrivacyExceptionListToWire(list LCSPrivacyExceptionList) (*gsm_map.LCSPrivacyExceptionList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.LCSPrivacyExceptionList{Values: make([]gsm_map.LCSPrivacyClass, len(list))}
	for i, c := range list {
		w, err := convertLCSPrivacyClassToWire(&c)
		if err != nil {
			return nil, fmt.Errorf("LCSPrivacyExceptionList[%d]: %w", i, err)
		}
		out.Values[i] = *w
	}
	return &out, nil
}

func convertWireToLCSPrivacyExceptionList(w *gsm_map.LCSPrivacyExceptionList) (LCSPrivacyExceptionList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(LCSPrivacyExceptionList, len(w.Values))
	for i, c := range w.Values {
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

func convertMOLRClassToWire(c *MOLRClass) *gsm_map.MOLRClass {
	if c == nil {
		return nil
	}

	return &gsm_map.MOLRClass{
		SsCode:   gsm_map.SSCode{byte(c.SsCode)},
		SsStatus: gsm_map.ExtSSStatus(c.SsStatus),
	}
}

func convertWireToMOLRClass(w *gsm_map.MOLRClass) *MOLRClass {
	if w == nil {
		return nil
	}

	return &MOLRClass{
		SsCode:   SsCode(w.SsCode[0]),
		SsStatus: HexBytes(w.SsStatus),
	}
}

func convertMOLRListToWire(list MOLRList) *gsm_map.MOLRList {
	if list == nil {
		return nil
	}

	out := gsm_map.MOLRList{Values: make([]gsm_map.MOLRClass, len(list))}
	for i, c := range list {
		out.Values[i] = *convertMOLRClassToWire(&c)
	}
	return &out
}

func convertWireToMOLRList(w *gsm_map.MOLRList) MOLRList {
	if w == nil {
		return nil
	}

	out := make(MOLRList, len(w.Values))
	for i, c := range w.Values {
		out[i] = *convertWireToMOLRClass(&c)
	}
	return out
}

// ============================================================================
// GMLCList — TS 29.002 MAP-MS-DataTypes.asn:1503
// ============================================================================

func convertGMLCListToWire(list GMLCList) (*gsm_map.GMLCList, error) {
	if list == nil {
		return nil, nil
	}

	out := gsm_map.GMLCList{Values: make([]gsm_map.ISDNAddressString, len(list))}
	for i, a := range list {
		if a.Address == "" {
			return nil, fmt.Errorf("GMLCList[%d]: %w", i, ErrGMLCAddressEmpty)
		}
		isdn, err := encodeAddressField(a.Address, a.Nature, a.Plan)
		if err != nil {
			return nil, fmt.Errorf("GMLCList[%d]: %w", i, err)
		}
		out.Values[i] = isdn
	}
	return &out, nil
}

func convertWireToGMLCList(w *gsm_map.GMLCList) (GMLCList, error) {
	if w == nil {
		return nil, nil
	}

	out := make(GMLCList, len(w.Values))
	for i, a := range w.Values {
		s, nature, plan, err := decodeAddressField(a)
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
		out.MolrList = convertMOLRListToWire(l.MolrList)
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
		out.MolrList = convertWireToMOLRList(w.MolrList)
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
