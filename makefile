test:
	go clean -testcache
	go test -coverprofile testResults.out
	go tool cover -html=testResults.out