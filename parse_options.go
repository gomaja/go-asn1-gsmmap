package gsmmap

import "github.com/gomaja/go-asn1/telecom/ss7/gsm_map/runtime/ber"

type parseConfig struct {
	violations *[]Violation
	log        ber.ViolationLog
}

// WithTolerance makes ParseSriSm, ParseMoFsm and ParseMtFsm decode with
// go-asn1's ber.WithTolerance and, when Parse succeeds, append the admitted
// violations to *violations. It admits representable constraint and protocol-rule
// violations, including the BER forms of TS 29.002 V19.1.0 §17.1.1.
// Package conversion rules still apply. Marshal remains strict; present empty
// optional octet strings encode as absent when presence is determined by length.
// A nil destination or a repeated WithTolerance returns ErrParseOptionInvalid.
// Parse appends without synchronization, so calls running at the same time
// need separate slices.
func WithTolerance(violations *[]Violation) ParseOption {
	return func(config *parseConfig) error {
		if violations == nil || config.violations != nil {
			return ErrParseOptionInvalid
		}
		config.violations = violations
		return nil
	}
}

func makeParseConfig(opts []ParseOption) (*parseConfig, error) {
	config := &parseConfig{}
	for _, opt := range opts {
		if opt == nil {
			return nil, ErrParseOptionInvalid
		}
		if err := opt(config); err != nil {
			return nil, err
		}
	}
	return config, nil
}

func (config *parseConfig) decodeOptions() []ber.DecodeOption {
	if config.violations == nil {
		return nil
	}
	return []ber.DecodeOption{ber.WithTolerance(&config.log)}
}

func (config *parseConfig) publish() {
	if config.violations != nil {
		*config.violations = append(*config.violations, config.log.Snapshot()...)
	}
}
