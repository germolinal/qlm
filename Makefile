orchestrator:
	cd src && go run ./send/send.go

worker:
	cd src && go run ./receive/receive.go

rabbit:	
	docker run -it --rm  -p 5672:5672 -p 15672:15672 rabbitmq:4.1-rc-management-alpine

# Admin UI is available in localhost:15672 
# Auth: user = guest, password = guest

# Install colima by doing:
# brew install colima
docker_daemon:
	colima start

up:
	docker-compose build && docker-compose up
