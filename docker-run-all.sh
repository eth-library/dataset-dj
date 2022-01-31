#! /bin/sh
mkdir -p ./data/{sources,archives}
# redis
echo "starting Redis"
docker run --name data-dj_redis_1 \
    --net=data-dj \
    --restart=unless-stopped \
    -d \
    --restart=unless-stopped \
    redis

# mongodb
echo "starting MongoDB"
docker run --name data-dj_db_1 \
    --net=data-dj \
    --restart=unless-stopped \
    -v $(pwd)/mongodb_data:/data/db \
    -d \
    --restart=unless-stopped \
    mongo

# wait for backend services to start
echo "waiting for backend services"
sleep 15

# api
echo "starting API"
docker run --name data-dj_api_1 \
    --net=data-dj \
    -p 8765:8765 \
    --env-file=$(pwd)/.env.prod \
    -v `pwd`/secrets:/app/secrets:ro \
    -v `pwd`/data:/data-mount \
    -d \
    --restart=unless-stopped \
    registry.ethz.ch/bsunderland/data-dj-images/api:latest

# taskhandler
docker run --name data-dj_taskhandler_1 \
    --net=data-dj \
    --env-file=$(pwd)/.env.prod \
    -v $(pwd)/secrets:/app/secrets:ro \
    -v $(pwd)/data:/data-mount \
    -d \
    --restart=unless-stopped \
    registry.ethz.ch/bsunderland/data-dj-images/taskhandler:latest