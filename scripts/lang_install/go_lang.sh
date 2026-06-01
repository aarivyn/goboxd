#!/bin/bash
wget https://go.dev/dl/go1.22.3.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.22.3.linux-amd64.tar.gz
rm go1.22.3.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
