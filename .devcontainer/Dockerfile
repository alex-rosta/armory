FROM golang:1.23.4

# Install additional tools
RUN apt-get update && apt-get install -y \
    curl \
    git \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /workspace

EXPOSE 3000