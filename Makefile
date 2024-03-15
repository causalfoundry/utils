check: 
	@golangci-lint run ./...
	@go test ./...

release: check
ifndef TAG
	$(error TAG is not defined. Usage: make tag-and-push TAG=<tag-name>)
endif
	git push origin main && git tag $(TAG) && git push origin $(TAG)
