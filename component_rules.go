package gsmmap

import (
	"fmt"
	"reflect"
	"sync"

	mapvalues "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	phase2 "github.com/gomaja/go-asn1/telecom/ss7/gsm_map/gsm0902_v4_19_1"
)

type componentRuleKind uint8

const (
	componentRuleOperation componentRuleKind = iota
	componentRuleError
	componentRuleParameterAbsent
	componentRuleParameterPresent
	componentRuleFieldAbsent
	componentRuleFieldPresent
	componentRuleInteger
)

type componentVersionRule struct {
	id                     string
	edition                Edition
	module                 string
	line, endLine          int
	clause, text           string
	minVersion, maxVersion int
	kind                   componentRuleKind
	role                   componentRole
	opCode, errorCode      int64
	target                 reflect.Type
	field                  string
	number                 int64
}

// The Phase 2 rows are the 35 NU>1, OA>1, OP>1 and NU1 NU2 annotations
// from GSM 09.02 V4.19.1's VERSION-COMMENTS.json. Their line numbers refer
// to the official ASN.1.txt member of 0902-4j1.zip. The source annotations
// constrain the negotiated subset (GSM 09.02 V4.19.1 section 14.1.4).
var (
	componentVersionRulesOnce  sync.Once
	componentVersionRulesValue []componentVersionRule
)

// componentVersionRules builds the rule table on first use, like the catalogues.
func componentVersionRules() []componentVersionRule {
	componentVersionRulesOnce.Do(func() { componentVersionRulesValue = componentVersionRuleTable() })
	return componentVersionRulesValue
}

