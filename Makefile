RABBIT_PORT=5672
RABBIT_URL=127.0.0.1
RABBIT_USERNAME=guest
RABBIT_PASSWORD=guest
ENV=RABBIT_URL=$(RABBIT_URL) RABBIT_PASSWORD=$(RABBIT_PASSWORD) RABBIT_USERNAME=$(RABBIT_USERNAME) RABBIT_PORT=$(RABBIT_PORT)

orchestrator:
	cd core && $(ENV) go run ./orchestrator/orchestrator.go

build_orchestrator:
	docker build --pull --rm -f Dockerfile.orchestrator -t qlm-orchestrator:latest . 

run_orchestrator:
	docker run --network host -p 8080:8080/tcp qlm-orchestrator:latest 

worker:
	cd core && CONCURRENCY=1 $(ENV) go run ./worker/worker.go

playground: FORCE
	cd playground && $(ENV) go run ./playground.go && cd ..

rabbit:	
# Admin UI is available in localhost:15672 
# Auth: user = guest, password = guest
	docker run -it --rm  -p $(RABBIT_PORT):$(RABBIT_PORT) -p 15672:15672 rabbitmq:4.1-rc-management-alpine


# Install colima by doing:
# brew install colima
docker_daemon:
	colima start


FORCE: ;
