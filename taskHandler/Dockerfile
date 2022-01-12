FROM golang:1.17 as build-stage

WORKDIR /home/data-dj

#add a directory that other directories can be mounted to
RUN mkdir /home/data-dj/data-mount

ADD . ./app
ADD ./secrets /home/data-dj/secrets
# RUN chmod -R 744 /home/data-dj/secrets 
COPY ./go.mod ./app/go.mod
COPY ./go.sum ./app/go.sum
WORKDIR /home/data-dj/app

RUN go build -o main .

FROM build-stage

RUN useradd --create-home --shell /bin/bash data-dj
USER data-dj

# CMD ["/bin/bash"]
CMD ["/home/data-dj/app/main"]