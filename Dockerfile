# build
FROM golang:1.25-alpine AS builder

ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
ENV CGO_ENABLED=0

RUN apk add --no-cache git

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -trimpath -ldflags="-s -w" -o /usr/local/bin/api-empl-k8s-go ./


# runtime
FROM alpine:latest

LABEL org.opencontainers.image.source="https://github.com/frhan/api-empl-k8s-go"
LABEL org.opencontainers.image.description="Employee API Service"

RUN addgroup -S app && adduser -S -G app app \
    && apk add --no-cache ca-certificates \
    && update-ca-certificates

WORKDIR /home/app

COPY --from=builder /usr/local/bin/api-empl-k8s-go /usr/local/bin/api-empl-k8s-go
COPY --from=builder /app/migrations /migrations

RUN chown app:app /usr/local/bin/api-empl-k8s-go \
    && chmod +x /usr/local/bin/api-empl-k8s-go

USER app

EXPOSE 5000

CMD ["/usr/local/bin/api-empl-k8s-go"]
