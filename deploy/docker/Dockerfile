FROM alpine

RUN apk add -U tzdata
RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai  /etc/localtime

COPY heapster eventer /
COPY ca-certificates.crt /etc/ssl/certs/
RUN chmod +x /heapster
ENTRYPOINT ["/heapster"]
