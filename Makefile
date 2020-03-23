export GOBIN:=$(PWD)/bin

$(GOBIN)/golint:
	go install golang.org/x/lint/golint

.PHONY: test

test: $(GOBIN)/golint
	go test ./pkg/...
	bin/golint ./pkg/...
