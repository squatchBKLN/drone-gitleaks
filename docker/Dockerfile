FROM alpine:latest as alpine
RUN apk add -U --no-cache ca-certificates

FROM alpine:latest
ENV GODEBUG netdns=go

COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

RUN wget https://github.com/squatchBKLN/gitleaks/releases/download/v9/gitleaks-linux-amd64
RUN mv gitleaks-linux-amd64 /bin/gitleaks
RUN chmod +x /bin/gitleaks

RUN apk add --no-cache bash git openssh-client
ADD release/linux/amd64/plugin /bin/
ENTRYPOINT ["/bin/plugin"]