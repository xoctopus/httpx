PACKAGES=$(shell go list ./... | grep -E -v 'example|proto|testdata')

debug:
	@echo ${PACKAGES}

tidy:
	@go mod tidy

cover: tidy vet
	@go test -race -failfast -parallel 1 -gcflags="all=-N -l" ${PACKAGES} -covermode=atomic -coverprofile=cover.out

test: tidy vet
	@go test -race -failfast -parallel 1 -gcflags="all=-N -l" ${PACKAGES}

vet:
	@go vet ${PACKAGES}

report:
	@echo ">>>static checking"
	@go vet ./...
	@echo "done\n"
	@echo ">>>detecting ineffectual assignments"
	@ineffassign ./...
	@echo "done\n"
	@echo ">>>detecting icyclomatic complexities over 10 and average"
	@gocyclo -over 10 -avg -ignore '_test|vendor' . || true
	@echo "done\n"

check: tidy vet test
