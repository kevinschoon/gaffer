FROM quay.io/vektorcloud/go:dep AS build

RUN apk add --no-cache protobuf protobuf-dev

RUN mkdir -p /go/src/github.com/golang \
  && cd /go/src/github.com/golang \
  && git clone https://github.com/golang/protobuf.git \
  && cd protobuf \
  && make \
  && go get -u github.com/jteeuwen/go-bindata/...

COPY . /go/src/github.com/mesanine/gaffer

RUN cd /go/src/github.com/mesanine/gaffer \
  && make 

FROM quay.io/vektorcloud/base:3.6 

COPY --from=build /go/src/github.com/mesanine/gaffer/bin/gaffer /usr/bin/

ENTRYPOINT ["/usr/bin/gaffer"]


