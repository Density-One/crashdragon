FROM golang:1.20.3-bullseye as crashdragon_builder
ENV DEBIAN_FRONTEND=noninteractive 
ENV GOPATH /go
ENV PATH $PATH:$GOPATH/bin
ENV GIN_MODE release

RUN apt-get update && apt-get -y install libcurl4-gnutls-dev rsync postgresql sassc autotools-dev autoconf libjsoncpp-dev make golang git wget libghc-zlib-dev

WORKDIR /crashdragon

RUN git clone https://github.com/Density-One/crashdragon.git
RUN git clone https://chromium.googlesource.com/chromium/tools/depot_tools.git

ENV PATH $PATH:/crashdragon/depot_tools

RUN cd crashdragon/third_party && fetch breakpad && mv src/* breakpad/ && cd breakpad && ./configure && make -j$(nproc)
RUN cd crashdragon && go mod tidy && make && make prefix=/opt/crashdragon install

FROM golang:1.20.3-bullseye

COPY --from=crashdragon_builder /opt/crashdragon /opt/crashdragon
COPY entrypoint.sh /entrypoint.sh

WORKDIR /opt/crashdragon/bin
ENV GIN_MODE release

RUN mkdir /etc/crashdragon
RUN chmod +x /entrypoint.sh

EXPOSE 8080
ENTRYPOINT ["/entrypoint.sh", "./crashdragon"]
