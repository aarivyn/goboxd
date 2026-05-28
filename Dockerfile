FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies
RUN apt-get update && apt-get install -y \
    golang-go \
    git \
    build-essential \
    libprotobuf-dev \
    protobuf-compiler \
    libnl-route-3-dev \
    pkg-config \
    flex \
    bison \
    python3 \
    gcc \
    g++ \
    default-jdk \
    nodejs \
    iverilog \
    bash \
    && rm -rf /var/lib/apt/lists/*

# Build nsjail from source at tag 3.4
RUN git clone --branch 3.4 --depth 1 \
    https://github.com/google/nsjail /opt/nsjail && \
    cd /opt/nsjail && make && \
    cp nsjail /usr/sbin/nsjail && \
    chmod +x /usr/sbin/nsjail

WORKDIR /app
COPY . .

RUN go build -o goboxd ./cmd/goboxd

EXPOSE 8080
CMD ["./goboxd"]