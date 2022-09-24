# BUILD IMAGE --------------------------------------------------------
ARG GO_VERSION=unknown
FROM golang:${GO_VERSION}-alpine as builder

# Get build tools and required header files
RUN apk add --no-cache build-base curl

WORKDIR /app
COPY . .

# Build the final node binary
ARG GIT_COMMIT=unknown
RUN go build -ldflags="-X 'main.Commit=$GIT_COMMIT'" -o bin/xmtpd cmd/xmtpd/main.go

# Initialize tendermint config
WORKDIR /tmp
RUN curl -L https://github.com/tendermint/tendermint/releases/download/v0.34.21/tendermint_0.34.21_linux_amd64.tar.gz -o tendermint.tar.gz && \
    tar xvzf tendermint.tar.gz && \
    mv tendermint /usr/bin/
RUN tendermint init

# ACTUAL IMAGE -------------------------------------------------------

FROM alpine:3.12

ARG GIT_COMMIT=unknown

LABEL maintainer="engineering@xmtp.com"
LABEL source="https://github.com/xmtp/xmtp-node-go"
LABEL description="XMTP Node Software"
LABEL commit=$GIT_COMMIT

# color, nocolor, json
ENV GOLOG_LOG_FMT=nocolor

# go-waku default port
EXPOSE 9000

COPY --from=builder /app/bin/xmtpd /usr/bin/

# tendermint config
WORKDIR /opt
COPY --from=builder /root/.tendermint /opt/.tendermint

ENTRYPOINT ["/usr/bin/xmtpd"]
# By default just show help if called without arguments
CMD ["--help"]
