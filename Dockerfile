FROM golang
WORKDIR /go/src/kc868-mqtt-transport

ADD go.mod go.sum /go/src/kc868-mqtt-transport/
RUN go mod download

ADD *.go /go/src/kc868-mqtt-transport/

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm go build -ldflags="-s -w" -a -installsuffix cgo -o /go/bin/kc868-mqtt-transport

FROM scratch

COPY --from=0 /go/bin/kc868-mqtt-transport /usr/bin/kc868-mqtt-transport

CMD ["/usr/bin/kc868-mqtt-transport"]
