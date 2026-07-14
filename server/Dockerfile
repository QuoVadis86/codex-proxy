# Build stage
FROM golang:1.26-alpine AS builder
WORKDIR /src
COPY main.go go.mod ./
RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o statsig-proxy

# Runtime stage
FROM scratch
COPY --from=builder /src/statsig-proxy /app/statsig-proxy
COPY server.crt /app/server.crt
COPY server.key /app/server.key
COPY ca.crt /app/ca.crt
COPY statsig_response.json /app/statsig_response.json
EXPOSE 80 443
WORKDIR /app
CMD ["/app/statsig-proxy"]
