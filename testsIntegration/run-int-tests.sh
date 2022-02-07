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
docker-compose -f ../docker-compose.deploy.yml down
docker-compose -f ../docker-compose.deploy.yml up --build -d
source $(pwd)/../$ENV_FILE && export $(cut -d= -f1 $(pwd)/../$ENV_FILE)
# wait for services to start
echo "waiting for backend services"
sleep 15
go test -v
echo "--- Finished Integration Tests ---"
docker-compose -f ../docker-compose.deploy.yml down