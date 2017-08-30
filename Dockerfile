FROM quay.io/vektorcloud/base:3.6

ADD bin/gaffer /usr/bin/

ENTRYPOINT ["/usr/bin/gaffer"]
