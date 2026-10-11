module github.com/gomaja/go-asn1-gsmmap

go 1.26.9

require (
	github.com/gomaja/go-asn1 v0.0.0-20261011084734-d77484b8fa99
	github.com/gomaja/go-sms v0.0.0-20261002213211-02959b2bed2e
	github.com/google/go-cmp v0.7.0
)

retract (
	v1.0.3 // Published only to retract v1.0.0 to v1.0.2; depend on the main branch.
	v1.0.2 // Predates the main-only API; depend on the main branch.
	v1.0.1 // Predates the main-only API; depend on the main branch.
	v1.0.0 // Predates the main-only API; depend on the main branch.
)
