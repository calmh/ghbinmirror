FROM alpine:latest
COPY ghbinmirror-linux-amd64 /bin/ghbinmirror
ENTRYPOINT ["/bin/ghbinmirror"]
