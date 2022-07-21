# BUILD IMAGE --------------------------------------------------------
FROM golang:1.17-alpine as builder

ARG GIT_COMMIT=unknown

# Get build tools and required header files
RUN apk add --no-cache build-base

WORKDIR /app
COPY . .

# Build the final node binary
RUN go build -ldflags="-X 'main.Commit=$GIT_COMMIT'" -o bin/xmtpd cmd/xmtpd/main.go

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

ENTRYPOINT ["/usr/bin/xmtpd"]
# By default just show help if called without arguments
CMD ["--help"]
