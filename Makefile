
.PHONY: build backend frontend docker-build

backend:
	cd backend; go run . 

frontend:
	cd frontend; npm run start

build:
	cd backend; go build

docker-build:
	docker build -f build/Dockerfile -t scm .
	{ \
		TAG=`docker images scm -q`; \
		docker tag $$TAG feeditout/scmwatcher; \
		docker push feeditout/scmwatcher; \
	}

docker-run:
	- docker rm -f scm
	- docker network create scm
	docker run -d --name scm --network scm -p 8080:8080 -p 1323:1323 scm

