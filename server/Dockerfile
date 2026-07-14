FROM scratch
COPY statsig-proxy-compressed /app/statsig-proxy
COPY server.crt /app/server.crt
COPY server.key /app/server.key
EXPOSE 443
WORKDIR /app
CMD ["/app/statsig-proxy"]
