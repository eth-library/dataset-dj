# dataset-dj
file aggregation and compression


# task queue / pub sub model with redis

download and run redis image with docker
```
docker pull redis
docker run --name redis-test-instance -p 6379:6379 -d redis
```
run the taskSubscriber to start listening to tasks.  
from the project root  
```
go run taskSubscriber/main.go
``` 

in an another terminal from the project root, publish task from the command line  
the command line arguments will be added as the list of files in the task. for example:   
```
go run taskPublisher/main.go img151.png img8.png
```

run the script to publish a new task