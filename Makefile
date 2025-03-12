RABBIT_PORT=5672
RABBIT_URL=RABBIT_URL=localhost:$(RABBIT_PORT)
RABBIT_USERNAME=RABBIT_USERNAME=guest
RABBIT_PASSWORD=RABBIT_PASSWORD=guest
ENV=$(RABBIT_URL) $(RABBIT_PASSWORD) $(RABBIT_USERNAME)

orchestrator:
	cd src && $(ENV) go run ./send/send.go

worker:
	cd src && $(ENV) go run ./receive/receive.go

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
