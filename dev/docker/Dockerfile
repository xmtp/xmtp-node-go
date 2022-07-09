ARG GO_VERSION
FROM golang:${GO_VERSION}-alpine AS build
RUN apk add --no-cache git openssl ca-certificates moreutils
COPY go.mod /code/
COPY go.sum /code/
WORKDIR /code
RUN go mod download
COPY cmd cmd
COPY internal internal
COPY pkg pkg
ENV CGO_ENABLED=0 GOOS=linux
RUN go build -o /bin/xmtpd ./cmd/xmtpd
RUN go build -o /bin/xmtpdctl ./cmd/xmtpdctl

FROM alpine
COPY --from=build /bin/xmtp* /bin/
WORKDIR /opt
CMD ["/bin/xmtpd"]
