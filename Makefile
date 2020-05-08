

.PHONY: test_integration
test_integration:
	INTEGRATION_TEST=YES go test -cover -v `go list ./...`