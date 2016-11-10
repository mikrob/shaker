FROM golang:1.6
COPY . /go/src/shaker
WORKDIR /go/src/shaker
EXPOSE 8080
RUN go get -v && go build
CMD ["shaker"]
