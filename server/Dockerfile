FROM alpine:3.19
COPY statsig-proxy /app/statsig-proxy
COPY server.crt /app/server.crt
COPY server.key /app/server.key
COPY ca.crt /app/ca.crt
COPY statsig_response.json /app/statsig_response.json
EXPOSE 80 443
WORKDIR /app
CMD ["/app/statsig-proxy"]
