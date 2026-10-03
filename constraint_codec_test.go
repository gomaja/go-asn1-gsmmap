package gsmmap

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/gomaja/go-asn1/runtime/ber"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
)

// strictWire exercises the generated BER boundary after a converter builds a
// wire value. It preserves converter errors so semantic checks remain tested.
func strictWire[T any](wire T, err error) (T, error) {
	if err != nil {
		return wire, err
	}
	var parent interface {
		MarshalBER(...ber.EncodeOption) ([]byte, error)
	}
	switch list := any(wire).(type) {
	case *gsm_map.AreaList:
		parent = &gsm_map.AreaDefinition{AreaList: list}
	case *gsm_map.EPSDataList:
		parent = &gsm_map.APNConfigurationProfile{DefaultContext: 1, EpsDataList: list}
	case *gsm_map.GPRSDataList:
		parent = &gsm_map.GPRSSubscriptionData{GprsDataList: list}
	case *gsm_map.GPRSCamelTDPDataList:
		parent = &gsm_map.GPRSCSI{GprsCamelTDPDataList: list}
	case *gsm_map.GMLCList:
		parent = &gsm_map.LCSInformation{GmlcList: list}
	case *gsm_map.LCSPrivacyExceptionList:
		parent = &gsm_map.LCSInformation{LcsPrivacyExceptionList: list}
	case *gsm_map.MOLRList:
		parent = &gsm_map.LCSInformation{MolrList: list}
	case *gsm_map.LSADataList:
		parent = &gsm_map.LSAInformation{LsaDataList: list}
	case *gsm_map.ExternalClientList:
		parent = &gsm_map.LCSPrivacyClass{SsCode: []byte{0x11}, SsStatus: []byte{0x01}, ExternalClientList: list}
	case *gsm_map.ExtExternalClientList:
		parent = &gsm_map.LCSPrivacyClass{SsCode: []byte{0x11}, SsStatus: []byte{0x01}, ExtExternalClientList: list}
	case *gsm_map.PLMNClientList:
		parent = &gsm_map.LCSPrivacyClass{SsCode: []byte{0x11}, SsStatus: []byte{0x01}, PlmnClientList: list}
	case *gsm_map.ServiceTypeList:
		parent = &gsm_map.LCSPrivacyClass{SsCode: []byte{0x11}, SsStatus: []byte{0x01}, ServiceTypeList: list}
	case *gsm_map.SpecificAPNInfoList:
		fixture := makeAPNConfiguration()
		configuration, fixtureErr := convertAPNConfigurationToWire(&fixture)
		if fixtureErr != nil {
			return wire, fixtureErr
		}
		configuration.SpecificAPNInfoList = list
		parent = configuration
	case *gsm_map.CSGSubscriptionDataList:
		parent = &gsm_map.InsertSubscriberDataArg{CsgSubscriptionDataList: list}
	case *gsm_map.VPLMNCSGSubscriptionDataList:
		parent = &gsm_map.InsertSubscriberDataArg{VplmnCsgSubscriptionDataList: list}
	case *gsm_map.AdjacentAccessRestrictionDataList:
		parent = &gsm_map.InsertSubscriberDataArg{AdjacentAccessRestrictionDataList: list}
	case *gsm_map.EDRXCycleLengthList:
		parent = &gsm_map.InsertSubscriberDataArg{EDRXCycleLengthList: list}
	case *gsm_map.ResetIdList:
		parent = &gsm_map.InsertSubscriberDataArg{ResetIdList: list}
	case *gsm_map.IMSIGroupIdList:
		parent = &gsm_map.InsertSubscriberDataArg{ImsiGroupIdList: list}
	case *gsm_map.ZoneCodeList:
		parent = &gsm_map.InsertSubscriberDataArg{RegionalSubscriptionData: list}
	case *gsm_map.VBSDataList:
		parent = &gsm_map.InsertSubscriberDataArg{VbsSubscriptionData: list}
	case *gsm_map.VGCSDataList:
		parent = &gsm_map.InsertSubscriberDataArg{VgcsSubscriptionData: list}
	default:
		parent, _ = any(wire).(interface {
			MarshalBER(...ber.EncodeOption) ([]byte, error)
		})
		if parent == nil {
			value := reflect.ValueOf(wire)
			if value.Kind() != reflect.Pointer {
				pointer := reflect.New(value.Type())
				pointer.Elem().Set(value)
				parent, _ = pointer.Interface().(interface {
					MarshalBER(...ber.EncodeOption) ([]byte, error)
				})
			}
		}
	}
	if parent == nil {
		return wire, fmt.Errorf("%T has no BER encoder", wire)
	}
	_, err = parent.MarshalBER()
	return wire, err
}

func matchesConstraint(err error, path, constraint string) bool {
	if path == "" || constraint == "" {
		return false
	}
	var ce *ber.ConstraintError
	return errors.As(err, &ce) && ce.Path == path && ce.Constraint == constraint
}

func matchesExpected(err, want error, path, constraint string) bool {
	if want == nil {
		if path == "" || constraint == "" {
			return false
		}
		return matchesConstraint(err, path, constraint)
	}
	return errors.Is(err, want)
}

// strictDecodeWire makes a constraint-invalid BER value with the codec's
// tolerance option, then checks that strict decode rejects the same value.
func strictDecodeWire(wire any) error {
	encoder, ok := wire.(interface {
		MarshalBER(...ber.EncodeOption) ([]byte, error)
	})
	if !ok {
		return fmt.Errorf("%T has no BER encoder", wire)
	}
	data, err := encoder.MarshalBER(ber.WithConstraintTolerance(&ber.ViolationLog{}))
	if err != nil {
		return err
	}
	type decoder interface {
		UnmarshalBER([]byte, ...ber.DecodeOption) error
	}
	value := reflect.ValueOf(wire)
	if value.Kind() != reflect.Pointer {
		return fmt.Errorf("%T is not a pointer", wire)
	}
	decoded, ok := reflect.New(value.Elem().Type()).Interface().(decoder)
	if !ok {
		return fmt.Errorf("%T has no BER decoder", wire)
	}
	return decoded.UnmarshalBER(data)
}
