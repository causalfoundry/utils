teardown-docker:
	@docker container prune -f
	@docker volume prune -f


network-up:
	@docker network inspect local-net-util >/dev/null 2>&1 || \
    docker network create --driver bridge local-net-util


storage-down:
	@docker kill psg-util && docker rm psg-util || true
	@docker kill clh-util && docker rm clh-util || true

storage-up: storage-down teardown-docker network-up 
	@(docker run -d -p 9009:9000 --network local-net-util -e CLICKHOUSE_USER=user -e CLICKHOUSE_PASSWORD=pwd -e CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT=1 --name clh-util --ulimit nofile=262144:262144 clickhouse/clickhouse-server:25.3)
	@(docker run -d -p 5439:5432 --network local-net-util -e POSTGRES_PASSWORD=pwd -e POSTGRES_USER=user -e POSTGRES_DB=postgres --name psg-util postgres:15 -c max_connections=500)

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
