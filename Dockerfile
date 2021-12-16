FROM alpine:3.10.2
WORKDIR /app
RUN apk update \
        && apk upgrade \
        && apk add --no-cache \
        ca-certificates \
        && update-ca-certificates 2>/dev/null || true

COPY ./bin/smol ./smol
COPY ./swagger ./swagger
# fix: standard_init_linux.go:211: exec user process caused "no such file or directory"
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2

EXPOSE 8000
CMD ["./smol", "server"]
