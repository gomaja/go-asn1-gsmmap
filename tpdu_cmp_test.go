package gsmmap

import (
	"bytes"

	"github.com/gomaja/go-sms/encoding/tpdu"
	"github.com/google/go-cmp/cmp"
)

// equateTPDU compares tpdu.TPDU values by what a MAP codec must preserve:
// the TPDU octets carried in SM-RP-UI, together with the Direction and
// RPMessage that select how those octets are read. tpdu.TPDU keeps
// decoding state in unexported fields, which cmp cannot inspect.
var equateTPDU = cmp.Comparer(func(a, b tpdu.TPDU) bool {
	ab, errA := a.MarshalBinary()
	bb, errB := b.MarshalBinary()
	return errA == nil && errB == nil &&
		a.Direction == b.Direction && a.RPMessage == b.RPMessage &&
		bytes.Equal(ab, bb)
})
