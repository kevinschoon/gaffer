FROM quay.io/vektorcloud/go:1.9 AS source

COPY . /go/src/github.com/mesanine/gaffer

RUN cd /go/src/github.com/mesanine/gaffer \
  && make test \
  && make build

FROM quay.io/vektorcloud/base:3.6

COPY --from=source /go/src/github.com/mesanine/gaffer/bin/gaffer /bin/gaffer

ENTRYPOINT ["/bin/gaffer"]
