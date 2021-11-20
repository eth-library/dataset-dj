# DataDJ

The DataDJ can be accessed at https://data-dj-2021.oa.r.appspot.com/


## Quickstart Guide


### 1. List available files (GET)
https://data-dj-2021.oa.r.appspot.com/files

You can call the endpoint within your browser or with `curl`.

Example:

```bash
curl https://data-dj-2021.oa.r.appspot.com/files 
```


### 2. Request an archive from the list of files

Edit the curl request below to include your `email` and the list of `files` that you want to download (note the included filepath).The endpoint can be called using `curl`. After a few minutes, you should receive an email with the download link.

Example:
```bash
curl https://data-dj-2021.oa.r.appspot.com/archive \
--include \
--header "Content-Type: application/json" \
--request "POST" \
--data '{"email":"your.name@librarylab.ethz.ch",
         "files":["data-mirror/cmt-001_1917_001_0015.jpg",
                   "data-mirror/cmt-001_1917_001_0019.jpg",
                   "data-mirror/cmt-001_1917_001_0057.jpg"]
        }'
```
---
## API Endpoints

### Listing all available files (GET)
https://data-dj-2021.oa.r.appspot.com/files

You can call the endpoint within your browser or with `curl`.

Example:

```bash
curl https://data-dj-2021.oa.r.appspot.com/files 
```

---

### Creating, modifying or downloading archives (POST)
https://data-dj-2021.oa.r.appspot.com/archive

This endpoint expects a request that contains three fields:

```json
{
  "email":"",
  "archiveID":"",
  "files":[]
}
```
`email` is a string, `archiveID` as well, being a truncated UUID as string and `files` is a list of strings containing the names of the files.
Depending on which fields are left empty, the API triggers different operations.


#### 1. Create an archive from a list of files

Both `email` and `archiveID` are left empty, whereas `files` contains the names of the files the archive should be initialised with. The endpoint can be called using `curl`.

Example:
```bash
curl https://data-dj-2021.oa.r.appspot.com/archive \
--include \
--header "Content-Type: application/json" \
--request "POST" \
--data '{"email":"your.name@librarylab.ethz.ch",
         "archiveID":"",
         "files":["data-mirror/cmt-001_1917_001_0015.jpg",
                   "data-mirror/cmt-001_1917_001_0019.jpg",
                   "data-mirror/cmt-001_1917_001_0057.jpg"]
        }'
```

#### 2. Add a list of files to an archive

`email` is left empty. `archiveID` contains the identifier of a previously created archive and `files` the list of files you want to add to the archive.

Example:
```bash
curl https://data-dj-2021.oa.r.appspot.com/archive \
--include \
--header "Content-Type: application/json" \
--request "POST" \
--data '{"email":"",
         "archiveID":"9d0b43d5",
         "files":["data-mirror/cmt-001_1917_001_0016.jpg",
                   "data-mirror/cmt-001_1917_001_0017.jpg",
                   "data-mirror/cmt-001_1917_001_0059.jpg"]
        }'
```

#### 3. Download an archive

`email` contains the email address the download link is being sent to, `archiveID` specifies the archive you want to download and `files` is left empty. The DataDj will send you a download link that allows you to download the archive as a .zip file.

Example:
```bash
curl https://data-dj-2021.oa.r.appspot.com/archive \
--include \
--header "Content-Type: application/json" \
--request "POST" \
--data '{"email":"your.name@librarylab.ethz.ch",
         "archiveID":"9d0b43d5",
         "files":[]
        }'
```

#### 4. Directly download a list of files as archive

`email` contains the email address the download link is being sent to, `archiveID` is left empty and `files` contains the names of the files you want to download.
The DJ creates an archive of the files in the request and will also return its identifier in the response, in case that archive needs to be accessed or modified later on. However it is not necessary to separatly trigger the notification containing the download link as this is going to happen automatically.

Example:
```bash
curl https://data-dj-2021.oa.r.appspot.com/archive \
--include \
--header "Content-Type: application/json" \
--request "POST" \
--data '{"email":"your.name@librarylab.ethz.ch",
         "archiveID":"",
         "files":["data-mirror/cmt-001_1917_001_0016.jpg",
                   "data-mirror/cmt-001_1917_001_0017.jpg",
                   "data-mirror/cmt-001_1917_001_0059.jpg"]
        }'
```

---

### Inspecting an archive (GET)

https://data-dj-2021.oa.r.appspot.com/archive/id

This endpoint allows to inspect the contents of an archive `id` either in the browser or via `curl`. The response is a JSON object that specifies the identifier and contents of the corresponding archive.

Example:
```bash
curl https://data-dj-2021.oa.r.appspot.com/archive/9d0b43d5
```



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
