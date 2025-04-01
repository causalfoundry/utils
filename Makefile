teardown-docker:
	@docker container prune -f
	@docker volume prune -f

storage-down:
	@docker kill psg && docker rm psg

storage-up: teardown-docker 
	@(docker container inspect clh > /dev/null && docker start clh) || (docker run -d -p 9000:9000 -e CLICKHOUSE_USER=user -e CLICKHOUSE_PASSWORD=pwd -e CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 --name clh --ulimit nofile=262144:262144 clickhouse/clickhouse-server:25.3)
	@(docker container inspect psg > /dev/null && docker start psg) || (docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=pwd -e POSTGRES_USER=user -e POSTGRES_DB=postgres --name psg postgres:15 -c max_connections=500)

check: storage-up
	@golangci-lint run ./...
	@go test ./...

clean-release: check
ifndef TAG
	$(error TAG is not defined. Usage: make tag-and-push TAG=<tag-name>)
endif
	git tag -d $(TAG) || git push --delete origin $(TAG) || (git push origin main && git tag $(TAG) && git push origin $(TAG))

release: check
ifndef TAG
	$(error TAG is not defined. Usage: make tag-and-push TAG=<tag-name>)
endif
	git push origin main && git tag $(TAG) && git push origin $(TAG)
