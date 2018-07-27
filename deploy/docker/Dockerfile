FROM scratch

COPY heapster eventer /
COPY ca-certificates.crt /etc/ssl/certs/
COPY zoneinfo.zip /usr/local/go/lib/time/

#   nobody:nobody
USER 65534:65534
ENTRYPOINT ["/heapster"]
