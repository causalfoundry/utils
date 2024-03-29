teardown-docker:
	@docker container prune -f
	@docker volume prune -f

storage-down:
	@docker kill psg && docker rm psg
	@docker kill rds && docker rm rds

storage-up: teardown-docker 
	@(docker container inspect psg > /dev/null && docker start psg) || (docker run -d -p 5432:5432 -e POSTGRES_PASSWORD=pwd -e POSTGRES_USER=user -e POSTGRES_DB=postgres --name psg postgres:15 -c max_connections=500)
	@(docker container inspect rds > /dev/null && docker start rds) || (docker run -d -p 6379:6379 --name rds redis:7.0 redis-server --requirepass "pwd")

check: storage-up
	@golangci-lint run ./...
	@go test ./...

release: check
ifndef TAG
	$(error TAG is not defined. Usage: make tag-and-push TAG=<tag-name>)
endif
	git push origin main && git tag $(TAG) && git push origin $(TAG)
