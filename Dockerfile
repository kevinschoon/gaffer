FROM quay.io/vektorcloud/base:3.5

COPY bin/gaffer /bin/
COPY entrypoint.sh /

RUN mkdir /gaffer \
  && apk add --no-cache sqlite

WORKDIR /gaffer
VOLUME /gaffer

CMD ["gaffer"]
ENTRYPOINT ["/entrypoint.sh"]
