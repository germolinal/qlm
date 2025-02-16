orchestrator:
	cargo run --bin orchestrator -- ./config.json

worker:
	cargo run --bin worker

rabbit:	
# user = guest, password = guest
	docker run -it --rm  -p 5672:5672 -p 15672:15672 rabbitmq:4.1-rc-management-alpine

docker_daemon:
# brew install colima
	colima start