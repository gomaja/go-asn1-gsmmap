package gsmmap

import "fmt"

type senderOperation string

const (
	senderInsertSubscriberData senderOperation = "InsertSubscriberData"
	senderSendRoutingInfo      senderOperation = "SendRoutingInfo"
)

// These mapped operations are outside call-independent supplementary services
// (3GPP TS 29.002 V19.1.0 §17.7.9 and §17.7.10).
func (operation senderOperation) excludesCompoundServices() bool {
	return operation == senderInsertSubscriberData || operation == senderSendRoutingInfo
}

// 3GPP TS 29.002 V19.1.0 §17.7.5 SS-Code reserves these assignments.
func validateSSCodeToWire(code SsCode, operation senderOperation) error {
	switch code {
	case 0x00, 0x10, 0x15, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
		0x30, 0x32, 0x40, 0x50, 0x60, 0x70, 0x80, 0xa0:
		return fmt.Errorf("%s SS-Code 0x%02x: %w", operation, code, ErrSSCodeReserved)
	}
	if operation == senderSendRoutingInfo && (code == 0x43 || code == 0x44) {
		return fmt.Errorf("%s SS-Code 0x%02x: %w", operation, code, ErrSSCodeOperationInvalid)
	}
	return nil
}

// 3GPP TS 29.002 V19.1.0 §17.7.1 ForwardingOptions and Ext-ForwOptions.
func validateForwardingOptionsToWire(raw []byte) error {
	if len(raw) > 0 && raw[0]&0x13 != 0 {
		return ErrExtForwOptionsUnusedBits
	}
	return nil
}

// 3GPP TS 29.002 V19.1.0 §17.7.1 Ext-ForwOptions reserves octets 2..5.
func validateExtForwardingOptionsToWire(raw []byte) error {
	if err := validateForwardingOptionsToWire(raw); err != nil {
		return err
	}
	if len(raw) > 1 && len(raw) <= 5 {
		return ErrExtForwOptionsReservedOctets
	}
	return nil
}

// 3GPP TS 29.002 V19.1.0 §17.7.8 ISDN-SubaddressString.
func validateISDNSubaddressToWire(raw []byte) error {
	if len(raw) > 0 {
		header := raw[0]
		typeBits := header & 0x70
		if header&0x80 == 0 || (typeBits != 0 && typeBits != 0x20) || header&0x07 != 0 {
			return ErrISDNSubaddressHeaderInvalid
		}
	}
	return nil
}

// 3GPP TS 29.002 V19.1.0 §17.7.8 Ext-SS-Status.
func validateExtSSStatusToWire(raw []byte) error {
	if len(raw) > 0 && raw[0]&0xf0 != 0 {
		return ErrExtSSStatusUnusedBits
	}
	if len(raw) > 1 && len(raw) <= 5 {
		return ErrExtSSStatusReservedOctets
	}
	return nil
}

// 3GPP TS 29.002 V19.1.0 §17.7.8 LCSServiceTypeID navigation.
func validateLCSServiceTypeIDToWire(value int64) error {
	if value == 8 {
		return ErrLCSServiceTypeIDReserved
	}
	return nil
}

// 3GPP TS 29.002 V19.1.0 §17.7.1 LCS-PrivacyClass; SS-Code assignments §17.7.5.
func validateLCSPrivacyClassToWire(c *LCSPrivacyClass) error {
	callSession := c.SsCode == 0xb2 || c.SsCode == 0xb3
	switch {
	case c.NotificationToMSUser != nil && !callSession:
		return ErrLCSPrivacyNotificationNotApplicable
	case c.ExternalClientList != nil && !callSession:
		return ErrLCSPrivacyExternalClientListNotApplicable
	case c.PlmnClientList != nil && c.SsCode != 0xb4:
		return ErrLCSPrivacyPLMNClientListNotApplicable
	case c.ExtExternalClientList != nil && len(c.ExternalClientList)+len(c.ExtExternalClientList) <= 5:
		return ErrLCSPrivacyExtendedClientCountInvalid
	case c.ServiceTypeList != nil && c.SsCode != 0xb5:
		return ErrLCSPrivacyServiceTypeListNotApplicable
	}
	return nil
}

// 3GPP TS 29.002 V19.1.0 §17.7.9 Ext-TeleserviceCode and compound groups.
func validateExtTeleserviceCodeToWire(raw []byte, operation senderOperation) error {
	if len(raw) > 1 && len(raw) <= 5 {
		return ErrExtTeleserviceCodeReservedOctets
	}
	if len(raw) > 0 && operation.excludesCompoundServices() && (raw[0] == 0x70 || raw[0] == 0x80) {
		return fmt.Errorf("%s: %w", operation, ErrCompoundTeleserviceCodeNotAllowed)
	}
	return nil
}

// 3GPP TS 29.002 V19.1.0 §17.7.10 Ext-BearerServiceCode and compound groups.
func validateExtBearerServiceCodeToWire(raw []byte, operation senderOperation) error {
	if len(raw) > 1 && len(raw) <= 5 {
		return ErrExtBearerServiceCodeReservedOctets
	}
	if len(raw) > 0 {
		code := raw[0]
		if code&0x80 != 0 && code&0xf0 != 0xd0 {
			return ErrBearerServiceCodeUnusedBit
		}
		if operation.excludesCompoundServices() {
			switch code {
			case 0x50, 0x58, 0x60, 0x68:
				return fmt.Errorf("%s: %w", operation, ErrCompoundBearerServiceCodeNotAllowed)
			}
		}
	}
	return nil
}
