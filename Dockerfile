FROM golang:1.17

RUN mkdir /app
ADD ./api /app
WORKDIR /app
#add a directory that other directories can be mounted to
RUN mkdir /data-mount
ADD ./secrets /app/secrets
COPY ./go.mod /app/go.mod
COPY ./go.sum /app/go.sum

RUN go build -o main .
CMD ["/app/main"]