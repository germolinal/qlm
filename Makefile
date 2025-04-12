RABBIT_PORT=5672
RABBIT_URL=127.0.0.1
RABBIT_USERNAME=guest
RABBIT_PASSWORD=guest
ENV=RABBIT_URL=$(RABBIT_URL) RABBIT_PASSWORD=$(RABBIT_PASSWORD) RABBIT_USERNAME=$(RABBIT_USERNAME) RABBIT_PORT=$(RABBIT_PORT)

orchestrator: FORCE
	cd core && $(ENV) go run ./orchestrator/orchestrator.go

build_orchestrator:
	docker build --pull --rm -f Dockerfile.orchestrator -t qlm-orchestrator:latest . 

run_orchestrator:
	docker run --network host -p 8080:8080/tcp qlm-orchestrator:latest 


build_worker:
	docker build --pull --rm -f Dockerfile.worker -t qlm-worker:latest . 

worker: FORCE
	cd core && CONCURRENCY=1 $(ENV) go run ./worker/worker.go

run_worker:
	docker run --network host -p $(RABBIT_PORT):$(RABBIT_PORT)/tcp qlm-worker:latest


playground: FORCE
	cd playground &&  go run ./playground.go && cd ..

build_playground:
	docker build --pull --rm -f Dockerfile.playground -t qlm-playground:latest . 

run_playground:
	docker run --network host -p 3000:3000/tcp qlm-playground:latest




rabbit:	
# Admin UI is available in localhost:15672 
# Auth: user = guest, password = guest
	docker run -it --rm  -p $(RABBIT_PORT):$(RABBIT_PORT) -p 15672:15672 rabbitmq:4.1-rc-management-alpine


# Install colima by doing:
# brew install colima
docker_daemon:
	colima start

up:
	docker-compose build && docker-compose up


FORCE: ;
