FROM alpine:3.4

ADD run-backup.sh /

VOLUME /upload

RUN apk add --update bash python py-pip \
    && pip install awscli \
    && chmod +x /run-backup.sh

CMD ["/run-backup.sh"]