func componentVersionRuleTable() []componentVersionRule {
	return []componentVersionRule{
		{id: "p2-2445", edition: EditionPhase2, module: "MAP-CH-DataTypes", line: 2445, endLine: 2445, text: "msc-Number must be present in version greater 1", kind: componentRuleFieldPresent, target: reflect.TypeFor[phase2.ProvideRoamingNumberArg](), field: "MscNumber"},
		{id: "p2-3288", edition: EditionPhase2, module: "MAP-CommonDataTypes", line: 3288, endLine: 3289, text: "Value 3 must not be used in version 1 and must not be used in version 2.", kind: componentRuleInteger, target: reflect.TypeFor[phase2.ProtocolId](), number: 3},
		{id: "p2-3311", edition: EditionPhase2, module: "MAP-CommonDataTypes", line: 3311, endLine: 3311, text: "roamingNumber must not be used in version greater 1", kind: componentRuleInteger, target: reflect.TypeFor[phase2.LocationInfo](), field: "Choice", number: phase2.LocationInfoChoiceRoamingNumber},
		{id: "p2-3096", edition: EditionPhase2, module: "MAP-ER-DataTypes", line: 3096, endLine: 3097, text: "sm-EnumeratedDeliveryFailureCause must not be used in version greater 1", kind: componentRuleInteger, target: reflect.TypeFor[phase2.SMDeliveryFailureCause](), field: "Choice", number: phase2.SMDeliveryFailureCauseChoiceSmEnumeratedDeliveryFailureCause},
		{id: "p2-1897", edition: EditionPhase2, module: "MAP-Errors", line: 1897, endLine: 1897, text: "roamingNotAllowedCause must be present in version greater 1", kind: componentRuleParameterPresent, role: componentError, errorCode: 8},
		{id: "p2-1929", edition: EditionPhase2, module: "MAP-Errors", line: 1929, endLine: 1929, text: "mwd-Set must be absent in version greater 1", kind: componentRuleParameterAbsent, role: componentError, errorCode: 27},
		{id: "p2-1958", edition: EditionPhase2, module: "MAP-Errors", line: 1958, endLine: 1958, text: "ss-SubscriptionOption must not be used in version greater 1", kind: componentRuleParameterAbsent, role: componentError, errorCode: 19},
		{id: "p2-2075", edition: EditionPhase2, module: "MAP-MS-DataTypes", line: 2075, endLine: 2075, text: "hlr-Number must not be used in version greater 1", kind: componentRuleInteger, target: reflect.TypeFor[phase2.UpdateLocationRes](), field: "Choice", number: phase2.UpdateLocationResChoiceHlrNumber},
		{id: "p2-2296", edition: EditionPhase2, module: "MAP-MS-DataTypes", line: 2296, endLine: 2296, text: "networkResource must be absent in version greater 1", kind: componentRuleFieldAbsent, target: reflect.TypeFor[phase2.ResetArg](), field: "NetworkResource"},
		{id: "p2-1003", edition: EditionPhase2, module: "MAP-MobileServiceOperations", line: 1003, endLine: 1004, text: "UnidentifiedSubscriber must not be used in version greater 1", kind: componentRuleError, opCode: 3, errorCode: 5},
		{id: "p2-1050", edition: EditionPhase2, module: "MAP-MobileServiceOperations", line: 1050, endLine: 1050, text: "bss-APDU must be present in version greater 1", kind: componentRuleParameterPresent, role: componentArgument, opCode: 29},
		{id: "p2-1109", edition: EditionPhase2, module: "MAP-MobileServiceOperations", line: 1109, endLine: 1109, text: "UnexpectedDataValue must not be used in version greater 1", kind: componentRuleError, opCode: 43, errorCode: 36},
		{id: "p2-717", edition: EditionPhase2, module: "MAP-Protocol", line: 717, endLine: 717, text: "performHandover must not be used in version greater 1", kind: componentRuleOperation, opCode: 28},
		{id: "p2-725", edition: EditionPhase2, module: "MAP-Protocol", line: 725, endLine: 725, text: "performSubsequentHandover must not be used in version greater 1", kind: componentRuleOperation, opCode: 30},
		{id: "p2-742", edition: EditionPhase2, module: "MAP-Protocol", line: 742, endLine: 742, text: "sendParameters must not be used in version greater 1", kind: componentRuleOperation, opCode: 9},
		{id: "p2-759", edition: EditionPhase2, module: "MAP-Protocol", line: 759, endLine: 759, text: "traceSubscriberActivity must not be used in version greater 1", kind: componentRuleOperation, opCode: 52},
		{id: "p2-761", edition: EditionPhase2, module: "MAP-Protocol", line: 761, endLine: 761, text: "noteInternalHandover must not be used in version greater 1", kind: componentRuleOperation, opCode: 35},
		{id: "p2-779", edition: EditionPhase2, module: "MAP-Protocol", line: 779, endLine: 779, text: "processUnstructuredSS-Data must not be used in version greater 1", kind: componentRuleOperation, opCode: 19},
		{id: "p2-790", edition: EditionPhase2, module: "MAP-Protocol", line: 790, endLine: 790, text: "beginSubscriberActivity must not be used in version greater 1", kind: componentRuleOperation, opCode: 54},
		{id: "p2-798", edition: EditionPhase2, module: "MAP-Protocol", line: 798, endLine: 798, text: "noteSubscriberPresent must not be used in version greater 1", kind: componentRuleOperation, opCode: 48},
		{id: "p2-801", edition: EditionPhase2, module: "MAP-Protocol", line: 801, endLine: 802, text: "alertServiceCentreWithoutResult must not be used in version greater 1", kind: componentRuleOperation, opCode: 49},
		{id: "p2-2945", edition: EditionPhase2, module: "MAP-SM-DataTypes", line: 2945, endLine: 2945, text: "teleservice must be absent in version greater 1", kind: componentRuleFieldAbsent, target: reflect.TypeFor[phase2.RoutingInfoForSMArg](), field: "Teleservice"},
		{id: "p2-2952", edition: EditionPhase2, module: "MAP-SM-DataTypes", line: 2952, endLine: 2952, text: "mwd-Set must be absent in version greater 1", kind: componentRuleFieldAbsent, target: reflect.TypeFor[phase2.RoutingInfoForSMRes](), field: "MwdSet"},
		{id: "p2-2973", edition: EditionPhase2, module: "MAP-SM-DataTypes", line: 2973, endLine: 2973, text: "roaming number must not be used in version greater 1", kind: componentRuleInteger, target: reflect.TypeFor[phase2.SMRPDA](), field: "Choice", number: phase2.SMRPDAChoiceRoamingNumber},
		{id: "p2-2989", edition: EditionPhase2, module: "MAP-SM-DataTypes", line: 2989, endLine: 2989, text: "sm-DeliveryOutcome must be present in version greater 1", kind: componentRuleFieldPresent, target: reflect.TypeFor[phase2.ReportSMDeliveryStatusArg](), field: "SmDeliveryOutcome"},
		{id: "p2-2693", edition: EditionPhase2, module: "MAP-SS-DataTypes", line: 2693, endLine: 2693, text: "forwardedToNumber must not be used in version greater 1", kind: componentRuleInteger, target: reflect.TypeFor[phase2.InterrogateSSRes](), field: "Choice", number: phase2.InterrogateSSResChoiceForwardedToNumber},
		{id: "p2-2731", edition: EditionPhase2, module: "MAP-SS-DataTypes", line: 2731, endLine: 2731, text: "badPW-TryAgain must not be used in version greater 1", kind: componentRuleInteger, target: reflect.TypeFor[phase2.GuidanceInfo](), number: 3},
		{id: "p2-2733", edition: EditionPhase2, module: "MAP-SS-DataTypes", line: 2733, endLine: 2733, text: "badPW-FormatTryAgain must not be used in version greater 1", kind: componentRuleInteger, target: reflect.TypeFor[phase2.GuidanceInfo](), number: 4},
		{id: "p2-1435", edition: EditionPhase2, module: "MAP-SupplementaryServiceOperations", line: 1435, endLine: 1435, text: "UnknownSubscriber must not be used in version greater 1", kind: componentRuleError, opCode: 10, errorCode: 1},
		{id: "p2-1443", edition: EditionPhase2, module: "MAP-SupplementaryServiceOperations", line: 1443, endLine: 1444, text: "SS-SubscriptionViolation must not be used in version greater 1", kind: componentRuleError, opCode: 10, errorCode: 19},
		{id: "p2-1459", edition: EditionPhase2, module: "MAP-SupplementaryServiceOperations", line: 1459, endLine: 1459, text: "UnknownSubscriber must not be used in version greater 1", kind: componentRuleError, opCode: 11, errorCode: 1},
		{id: "p2-1469", edition: EditionPhase2, module: "MAP-SupplementaryServiceOperations", line: 1469, endLine: 1470, text: "SS-SubscriptionViolation must not be used in version greater 1", kind: componentRuleError, opCode: 11, errorCode: 19},
		{id: "p2-1486", edition: EditionPhase2, module: "MAP-SupplementaryServiceOperations", line: 1486, endLine: 1486, text: "UnknownSubscriber must not be used in version greater 1", kind: componentRuleError, opCode: 12, errorCode: 1},
		{id: "p2-1513", edition: EditionPhase2, module: "MAP-SupplementaryServiceOperations", line: 1513, endLine: 1513, text: "UnknownSubscriber must not be used in version greater 1", kind: componentRuleError, opCode: 13, errorCode: 1},
		{id: "p2-1539", edition: EditionPhase2, module: "MAP-SupplementaryServiceOperations", line: 1539, endLine: 1539, text: "UnknownSubscriber must not be used in version greater 1", kind: componentRuleError, opCode: 14, errorCode: 1},
		{id: "ts2-DataMissingParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "DataMissingParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 35},
		{id: "ts2-UnexpectedDataParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "UnexpectedDataParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 36},
		{id: "ts2-FacilityNotSupParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "FacilityNotSupParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 21},
		{id: "ts2-UnknownSubscriberParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "UnknownSubscriberParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 1},
		{id: "ts2-UnidentifiedSubParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "UunidentifiedSubParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 5},
		{id: "ts2-IllegalSubscriberParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "IllegalSubscriberParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 9},
		{id: "ts2-IllegalEquipmentParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "IllegalEquipmentParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 12},
		{id: "ts2-BearerServNotProvParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "BearerServNotProvParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 10},
		{id: "ts2-TeleservNotProvParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "TeleservNotProvParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 11},
		{id: "ts2-AbsentSubscriberParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "AbsentSubscriberParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 27},
		{id: "ts2-IllegalSSOperationParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "IllegalSS-OperationParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 16},
		{id: "ts2-SSNotAvailableParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "SS-NotAvailableParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 18},
		{id: "ts2-SSSubscriptionViolationParam", edition: EditionTS29002, module: "MAP-Errors", clause: "17.6.6", text: "SS-SubscriptionViolationParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleParameterAbsent, role: componentError, errorCode: 19},
		{id: "ts2-callBarred", edition: EditionTS29002, module: "MAP-ER-DataTypes", clause: "17.7.7", text: "extensibleCallBarredParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleInteger, target: reflect.TypeFor[mapvalues.CallBarredParam](), field: "Choice", number: mapvalues.CallBarredParamChoiceExtensibleCallBarredParam},
		{id: "ts2-systemFailure", edition: EditionTS29002, module: "MAP-ER-DataTypes", clause: "17.7.7", text: "extensibleSystemFailureParam must not be used in version <3", minVersion: 2, maxVersion: 2, kind: componentRuleInteger, target: reflect.TypeFor[mapvalues.SystemFailureParam](), field: "Choice", number: mapvalues.SystemFailureParamChoiceExtensibleSystemFailureParam},
		{id: "ts3-callBarred", edition: EditionTS29002, module: "MAP-ER-DataTypes", clause: "17.7.7; 17.1.4", text: "call BarringCause must not be used in version 3 and higher", minVersion: 3, kind: componentRuleInteger, target: reflect.TypeFor[mapvalues.CallBarredParam](), field: "Choice", number: mapvalues.CallBarredParamChoiceCallBarringCause},
		{id: "ts3-systemFailure", edition: EditionTS29002, module: "MAP-ER-DataTypes", clause: "17.7.7; 17.1.4", text: "networkResource must not be used in version 3", minVersion: 3, kind: componentRuleInteger, target: reflect.TypeFor[mapvalues.SystemFailureParam](), field: "Choice", number: mapvalues.SystemFailureParamChoiceNetworkResource},
		{id: "ts2-externalSignalInfo", edition: EditionTS29002, module: "MAP-CommonDataTypes", clause: "17.7.8", text: "extensionContainer must not be used in version 2", minVersion: 2, maxVersion: 2, kind: componentRuleFieldAbsent, target: reflect.TypeFor[mapvalues.ExternalSignalInfo](), field: "ExtensionContainer"},
	}
}

func (r componentVersionRule) applies(c ComponentCodec) bool {
	return r.edition == c.edition && c.version >= r.minVersion && (r.maxVersion == 0 || c.version <= r.maxVersion)
}

func (r componentVersionRule) violation() error {
	location := "section " + r.clause
	if r.line != 0 {
		location = fmt.Sprintf("ASN.1.txt line %d", r.line)
		if r.endLine != r.line {
			location = fmt.Sprintf("ASN.1.txt lines %d-%d", r.line, r.endLine)
		}
	}
	return fmt.Errorf("%w: %s, %s, %s: %s", ErrVersionRule, r.edition, r.module, location, r.text)
}

// Operation prohibitions are checked before role presence and ERRORS membership:
// all component forms of an operation excluded from the version are excluded.
func (c ComponentCodec) checkComponentOperation(opCode int64) error {
	for _, r := range componentVersionRules() {
		if r.kind == componentRuleOperation && r.applies(c) && r.opCode == opCode {
			return r.violation()
		}
	}
	return nil
}

func (c ComponentCodec) checkComponentRules(opCode, errorCode int64, role componentRole, value any) error {
	for _, r := range componentVersionRules() {
		if !r.applies(c) || r.kind == componentRuleOperation {
			continue
		}
		if r.target != nil {
			if r.breaksValue(reflect.ValueOf(value), make(map[uintptr]bool)) {
				return r.violation()
			}
			continue
		}
		if (r.opCode != 0 && r.opCode != opCode) || (r.errorCode != 0 && r.errorCode != errorCode) {
			continue
		}
		if r.kind == componentRuleError {
			if role == componentError {
				return r.violation()
			}
		} else if role == r.role && ((r.kind == componentRuleParameterAbsent && value != nil) ||
			(r.kind == componentRuleParameterPresent && value == nil)) {
			return r.violation()
		}
	}
	return nil
}

// Only exported value fields are visited. Raw unknown additions and private BER
// replay state do not claim a schema, so their bytes are not interpreted here.
// Pointer identities prevent a caller-supplied cyclic value from recursing forever.
func (r componentVersionRule) breaksValue(v reflect.Value, seen map[uintptr]bool) bool {
	if !v.IsValid() {
		return false
	}
	if v.Type() == r.target {
		field := v
		if r.field != "" {
			field = v.FieldByName(r.field)
		}
		switch r.kind {
		case componentRuleFieldAbsent:
			return !field.IsNil()
		case componentRuleFieldPresent:
			return field.IsNil()
		case componentRuleInteger:
			return field.Int() == r.number
		}
	}
	switch v.Kind() {
	case reflect.Pointer:
		if v.IsNil() || seen[v.Pointer()] {
			return false
		}
		seen[v.Pointer()] = true
		return r.breaksValue(v.Elem(), seen)
	case reflect.Interface:
		return r.breaksValue(v.Elem(), seen)
	case reflect.Struct:
		for i := 0; i < v.NumField(); i++ {
			field := v.Type().Field(i)
			if field.IsExported() && field.Tag.Get("asn1") != "-" && r.breaksValue(v.Field(i), seen) {
				return true
			}
		}
	case reflect.Slice:
		// No constrained MAP type is an octet. Avoid traversing opaque octets.
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return false
		}
		for i := 0; i < v.Len(); i++ {
			if r.breaksValue(v.Index(i), seen) {
				return true
			}
		}
	}
	return false
}
