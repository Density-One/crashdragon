FROM golang:1.20.3-bullseye as crashdragon_builder
ENV DEBIAN_FRONTEND=noninteractive
ENV GOPATH /go
ENV PATH $PATH:$GOPATH/bin
ENV GIN_MODE release
RUN apt-get update && apt-get -y install libcurl4-gnutls-dev rsync postgresql sassc autotools-dev autoconf libjsoncpp-dev make golang git wget libghc-zlib-dev
WORKDIR /crashdragon
RUN git clone https://code.videolan.org/videolan/CrashDragon.git
RUN git clone https://chromium.googlesource.com/chromium/tools/depot_tools.git
ENV PATH $PATH:/crashdragon/depot_tools
RUN cd CrashDragon/third_party && fetch breakpad && mv src/* breakpad/ && cd breakpad && ./configure && make -j$(nproc)
RUN cd CrashDragon && sed -i '64d' Makefile && go mod tidy && make && make prefix=/opt/crashdragon install

FROM golang:1.20.3-bullseye
RUN apt-get update && apt-get -y install postgresql
COPY --from=crashdragon_builder /opt/crashdragon /opt/crashdragon
WORKDIR /opt/crashdragon/bin
ENV GIN_MODE release
RUN /etc/init.d/postgresql start && su postgres -c 'createuser -w crashdragon' && su postgres -c 'createdb -w -O crashdragon crashdragon'
RUN mkdir -p /etc/crashdragon
RUN echo "local all all trust" > /etc/postgresql/13/main/pg_hba.conf
RUN echo "host all all all trust" >> /etc/postgresql/13/main/pg_hba.conf

RUN echo "[db]" > /etc/crashdragon/config.toml
RUN echo "  connection = \"host=localhost user=crashdragon dbname=crashdragon sslmode=disable\"" >> /etc/crashdragon/config.toml

RUN echo "[directory]" >> /etc/crashdragon/config.toml
RUN echo "  assets = \"/opt/crashdragon/share/crashdragon/assets\"" >> /etc/crashdragon/config.toml
RUN echo "  content = \"/opt/crashdragon/share/crashdragon/files\"" >> /etc/crashdragon/config.toml
RUN echo "  templates = \"/opt/crashdragon/share/crashdragon/templates\"" >> /etc/crashdragon/config.toml

RUN echo "[housekeeping]" >> /etc/crashdragon/config.toml
RUN echo "  reportretentiontime = \"2190h\"" >> /etc/crashdragon/config.toml

RUN echo "[symbolicator]" >> /etc/crashdragon/config.toml
RUN echo "  executable = \"./minidump_stackwalk\"" >> /etc/crashdragon/config.toml
RUN echo "  trimmodulenames = true" >> /etc/crashdragon/config.toml

RUN echo "[web]" >> /etc/crashdragon/config.toml
RUN echo "  bindaddress = \":8080\"" >> /etc/crashdragon/config.toml
RUN echo "  bindsocket = \"/var/run/crashdragon/crashdragon.sock\"" >> /etc/crashdragon/config.toml
RUN echo "  usesocket = false" >> /etc/crashdragon/config.toml

EXPOSE 8080
CMD /etc/init.d/postgresql start && sleep 15 && ./crashdragon
