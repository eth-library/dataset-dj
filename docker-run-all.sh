#! /bin/sh
echo "loading environment variables from file"
source .env.prod && export $(cut -d= -f1 .env.prod)

# make data folders if not already existing
mkdir -p ./data/{sources,archives}
# make container network
echo "creating data-dj network"
docker network create --driver=bridge data-dj
# redis
echo "starting Redis"
docker run --name data-dj_redis_1 \
    --net=data-dj \
    --network-alias=redis \
    --restart=unless-stopped \
    -d \
    --restart=unless-stopped \
    redis

# mongodb
echo "starting MongoDB"
docker run --name data-dj_db_1 \
    --net=data-dj \
    --network-alias=db \
    --restart=unless-stopped \
    -v $(pwd)/mongodb_data:/data/db \
    -d \
    --restart=unless-stopped \
    mongo

# wait for backend services to start
echo "waiting 15s for backend services to be ready"
sleep 15

# api
echo "starting API"
docker run --name data-dj_api_1 \
    --net=data-dj \
    --network-alias=api \
    -p 8765:8765 \
    --env-file=$(pwd)/.env.prod \
    -v $(pwd)/secrets:/app/secrets:ro \
    -v $(pwd)/data:/data-mount \
    -d \
    --restart=unless-stopped \
    registry.ethz.ch/bsunderland/data-dj-images/api:latest

# taskhandler
echo "starting taskhandler"
docker run --name data-dj_taskhandler_1 \
    --net=data-dj \
    --network-alias=taskhandler \
    --env-file=$(pwd)/.env.prod \
    -v $(pwd)/secrets:/app/secrets:ro \
    -v $(pwd)/data:/data-mount \
    -d \
    --restart=unless-stopped \
    registry.ethz.ch/bsunderland/data-dj-images/taskhandler:latest


echo "finished starting services"
 docker ps -f name=data-dj*