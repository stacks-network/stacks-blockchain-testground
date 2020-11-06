FROM golang:buster AS builder
WORKDIR /build
ENV CGO_ENABLED 0
COPY . .
RUN cd plan/go && go build -a -o /testplan

FROM stacks-blockchain:testground-base
COPY --from=builder /testplan /testplan
COPY --from=builder /build/plan/scripts /scripts
ENV BLOCKSTACK_DEBUG 0
ENV BLOCKSTACK_LOG_JSON 1
EXPOSE 21443 21444 8080 18444 18443 28443 6060
ENTRYPOINT [ "/testplan"]
