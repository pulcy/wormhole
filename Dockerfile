# Load-balancer
FROM alpine:3.3

# ---------------------------------------------------------
# Installation
# ---------------------------------------------------------

# Install haproxy
RUN apk add -U haproxy

# ---------------------------------------------------------
# Configuration
# ---------------------------------------------------------

# Configure haproxy
RUN mkdir -p /data/logs
RUN mkdir -p /data/config
RUN mkdir -p /data/config
RUN mkdir -p /var/lib/haproxy/dev

# Added start process
ADD ./wormhole /app/

# Configure volumns
VOLUME ["/data"]
VOLUME ["/dev/log"]

# Exposed ports must be configured on the docker command line

# Start the proxy
ENTRYPOINT ["/app/wormhole"]
