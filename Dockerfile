 FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git build-essential libprotobuf-dev protobuf-compiler \
    libnl-route-3-dev pkg-config flex bison \
    python3 gcc g++ default-jdk nodejs iverilog \
    bash wget curl ruby lua5.4 \
    rustc ocaml ocaml-findlib unzip \
    && rm -rf /var/lib/apt/lists/*

RUN chmod 600 /etc/passwd /etc/shadow /etc/group

# Install Go 1.22 manually
RUN wget -q https://go.dev/dl/go1.22.3.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz && \
    rm go1.22.3.linux-amd64.tar.gz && \
    ln -sf /usr/local/go/bin/go /usr/bin/go
ENV PATH=$PATH:/usr/local/go/bin

# Mirror system Rust to the config path
RUN mkdir -p /home/godfather/.cargo/bin && \
    ln -sf /usr/bin/rustc /home/godfather/.cargo/bin/rustc

RUN wget -q https://github.com/JetBrains/kotlin/releases/download/v2.0.0/kotlin-compiler-2.0.0.zip && \
    unzip -q kotlin-compiler-2.0.0.zip -d /usr/local && \
    rm kotlin-compiler-2.0.0.zip
ENV PATH=$PATH:/usr/local/kotlinc/bin

ENV PATH=$PATH:/usr/local/go/bin

# Build nsjail from source at tag 3.4
RUN git clone --depth 1 --branch 3.4 \
    https://github.com/google/nsjail /opt/nsjail && \
    cd /opt/nsjail && make && \
    cp nsjail /usr/sbin/nsjail && \
    chmod +x /usr/sbin/nsjail

WORKDIR /app
COPY . .

RUN go build -o goboxd ./cmd/goboxd

EXPOSE 8080
CMD ["./goboxd"]