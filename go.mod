module github.com/gomaja/go-asn1-gsmmap

go 1.25.4

require (
	github.com/gomaja/go-asn1 v0.0.0-20260929180321-5f9ce3a46527
	github.com/gomaja/go-sms v0.0.0-20260928185505-66af3bfd4fa8
	github.com/google/go-cmp v0.7.0
)

retract (
	v1.0.3 // Published only to retract v1.0.0 to v1.0.2; depend on the main branch.
	v1.0.2 // Predates the main-only API; depend on the main branch.
	v1.0.1 // Predates the main-only API; depend on the main branch.
	v1.0.0 // Predates the main-only API; depend on the main branch.
)
