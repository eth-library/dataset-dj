#! /bin/bash

export ENV_FILE=.env.test
echo "using environment file: $ENV_FILE"
docker-compose -f ../docker-compose.deploy.yml down
docker-compose -f ../docker-compose.deploy.yml up --build -d
source ../$ENV_FILE && export $(cut -d= -f1 ../$ENV_FILE)
# wait for services to start
sleep 15
go test -v

docker-compose -f ../docker-compose.deploy.yml down