FROM golang:1.23.9-alpine AS builder

RUN addgroup -S appuser \
    && adduser -S appuser -G appuser

RUN apk update && \
    apk add --no-cache git ca-certificates && \
    update-ca-certificates

WORKDIR /microservice

COPY . .

RUN go mod download

RUN GOARCH=amd64 CGO_ENABLED=0 GOOS=linux go build -o app ./cmd/api/*

FROM scratch as runner

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

USER appuser
COPY --from=builder /microservice/app .

EXPOSE 8080

CMD ["/app"]