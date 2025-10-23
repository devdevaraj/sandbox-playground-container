FROM ubuntu:22.04

# Install required packages
RUN apt-get update && apt-get install -y \
    qemu-kvm \
    libvirt-daemon-system \
    libvirt-clients \
    curl \
    wget \
    tar \
    jq \
    iptables \
    dnsmasq \
    bridge-utils \
    iproute2 \
    procps \
    ca-certificates \
    git \
    build-essential \
    python3 \
    python3-pip \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

RUN wget https://github.com/firecracker-microvm/firecracker/releases/download/v1.11.0/firecracker-v1.11.0-x86_64.tgz
RUN tar -xf  firecracker-v1.11.0-x86_64.tgz

RUN mv release-v1.11.0-x86_64/firecracker-v1.11.0-x86_64 /usr/local/bin/firecracker \
    && chmod +x /usr/local/bin/firecracker

RUN mv release-v1.11.0-x86_64/jailer-v1.11.0-x86_64 /usr/local/bin/jailer \
    && chmod +x /usr/local/bin/jailer

RUN mkdir -p /dev/netdev

RUN chmod 666 /dev/kvm || true

RUN mkdir -p /root/firecracker/keys
COPY ./keys/id_rsa /root/firecracker/keys/ubuntu-24.04.id_rsa

COPY ./firestarter/firestarter /root/firecracker/firestarter
# COPY ./firestarter/firestarter /root/firecracker/firestarter
# COPY ./wss/wss /root/firecracker/ws-server

# COPY ./runner/runner.sh /root/firecracker/runner.sh
# RUN chmod +x /root/firecracker/runner.sh

WORKDIR /root/firecracker

ENTRYPOINT ["/root/firecracker/firestarter"]
# ENTRYPOINT ["/root/firecracker/runner.sh"]
CMD ["ubuntu2404n1"]