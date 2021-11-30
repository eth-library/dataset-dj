FROM golang:1.17

RUN mkdir /app
ADD ./api /app
WORKDIR /app

ADD ./secrets /app/../secrets
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum
RUN go build -o main .

#add a directory that other directories can be mounted to
RUN mkdir /data-mount

CMD ["/app/main"]