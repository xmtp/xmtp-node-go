ARG GO_VERSION=unknown
FROM golang:${GO_VERSION}-alpine as builder
RUN apk add --no-cache build-base
WORKDIR /code
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY . .
ARG GIT_COMMIT=unknown
RUN go build -ldflags="-w -s -X 'main.GitCommit=$GIT_COMMIT'" -o build/xmtpd-e2e cmd/xmtpd-e2e/main.go

FROM golang:${GO_VERSION}-alpine
ARG GIT_COMMIT=unknown
LABEL commit=$GIT_COMMIT
COPY --from=builder /code/build/xmtpd-e2e /usr/bin/
CMD [ "xmtpd-e2e" ]
