package gsmmap

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	gsm_map "github.com/gomaja/go-asn1/telecom/ss7/gsm_map"
	asn1runtime "github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime"
	"github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"
)

// Edition identifies the ASN.1 source of a MAP dialogue's component parameters.
type Edition uint8

const (
	// EditionPhase1 is GSM 09.02 V3.11.0, the final Phase 1 edition, for
	// application contexts of version 1.
	EditionPhase1 Edition = 1
	// EditionPhase2 is GSM 09.02 V4.19.1, the final Phase 2 edition, for
	// version 2 except the families EditionTS29002 lists.
	EditionPhase2 Edition = 2
	// EditionTS29002 is 3GPP TS 29.002 V19.1.0, for version 3 and later, and for
	// version 2 of networkFunctionalSs, networkUnstructuredSs, shortMsgAlert and
	// imsiRetrieval, which have no higher version (TS 29.002 §17.3.3).
	EditionTS29002 Edition = 3
)

// String returns the edition's document and version.
func (e Edition) String() string {
	switch e {
	case EditionPhase1:
		return "GSM 09.02 V3.11.0"
	case EditionPhase2:
		return "GSM 09.02 V4.19.1"
	case EditionTS29002:
		return "3GPP TS 29.002 V19.1.0"
	default:
		return fmt.Sprintf("Edition(%d)", uint8(e))
	}
}

// ApplicationContextVersion returns the version arc of a canonical MAP
// application context name {map-ac ac(n) version(v)} (TS 29.002 §17.3.3).
// It checks the name's form, not whether the context is assigned.
func ApplicationContextVersion(acn OID) (int, error) {
	_, version, err := componentContextName(acn)
	return version, err
}

func componentContextName(acn OID) (string, int, error) {
	if _, err := asn1runtime.ParseObjectIdentifier(string(acn)); err != nil {
		return "", 0, fmt.Errorf("%w: %q: %w", ErrNotMAPApplicationContext, acn, err)
	}
	arcs := strings.Split(string(acn), ".")
	if len(arcs) != 8 || strings.Join(arcs[:6], ".") != "0.4.0.0.1.0" {
		return "", 0, fmt.Errorf("%w: %q", ErrNotMAPApplicationContext, acn)
	}
	version, err := strconv.Atoi(arcs[7])
	if err != nil || version < 1 {
		return "", 0, fmt.Errorf("%w: invalid version in %q", ErrNotMAPApplicationContext, acn)
	}
	return arcs[6], version, nil
}

// ComponentCodec decodes and encodes the parameters of one dialogue's components.
// Values are the selected go-asn1 edition's generated values, not pointers.
// Decoding uses that edition's strict BER profile; encoding disables BER replay.
// V4.19.1 and TS 29.002 values keep unknown extension additions in ExtData_,
// and a fresh encoding copies them only if they pass go-asn1's DER element
// checks: an addition that needs schema-dependent repair parses but makes
// Marshal fail with ber.ErrInvalidValue. Clearing ExtData_ drops the additions.
//
// The codec checks ERRORS membership and the ASN.1 version comments. The caller
// checks application context membership, dialogue state, semantic parameter
// presence (TS 29.002 §17.1.2; V4.19.1 §14.1.2), PLMN extension restrictions,
// LINKED operations, and DialoguePDU's separate version rules.
// A zero codec's parameter methods return ErrComponentCodecUnset.
type ComponentCodec struct {
	edition Edition
	version int
}

// ComponentCodecFor returns the codec for the dialogue's application context
// name. An empty name selects version 1 (TS 29.002 §14.4). Other names must have
// exactly the form {0 4 0 0 1 0 family version}, with version at least 1.
// Version 2 uses Phase 2's v2-equivalent operations (TS 29.002 §17.2.1), except
// families whose highest defined version is 2 (§17.3.3); those keep TS 29.002.
// Neither assignment of the context nor operation membership is checked.
func ComponentCodecFor(acn OID) (ComponentCodec, error) {
	if acn == "" {
		return ComponentCodec{edition: EditionPhase1, version: 1}, nil
	}
	family, version, err := componentContextName(acn)
	if err != nil {
		return ComponentCodec{}, err
	}
	edition := EditionTS29002
	if version == 1 {
		edition = EditionPhase1
	} else if version == 2 && !componentVersion2Families()[family] {
		edition = EditionPhase2
	}
	return ComponentCodec{edition: edition, version: version}, nil
}

var (
	componentVersion2FamiliesOnce  sync.Once
	componentVersion2FamiliesValue map[string]bool
)

// componentVersion2Families returns the families whose highest version in
// go-asn1's TS 29.002 application context catalogue is 2.
func componentVersion2Families() map[string]bool {
	componentVersion2FamiliesOnce.Do(func() { componentVersion2FamiliesValue = componentVersion2FamiliesFrom(gsm_map.ApplicationContexts()) })
	return componentVersion2FamiliesValue
}

func componentVersion2FamiliesFrom(contexts []gsm_map.ApplicationContext) map[string]bool {
	highest := make(map[string]int)
	for _, ac := range contexts {
		arcs := strings.Split(ac.OID.String(), ".")
		if ac.Version > highest[arcs[6]] {
			highest[arcs[6]] = ac.Version
		}
	}
	families := make(map[string]bool)
	for family, version := range highest {
		if version == 2 {
			families[family] = true
		}
	}
	return families
}

