FROM alpine:3.10

RUN apk add --no-cache ca-certificates

ADD ./dns-network-policy-operator /dns-network-policy-operator

ENTRYPOINT ["/dns-network-policy-operator"]
