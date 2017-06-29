FROM quay.io/vektorcloud/base:3.6

COPY bin/gaffer /bin/

ENTRYPOINT ["/bin/gaffer"]
