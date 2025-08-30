run:
	docker-compose up --build --force-recreate

down:
	podman-compose down --remove-orphans
	podman rmi $(podman images -q --filter dangling=true) --force
