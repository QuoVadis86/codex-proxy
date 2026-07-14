FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY main.go ./
RUN go mod init statsig-proxy && CGO_ENABLED=0 go build -ldflags="-s -w" -o statsig-proxy

FROM scratch
COPY --from=builder /src/statsig-proxy /app/statsig-proxy
COPY server.crt /app/server.crt
COPY server.key /app/server.key
COPY ca.crt /app/ca.crt
EXPOSE 80 443
WORKDIR /app
CMD ["/app/statsig-proxy"]
