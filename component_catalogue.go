package gsmmap

import (
	"reflect"
	"sync"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/bindings"
	phase1 "github.com/gomaja/go-asn1/telecom/ss7/gsm_map/gsm0902_v3_11_0"
	phase2 "github.com/gomaja/go-asn1/telecom/ss7/gsm_map/gsm0902_v4_19_1"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

// These native role types include aliases and the built-in BOOLEAN. Names alone
// cannot identify their Go dynamic types (for example SendAuthenticationInfoArg
// is an alias of IMSI). Keep the type witnesses alongside the catalogue adapters.
func componentPhase1Types() map[string]reflect.Type {
	return map[string]reflect.Type{
		"ActivateSSArgument":                   componentCodecValueType(phase1.ActivateSSArgumentCodec()),
		"ActivateTraceModeArgument":            componentCodecValueType(phase1.ActivateTraceModeArgumentCodec()),
		"AlertServiceCentreArgument":           componentCodecValueType(phase1.AlertServiceCentreArgumentCodec()),
		"AuthenticateArgument":                 componentCodecValueType(phase1.AuthenticateArgumentCodec()),
		"BeginSubscriberActivityArgument":      componentCodecValueType(phase1.BeginSubscriberActivityArgumentCodec()),
		"CUGInterlock":                         componentCodecValueType(phase1.CUGInterlockCodec()),
		"CUGRejectParameter":                   componentCodecValueType(phase1.CUGRejectParameterCodec()),
		"CallBarredParameter":                  componentCodecValueType(phase1.CallBarredParameterCodec()),
		"CancelLocationArgument":               componentCodecValueType(phase1.CancelLocationArgumentCodec()),
		"CompleteCallArgument":                 componentCodecValueType(phase1.CompleteCallArgumentCodec()),
		"ConnectToFollowingAddressArgument":    componentCodecValueType(phase1.ConnectToFollowingAddressArgumentCodec()),
		"DeactivateSSArgument":                 componentCodecValueType(phase1.DeactivateSSArgumentCodec()),
		"DeactivateTraceModeArgument":          componentCodecValueType(phase1.DeactivateTraceModeArgumentCodec()),
		"DeleteSubscriberDataArgument":         componentCodecValueType(phase1.DeleteSubscriberDataArgumentCodec()),
		"EquipStatus":                          componentCodecValueType(phase1.EquipStatusCodec()),
		"EraseSSArgument":                      componentCodecValueType(phase1.EraseSSArgumentCodec()),
		"ExternalSignalInfo":                   componentCodecValueType(phase1.ExternalSignalInfoCodec()),
		"ForwardShortMessageArgument":          componentCodecValueType(phase1.ForwardShortMessageArgumentCodec()),
		"ForwardSsNotificationArgument":        componentCodecValueType(phase1.ForwardSsNotificationArgumentCodec()),
		"GuidanceInfo":                         componentCodecValueType(phase1.GuidanceInfoCodec()),
		"IMEI":                                 componentCodecValueType(phase1.IMEICodec()),
		"IMSI":                                 componentCodecValueType(phase1.IMSICodec()),
		"InsertSubscriberDataArgument":         componentCodecValueType(phase1.InsertSubscriberDataArgumentCodec()),
		"InterrogateSSArgument":                componentCodecValueType(phase1.InterrogateSSArgumentCodec()),
		"InterrogateSSResult":                  componentCodecValueType(phase1.InterrogateSSResultCodec()),
		"InvokeSSArgument":                     componentCodecValueType(phase1.InvokeSSArgumentCodec()),
		"IsdnAddressString":                    componentCodecValueType(phase1.IsdnAddressStringCodec()),
		"LocAreaId":                            componentCodecValueType(phase1.LocAreaIdCodec()),
		"NetworkResource":                      componentCodecValueType(phase1.NetworkResourceCodec()),
		"NoteInternalHandoverArgument":         componentCodecValueType(phase1.NoteInternalHandoverArgumentCodec()),
		"PageArgument":                         componentCodecValueType(phase1.PageArgumentCodec()),
		"Password":                             componentCodecValueType(phase1.PasswordCodec()),
		"PasswordRegistrationFailureParameter": componentCodecValueType(phase1.PasswordRegistrationFailureParameterCodec()),
		"PerformHandoverArgument":              componentCodecValueType(phase1.PerformHandoverArgumentCodec()),
		"PerformHandoverResult":                componentCodecValueType(phase1.PerformHandoverResultCodec()),
		"PerformSubsequentHandoverArgument":    componentCodecValueType(phase1.PerformSubsequentHandoverArgumentCodec()),
		"ProcessAccessRequestArgument":         componentCodecValueType(phase1.ProcessAccessRequestArgumentCodec()),
		"ProcessAccessRequestResult":           componentCodecValueType(phase1.ProcessAccessRequestResultCodec()),
		"ProvideRoamingNumberArgument":         componentCodecValueType(phase1.ProvideRoamingNumberArgumentCodec()),
		"RegisterChargingInformationArgument":  componentCodecValueType(phase1.RegisterChargingInformationArgumentCodec()),
		"RegisterSSArgument":                   componentCodecValueType(phase1.RegisterSSArgumentCodec()),
		"ResetArgument":                        componentCodecValueType(phase1.ResetArgumentCodec()),
		"SMDeliveryFailureParameter":           componentCodecValueType(phase1.SMDeliveryFailureParameterCodec()),
		"SSCode":                               componentCodecValueType(phase1.SSCodeCodec()),
		"SSCodeList":                           componentCodecValueType(phase1.SSCodeListCodec()),
		"SSIncompatibilityParameter":           componentCodecValueType(phase1.SSIncompatibilityParameterCodec()),
		"SSInformation":                        componentCodecValueType(phase1.SSInformationCodec()),
		"SSStatus":                             componentCodecValueType(phase1.SSStatusCodec()),
		"SSSubscriptionOption":                 componentCodecValueType(phase1.SSSubscriptionOptionCodec()),
		"SSUserData":                           componentCodecValueType(phase1.SSUserDataCodec()),
		"SearchForMobileSubscriberArgument":    componentCodecValueType(phase1.SearchForMobileSubscriberArgumentCodec()),
		"SendInfoForIncomingCallArgument":      componentCodecValueType(phase1.SendInfoForIncomingCallArgumentCodec()),
		"SendInfoForOutgoingCallArgument":      componentCodecValueType(phase1.SendInfoForOutgoingCallArgumentCodec()),
		"SendParametersArgument":               componentCodecValueType(phase1.SendParametersArgumentCodec()),
		"SendRoutingInfoForSMArgument":         componentCodecValueType(phase1.SendRoutingInfoForSMArgumentCodec()),
		"SendRoutingInfoForSMResult":           componentCodecValueType(phase1.SendRoutingInfoForSMResultCodec()),
		"SendRoutingInformationArgument":       componentCodecValueType(phase1.SendRoutingInformationArgumentCodec()),
		"SendRoutingInformationResult":         componentCodecValueType(phase1.SendRoutingInformationResultCodec()),
		"SentParameters":                       componentCodecValueType(phase1.SentParametersCodec()),
		"SetCipheringModeArgument":             componentCodecValueType(phase1.SetCipheringModeArgumentCodec()),
		"SetMessageWaitingDataArgument":        componentCodecValueType(phase1.SetMessageWaitingDataArgumentCodec()),
		"Sres":                                 componentCodecValueType(phase1.SresCodec()),
		"SubscriberId":                         componentCodecValueType(phase1.SubscriberIdCodec()),
		"TMSI":                                 componentCodecValueType(phase1.TMSICodec()),
		"TraceSubscriberActivityArgument":      componentCodecValueType(phase1.TraceSubscriberActivityArgumentCodec()),
		"UpdateLocationAreaArgument":           componentCodecValueType(phase1.UpdateLocationAreaArgumentCodec()),
		"UpdateLocationArgument":               componentCodecValueType(phase1.UpdateLocationArgumentCodec()),
		"bool":                                 reflect.TypeFor[phase1.BERValue[bool]](),
	}
}

func componentPhase2Types() map[string]reflect.Type {
	return map[string]reflect.Type{
		"ActivateTraceModeArg":       componentCodecValueType(phase2.ActivateTraceModeArgCodec()),
		"AlertServiceCentreArg":      componentCodecValueType(phase2.AlertServiceCentreArgCodec()),
		"BeginSubscriberActivityArg": componentCodecValueType(phase2.BeginSubscriberActivityArgCodec()),
		"CUGRejectCause":             componentCodecValueType(phase2.CUGRejectCauseCodec()),
		"CallBarringCause":           componentCodecValueType(phase2.CallBarringCauseCodec()),
		"CancelLocationArg":          componentCodecValueType(phase2.CancelLocationArgCodec()),
		"DeactivateTraceModeArg":     componentCodecValueType(phase2.DeactivateTraceModeArgCodec()),
		"DeleteSubscriberDataArg":    componentCodecValueType(phase2.DeleteSubscriberDataArgCodec()),
		"DeleteSubscriberDataRes":    componentCodecValueType(phase2.DeleteSubscriberDataResCodec()),
		"EquipmentStatus":            componentCodecValueType(phase2.EquipmentStatusCodec()),
		"ExternalSignalInfo":         componentCodecValueType(phase2.ExternalSignalInfoCodec()),
		"ForwardSMArg":               componentCodecValueType(phase2.ForwardSMArgCodec()),
		"GuidanceInfo":               componentCodecValueType(phase2.GuidanceInfoCodec()),
		"IMEI":                       componentCodecValueType(phase2.IMEICodec()),
		"IMSI":                       componentCodecValueType(phase2.IMSICodec()),
		"ISDNAddressString":          componentCodecValueType(phase2.ISDNAddressStringCodec()),
		"InformServiceCentreArg":     componentCodecValueType(phase2.InformServiceCentreArgCodec()),
		"InsertSubscriberDataArg":    componentCodecValueType(phase2.InsertSubscriberDataArgCodec()),
		"InsertSubscriberDataRes":    componentCodecValueType(phase2.InsertSubscriberDataResCodec()),
		"InterrogateSSRes":           componentCodecValueType(phase2.InterrogateSSResCodec()),
		"NetworkResource":            componentCodecValueType(phase2.NetworkResourceCodec()),
		"NoteInternalHOArg":          componentCodecValueType(phase2.NoteInternalHOArgCodec()),
		"PWRegistrationFailureCause": componentCodecValueType(phase2.PWRegistrationFailureCauseCodec()),
		"Password":                   componentCodecValueType(phase2.PasswordCodec()),
		"PerformHOArg":               componentCodecValueType(phase2.PerformHOArgCodec()),
		"PerformHORes":               componentCodecValueType(phase2.PerformHOResCodec()),
		"PerformSubsequentHOArg":     componentCodecValueType(phase2.PerformSubsequentHOArgCodec()),
		"PrepareHOArg":               componentCodecValueType(phase2.PrepareHOArgCodec()),
		"PrepareHORes":               componentCodecValueType(phase2.PrepareHOResCodec()),
		"PrepareSubsequentHOArg":     componentCodecValueType(phase2.PrepareSubsequentHOArgCodec()),
		"ProvideRoamingNumberArg":    componentCodecValueType(phase2.ProvideRoamingNumberArgCodec()),
		"PurgeMSArg":                 componentCodecValueType(phase2.PurgeMSArgCodec()),
		"ReadyForSMArg":              componentCodecValueType(phase2.ReadyForSMArgCodec()),
		"RegisterSSArg":              componentCodecValueType(phase2.RegisterSSArgCodec()),
		"ReportSMDeliveryStatusArg":  componentCodecValueType(phase2.ReportSMDeliveryStatusArgCodec()),
		"ResetArg":                   componentCodecValueType(phase2.ResetArgCodec()),
		"RestoreDataArg":             componentCodecValueType(phase2.RestoreDataArgCodec()),
		"RestoreDataRes":             componentCodecValueType(phase2.RestoreDataResCodec()),
		"RoamingNotAllowedCause":     componentCodecValueType(phase2.RoamingNotAllowedCauseCodec()),
		"RoutingInfoForSMArg":        componentCodecValueType(phase2.RoutingInfoForSMArgCodec()),
		"RoutingInfoForSMRes":        componentCodecValueType(phase2.RoutingInfoForSMResCodec()),
		"SMDeliveryFailureCause":     componentCodecValueType(phase2.SMDeliveryFailureCauseCodec()),
		"SSCode":                     componentCodecValueType(phase2.SSCodeCodec()),
		"SSForBSCode":                componentCodecValueType(phase2.SSForBSCodeCodec()),
		"SSIncompatibilityCause":     componentCodecValueType(phase2.SSIncompatibilityCauseCodec()),
		"SSInfo":                     componentCodecValueType(phase2.SSInfoCodec()),
		"SSStatus":                   componentCodecValueType(phase2.SSStatusCodec()),
		"SSSubscriptionOption":       componentCodecValueType(phase2.SSSubscriptionOptionCodec()),
		"SSUserData":                 componentCodecValueType(phase2.SSUserDataCodec()),
		"SendAuthenticationInfoArg":  componentCodecValueType(phase2.SendAuthenticationInfoArgCodec()),
		"SendAuthenticationInfoRes":  componentCodecValueType(phase2.SendAuthenticationInfoResCodec()),
		"SendIdentificationRes":      componentCodecValueType(phase2.SendIdentificationResCodec()),
		"SendParametersArg":          componentCodecValueType(phase2.SendParametersArgCodec()),
		"SendRoutingInfoArg":         componentCodecValueType(phase2.SendRoutingInfoArgCodec()),
		"SendRoutingInfoRes":         componentCodecValueType(phase2.SendRoutingInfoResCodec()),
		"SentParameterList":          componentCodecValueType(phase2.SentParameterListCodec()),
		"TMSI":                       componentCodecValueType(phase2.TMSICodec()),
		"TraceSubscriberActivityArg": componentCodecValueType(phase2.TraceSubscriberActivityArgCodec()),
		"USSDArg":                    componentCodecValueType(phase2.USSDArgCodec()),
		"USSDRes":                    componentCodecValueType(phase2.USSDResCodec()),
		"UpdateLocationArg":          componentCodecValueType(phase2.UpdateLocationArgCodec()),
		"UpdateLocationRes":          componentCodecValueType(phase2.UpdateLocationResCodec()),
		"bool":                       reflect.TypeFor[phase2.BERValue[bool]](),
	}
}

// The generic codec fixes T to the actual decoder return type, including BERValue.
func componentCodecValueType[T any](codec interface {
	DecodeBER([]byte, ...ber.DecodeOption) (T, error)
}) reflect.Type {
	return reflect.TypeFor[T]()
}

type componentErrorDefinition struct {
	module, definition string
	code               int64
}

// DefinitionModule/Definition is already the resolved target for both a macro
// reference and a value reference (IsValue). Name is the spelling at the use site.
// Ambiguous or missing targets fail closed; tests pin every catalogue reference.
func componentResolveError(module, definition string, definitions []componentErrorDefinition) (int64, bool) {
	var code int64
	matches := 0
	for _, e := range definitions {
		if e.module == module && e.definition == definition {
			code = e.code
			matches++
		}
	}
	return code, matches == 1
}

func componentPhase1Type(t *phase1.MacroType, types map[string]reflect.Type) *componentType {
	if t == nil {
		return nil
	}
	return &componentType{name: t.ASN1Name, valueType: types[t.GoName], decode: t.DecodeBER, encode: t.EncodeBER}
}

func componentPhase1Catalogue() componentCatalogue {
	cat := componentCatalogue{operations: make(map[int64]componentOperation), errors: make(map[int64]componentErrorInfo)}
	types := componentPhase1Types()
	var definitions []componentErrorDefinition
	for _, e := range phase1.MacroErrors() {
		cat.errors[e.Code] = componentErrorInfo{name: e.Name, parameter: componentPhase1Type(e.Parameter, types)}
		definitions = append(definitions, componentErrorDefinition{e.Module, e.Definition, e.Code})
	}
	for _, op := range phase1.MacroOperations() {
		errors := make(map[int64]bool)
		for _, ref := range op.Errors {
			if code, ok := componentResolveError(ref.DefinitionModule, ref.Definition, definitions); ok {
				errors[code] = true
			}
		}
		cat.operations[op.Code] = componentOperation{name: op.Name, argument: componentPhase1Type(op.Argument, types), result: componentPhase1Type(op.Result, types), resultPresent: op.ResultPresent, errors: errors}
	}
	return cat
}

func componentPhase2Type(t *phase2.MacroType, types map[string]reflect.Type) *componentType {
	if t == nil {
		return nil
	}
	return &componentType{name: t.ASN1Name, valueType: types[t.GoName], decode: t.DecodeBER, encode: t.EncodeBER}
}

func componentPhase2Catalogue() componentCatalogue {
	cat := componentCatalogue{operations: make(map[int64]componentOperation), errors: make(map[int64]componentErrorInfo)}
	types := componentPhase2Types()
	var definitions []componentErrorDefinition
	for _, e := range phase2.MacroErrors() {
		cat.errors[e.Code] = componentErrorInfo{name: e.Name, parameter: componentPhase2Type(e.Parameter, types)}
		definitions = append(definitions, componentErrorDefinition{e.Module, e.Definition, e.Code})
	}
	for _, op := range phase2.MacroOperations() {
		errors := make(map[int64]bool)
		for _, ref := range op.Errors {
			if code, ok := componentResolveError(ref.DefinitionModule, ref.Definition, definitions); ok {
				errors[code] = true
			}
		}
		cat.operations[op.Code] = componentOperation{name: op.Name, argument: componentPhase2Type(op.Argument, types), result: componentPhase2Type(op.Result, types), resultPresent: op.ResultPresent, errors: errors}
	}
	return cat
}

func componentTypedCodec[T any](name string, codec gsm_map.ValueCodec[T]) *componentType {
	return &componentType{name: name, valueType: reflect.TypeFor[T](), decode: func(data []byte, opts ...ber.DecodeOption) (any, error) {
		v, err := codec.DecodeBER(data, opts...)
		if err != nil {
			return nil, err
		}
		return v, nil
	}, encode: func(value any, opts ...ber.EncodeOption) ([]byte, error) { return codec.EncodeBER(value.(T), opts...) }}
}

func componentTSOperation[A, R, E any](binding bindings.Operation[A, R, E]) (int64, componentOperation) {
	info := binding.Info()
	arg, hasArg := binding.Argument()
	res, hasRes := binding.Result()
	var argument, result *componentType
	if hasArg {
		argument = componentTypedCodec(info.Argument.ASN1Name, arg)
	}
	if hasRes {
		result = componentTypedCodec(info.Result.ASN1Name, res)
	}
	errors := make(map[int64]bool)
	for _, e := range info.Errors {
		errors[int64(e.Code)] = true
	}
	return int64(info.Code), componentOperation{name: info.Name, argument: argument, result: result, resultPresent: info.ReturnResult, errors: errors}
}

func componentTSError[P any](binding bindings.ErrorDescriptor[P]) (int64, componentErrorInfo) {
	info := binding.Info()
	codec, present := binding.Parameter()
	var parameter *componentType
	if present {
		parameter = componentTypedCodec(info.Parameter.ASN1Name, codec)
	}
	return int64(info.Code), componentErrorInfo{name: info.Name, parameter: parameter}
}

func componentTSCatalogue() componentCatalogue {
	cat := componentCatalogue{operations: make(map[int64]componentOperation), errors: make(map[int64]componentErrorInfo)}
	addOp := func(code int64, op componentOperation) { cat.operations[code] = op }
	addError := func(code int64, e componentErrorInfo) { cat.errors[code] = e }
	addOp(componentTSOperation(bindings.ActivateSSOperation()))
	addOp(componentTSOperation(bindings.ActivateTraceModeOperation()))
	addOp(componentTSOperation(bindings.AlertServiceCentreOperation()))
	addOp(componentTSOperation(bindings.AnyTimeInterrogationOperation()))
	addOp(componentTSOperation(bindings.AnyTimeModificationOperation()))
	addOp(componentTSOperation(bindings.AnyTimeSubscriptionInterrogationOperation()))
	addOp(componentTSOperation(bindings.AuthenticationFailureReportOperation()))
	addOp(componentTSOperation(bindings.CancelLocationOperation()))
	addOp(componentTSOperation(bindings.CancelVcsgLocationOperation()))
	addOp(componentTSOperation(bindings.CheckIMEIOperation()))
	addOp(componentTSOperation(bindings.DeactivateSSOperation()))
	addOp(componentTSOperation(bindings.DeactivateTraceModeOperation()))
	addOp(componentTSOperation(bindings.DeleteSubscriberDataOperation()))
	addOp(componentTSOperation(bindings.EraseCCEntryOperation()))
	addOp(componentTSOperation(bindings.EraseSSOperation()))
	addOp(componentTSOperation(bindings.FailureReportOperation()))
	addOp(componentTSOperation(bindings.ForwardAccessSignallingOperation()))
	addOp(componentTSOperation(bindings.ForwardCheckSSIndicationOperation()))
	addOp(componentTSOperation(bindings.ForwardGroupCallSignallingOperation()))
	addOp(componentTSOperation(bindings.GetPasswordOperation()))
	addOp(componentTSOperation(bindings.InformServiceCentreOperation()))
	addOp(componentTSOperation(bindings.InsertSubscriberDataOperation()))
	addOp(componentTSOperation(bindings.InterrogateSSOperation()))
	addOp(componentTSOperation(bindings.IstAlertOperation()))
	addOp(componentTSOperation(bindings.IstCommandOperation()))
	addOp(componentTSOperation(bindings.MoForwardSMOperation()))
	addOp(componentTSOperation(bindings.MtForwardSMOperation()))
	addOp(componentTSOperation(bindings.MtForwardSMVGCSOperation()))
	addOp(componentTSOperation(bindings.NoteMMEventOperation()))
	addOp(componentTSOperation(bindings.NoteMsPresentForGprsOperation()))
	addOp(componentTSOperation(bindings.NoteSubscriberDataModifiedOperation()))
	addOp(componentTSOperation(bindings.PrepareGroupCallOperation()))
	addOp(componentTSOperation(bindings.PrepareHandoverOperation()))
	addOp(componentTSOperation(bindings.PrepareSubsequentHandoverOperation()))
	addOp(componentTSOperation(bindings.ProcessAccessSignallingOperation()))
	addOp(componentTSOperation(bindings.ProcessGroupCallSignallingOperation()))
	addOp(componentTSOperation(bindings.ProcessUnstructuredSSRequestOperation()))
	addOp(componentTSOperation(bindings.ProvideRoamingNumberOperation()))
	addOp(componentTSOperation(bindings.ProvideSubscriberInfoOperation()))
	addOp(componentTSOperation(bindings.ProvideSubscriberLocationOperation()))
	addOp(componentTSOperation(bindings.PurgeMSOperation()))
	addOp(componentTSOperation(bindings.ReadyForSMOperation()))
	addOp(componentTSOperation(bindings.RegisterCCEntryOperation()))
	addOp(componentTSOperation(bindings.RegisterPasswordOperation()))
	addOp(componentTSOperation(bindings.RegisterSSOperation()))
	addOp(componentTSOperation(bindings.ReleaseResourcesOperation()))
	addOp(componentTSOperation(bindings.RemoteUserFreeOperation()))
	addOp(componentTSOperation(bindings.ReportSMDeliveryStatusOperation()))
	addOp(componentTSOperation(bindings.ResetOperation()))
	addOp(componentTSOperation(bindings.RestoreDataOperation()))
	addOp(componentTSOperation(bindings.ResumeCallHandlingOperation()))
	addOp(componentTSOperation(bindings.SendAuthenticationInfoOperation()))
	addOp(componentTSOperation(bindings.SendEndSignalOperation()))
	addOp(componentTSOperation(bindings.SendGroupCallEndSignalOperation()))
	addOp(componentTSOperation(bindings.SendGroupCallInfoOperation()))
	addOp(componentTSOperation(bindings.SendIMSIOperation()))
	addOp(componentTSOperation(bindings.SendIdentificationOperation()))
	addOp(componentTSOperation(bindings.SendRoutingInfoOperation()))
	addOp(componentTSOperation(bindings.SendRoutingInfoForGprsOperation()))
	addOp(componentTSOperation(bindings.SendRoutingInfoForLCSOperation()))
	addOp(componentTSOperation(bindings.SendRoutingInfoForSMOperation()))
	addOp(componentTSOperation(bindings.SetReportingStateOperation()))
	addOp(componentTSOperation(bindings.SsInvocationNotificationOperation()))
	addOp(componentTSOperation(bindings.StatusReportOperation()))
	addOp(componentTSOperation(bindings.SubscriberLocationReportOperation()))
	addOp(componentTSOperation(bindings.UnstructuredSSNotifyOperation()))
	addOp(componentTSOperation(bindings.UnstructuredSSRequestOperation()))
	addOp(componentTSOperation(bindings.UpdateGprsLocationOperation()))
	addOp(componentTSOperation(bindings.UpdateLocationOperation()))
	addOp(componentTSOperation(bindings.UpdateVcsgLocationOperation()))
	addError(componentTSError(bindings.AbsentSubscriberErrorDescriptor()))
	addError(componentTSError(bindings.AbsentSubscriberSMErrorDescriptor()))
	addError(componentTSError(bindings.AtiNotAllowedErrorDescriptor()))
	addError(componentTSError(bindings.AtmNotAllowedErrorDescriptor()))
	addError(componentTSError(bindings.AtsiNotAllowedErrorDescriptor()))
	addError(componentTSError(bindings.BearerServiceNotProvisionedErrorDescriptor()))
	addError(componentTSError(bindings.BusySubscriberErrorDescriptor()))
	addError(componentTSError(bindings.CallBarredErrorDescriptor()))
	addError(componentTSError(bindings.CugRejectErrorDescriptor()))
	addError(componentTSError(bindings.DataMissingErrorDescriptor()))
	addError(componentTSError(bindings.FacilityNotSupportedErrorDescriptor()))
	addError(componentTSError(bindings.ForwardingFailedErrorDescriptor()))
	addError(componentTSError(bindings.ForwardingViolationErrorDescriptor()))
	addError(componentTSError(bindings.IllegalEquipmentErrorDescriptor()))
	addError(componentTSError(bindings.IllegalSSOperationErrorDescriptor()))
	addError(componentTSError(bindings.IllegalSubscriberErrorDescriptor()))
	addError(componentTSError(bindings.IncompatibleTerminalErrorDescriptor()))
	addError(componentTSError(bindings.InformationNotAvailableErrorDescriptor()))
	addError(componentTSError(bindings.LongTermDenialErrorDescriptor()))
	addError(componentTSError(bindings.MessageWaitingListFullErrorDescriptor()))
	addError(componentTSError(bindings.MmEventNotSupportedErrorDescriptor()))
	addError(componentTSError(bindings.NegativePWCheckErrorDescriptor()))
	addError(componentTSError(bindings.NoGroupCallNumberAvailableErrorDescriptor()))
	addError(componentTSError(bindings.NoHandoverNumberAvailableErrorDescriptor()))
	addError(componentTSError(bindings.NoRoamingNumberAvailableErrorDescriptor()))
	addError(componentTSError(bindings.NoSubscriberReplyErrorDescriptor()))
	addError(componentTSError(bindings.NumberChangedErrorDescriptor()))
	addError(componentTSError(bindings.NumberOfPWAttemptsViolationErrorDescriptor()))
	addError(componentTSError(bindings.OngoingGroupCallErrorDescriptor()))
	addError(componentTSError(bindings.OrNotAllowedErrorDescriptor()))
	addError(componentTSError(bindings.PositionMethodFailureErrorDescriptor()))
	addError(componentTSError(bindings.PwRegistrationFailureErrorDescriptor()))
	addError(componentTSError(bindings.ResourceLimitationErrorDescriptor()))
	addError(componentTSError(bindings.RoamingNotAllowedErrorDescriptor()))
	addError(componentTSError(bindings.ShortTermDenialErrorDescriptor()))
	addError(componentTSError(bindings.SmDeliveryFailureErrorDescriptor()))
	addError(componentTSError(bindings.SsErrorStatusErrorDescriptor()))
	addError(componentTSError(bindings.SsIncompatibilityErrorDescriptor()))
	addError(componentTSError(bindings.SsNotAvailableErrorDescriptor()))
	addError(componentTSError(bindings.SsSubscriptionViolationErrorDescriptor()))
	addError(componentTSError(bindings.SubscriberBusyForMTSMSErrorDescriptor()))
	addError(componentTSError(bindings.SubsequentHandoverFailureErrorDescriptor()))
	addError(componentTSError(bindings.SystemFailureErrorDescriptor()))
	addError(componentTSError(bindings.TargetCellOutsideGroupCallAreaErrorDescriptor()))
	addError(componentTSError(bindings.TeleserviceNotProvisionedErrorDescriptor()))
	addError(componentTSError(bindings.TracingBufferFullErrorDescriptor()))
	addError(componentTSError(bindings.UnauthorizedLCSClientErrorDescriptor()))
	addError(componentTSError(bindings.UnauthorizedRequestingNetworkErrorDescriptor()))
	addError(componentTSError(bindings.UnexpectedDataValueErrorDescriptor()))
	addError(componentTSError(bindings.UnidentifiedSubscriberErrorDescriptor()))
	addError(componentTSError(bindings.UnknownAlphabetErrorDescriptor()))
	addError(componentTSError(bindings.UnknownEquipmentErrorDescriptor()))
	addError(componentTSError(bindings.UnknownMSCErrorDescriptor()))
	addError(componentTSError(bindings.UnknownOrUnreachableLCSClientErrorDescriptor()))
	addError(componentTSError(bindings.UnknownSubscriberErrorDescriptor()))
	addError(componentTSError(bindings.UssdBusyErrorDescriptor()))
	return cat
}

var (
	componentCataloguesOnce  sync.Once
	componentCataloguesValue [EditionTS29002 + 1]componentCatalogue
)

// componentCatalogues builds the three catalogues on first use. Only the codec's
// methods reach it, so the linker can drop the catalogues and the edition codecs
// they reference from a program that never uses the codec.
func componentCatalogues() *[EditionTS29002 + 1]componentCatalogue {
	componentCataloguesOnce.Do(func() {
		componentCataloguesValue = [...]componentCatalogue{
			EditionPhase1:  componentPhase1Catalogue(),
			EditionPhase2:  componentPhase2Catalogue(),
			EditionTS29002: componentTSCatalogue(),
		}
	})
	return &componentCataloguesValue
}
