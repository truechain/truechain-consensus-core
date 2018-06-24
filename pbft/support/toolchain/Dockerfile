FROM ubuntu:latest
ENV PATH /go/bin:/src/bin:/root/go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go:/src
RUN apt-get update
RUN apt-get install -y gcc git tar curl protobuf-compiler && \
    curl -sSLf https://storage.googleapis.com/golang/go1.10.linux-amd64.tar.gz | tar -C /usr/local -zx && \
    go get -v github.com/alecthomas/gometalinter && \
    go get -v golang.org/x/tools/cmd/... && \
    go get -v github.com/FiloSottile/gvt && \
    go get github.com/tools/godep && \
    go get -v github.com/golang/mock/mockgen && \
    go get github.com/derekparker/delve/cmd/dlv && \
    go get -u github.com/golang/protobuf/protoc-gen-go && \
    gometalinter --install && \
    chmod -R a+rw /go && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*
