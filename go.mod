module github.com/gomaja/go-asn1-gsmmap

go 1.26.9

require (
	github.com/gomaja/go-3gpp-identity v0.0.0-20261015070659-866133e946dc
	github.com/gomaja/go-asn1 v0.0.0-20261021182037-fd8335c59a05
	github.com/gomaja/go-gad v0.0.0-20261015090110-463e9e2108fc
	github.com/gomaja/go-sms v0.0.0-20261016094937-0e8b7b4d2388
	github.com/gomaja/go-tbcd v0.0.0-20261015073026-64bd86dfe62e
	github.com/google/go-cmp v0.7.0
)

retract (
	v1.0.3 // Published only to retract v1.0.0 to v1.0.2; depend on the main branch.
	v1.0.2 // Predates the main-only API; depend on the main branch.
	v1.0.1 // Predates the main-only API; depend on the main branch.
	v1.0.0 // Predates the main-only API; depend on the main branch.
)
