FROM golang:1.17 AS builder

WORKDIR /opt

COPY . .

RUN go build

FROM gcr.io/distroless/base-debian11 AS runtime

COPY --from=builder /opt/turbo-doodle /usr/local/bin/turbo-doodle

ENTRYPOINT ["/usr/local/bin/turbo-doodle"]
