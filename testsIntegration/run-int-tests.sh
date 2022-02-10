#! /bin/bash

export ENV_FILE=.env.test
echo "using environment file: ${ENV_FILE}"

# run unit tests
echo "--- Start Unit Tests ---"
cd ../api
go test -v
cd ../taskHandler
go test -v
cd ../testsIntegration
echo "--- Finished Unit Tests ---"

# run integration tests
echo "--- Start Integration Tests ---"
docker-compose -f ../docker-compose.yml down
docker-compose -f ../docker-compose.yml up -d --build
source $(pwd)/../$ENV_FILE && export $(cut -d= -f1 $(pwd)/../$ENV_FILE)
# wait for services to start
echo "waiting 20s for backend services"
sleep 20
go test -v
echo "--- Finished Integration Tests ---"
docker-compose -f ../docker-compose.yml stop
echo "containers are stopped but not removed. To remove them run \n    docker-compose -f ../docker-compose.yml down \nor to inspect logs for a service e.g.: \n    docker logs dataset-dj_taskhandler_1"