// Edition returns the selected source edition, or zero for an unset codec.
func (c ComponentCodec) Edition() Edition { return c.edition }

// Version returns the application context version, or zero for an unset codec.
func (c ComponentCodec) Version() int { return c.version }

// ParseArgument decodes an operation's ARGUMENT. Empty data means no parameter.
func (c ComponentCodec) ParseArgument(opCode int64, data []byte) (any, error) {
	return c.parseComponent(opCode, 0, componentArgument, data)
}

// ParseResult decodes an operation's RESULT. An absent RESULT clause is refused.
func (c ComponentCodec) ParseResult(opCode int64, data []byte) (any, error) {
	return c.parseComponent(opCode, 0, componentResult, data)
}

// ParseError decodes a listed error's PARAMETER, using this edition's local codes.
func (c ComponentCodec) ParseError(opCode, errorCode int64, data []byte) (any, error) {
	return c.parseComponent(opCode, errorCode, componentError, data)
}

// MarshalArgument encodes exactly the generated value type returned by ParseArgument.
// A nil value means no parameter; a typed nil is not an absent parameter.
func (c ComponentCodec) MarshalArgument(opCode int64, value any) ([]byte, error) {
	return c.marshalComponent(opCode, 0, componentArgument, value)
}

// MarshalResult encodes exactly the generated value type returned by ParseResult.
func (c ComponentCodec) MarshalResult(opCode int64, value any) ([]byte, error) {
	return c.marshalComponent(opCode, 0, componentResult, value)
}

// MarshalError encodes a listed error's parameter using this edition's local codes.
func (c ComponentCodec) MarshalError(opCode, errorCode int64, value any) ([]byte, error) {
	return c.marshalComponent(opCode, errorCode, componentError, value)
}

type componentRole uint8

const (
	componentArgument componentRole = iota
	componentResult
	componentError
)

type componentType struct {
	name      string
	valueType reflect.Type
	decode    func([]byte, ...ber.DecodeOption) (any, error)
	encode    func(any, ...ber.EncodeOption) ([]byte, error)
}
type componentOperation struct {
	name             string
	argument, result *componentType
	resultPresent    bool
	errors           map[int64]bool
}
type componentErrorInfo struct {
	name      string
	parameter *componentType
}
type componentCatalogue struct {
	operations map[int64]componentOperation
	errors     map[int64]componentErrorInfo
}

func (c ComponentCodec) componentParameter(opCode, errorCode int64, role componentRole) (*componentType, error) {
	if c.edition < EditionPhase1 || c.edition > EditionTS29002 || c.version == 0 {
		return nil, ErrComponentCodecUnset
	}
	catalogue := componentCatalogues()[c.edition]
	op, ok := catalogue.operations[opCode]
	if !ok {
		return nil, fmt.Errorf("%w: %s code %d", ErrComponentUnknownOperation, c.edition, opCode)
	}
	if err := c.checkComponentOperation(opCode); err != nil {
		return nil, err
	}
	switch role {
	case componentArgument:
		return op.argument, nil
	case componentResult:
		if !op.resultPresent {
			return nil, fmt.Errorf("%w: %s %s", ErrComponentNoResult, c.edition, op.name)
		}
		return op.result, nil
	default:
		e, ok := catalogue.errors[errorCode]
		if !ok {
			return nil, fmt.Errorf("%w: %s code %d", ErrComponentUnknownError, c.edition, errorCode)
		}
		if !op.errors[errorCode] {
			return nil, fmt.Errorf("%w: %s %s / %s", ErrComponentErrorNotListed, c.edition, op.name, e.name)
		}
		return e.parameter, nil
	}
}

func (c ComponentCodec) parseComponent(opCode, errorCode int64, role componentRole, data []byte) (any, error) {
	typ, err := c.componentParameter(opCode, errorCode, role)
	if err != nil {
		return nil, err
	}
	var value any
	if len(data) != 0 {
		if typ == nil {
			return nil, fmt.Errorf("%w: %s operation %d", ErrComponentUnexpectedParameter, c.edition, opCode)
		}
		value, err = typ.decode(data)
		if err != nil {
			return nil, fmt.Errorf("decoding %s: %w", typ.name, err)
		}
	}
	if err := c.checkComponentRules(opCode, errorCode, role, value); err != nil {
		return nil, err
	}
	return value, nil
}

func (c ComponentCodec) marshalComponent(opCode, errorCode int64, role componentRole, value any) ([]byte, error) {
	typ, err := c.componentParameter(opCode, errorCode, role)
	if err != nil {
		return nil, err
	}
	if value != nil {
		if typ == nil {
			return nil, fmt.Errorf("%w: %s operation %d", ErrComponentUnexpectedParameter, c.edition, opCode)
		}
		if reflect.TypeOf(value) != typ.valueType {
			return nil, fmt.Errorf("%w: %s expects %v, got %T", ErrComponentValueType, typ.name, typ.valueType, value)
		}
	}
	if err := c.checkComponentRules(opCode, errorCode, role, value); err != nil {
		return nil, err
	}
	if value == nil {
		return nil, nil
	}
	data, err := typ.encode(value, ber.WithoutBERReplay())
	if err != nil {
		return nil, fmt.Errorf("encoding %s: %w", typ.name, err)
	}
	return data, nil
}
