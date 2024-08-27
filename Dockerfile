FROM alpine:latest

USER 30011
COPY ./developer-proxy /developer-proxy

ENTRYPOINT [ "/developer-proxy" ]