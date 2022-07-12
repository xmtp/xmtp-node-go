FROM golang:1.17-alpine as builder
RUN apk add --no-cache build-base
WORKDIR /code
COPY . .
RUN go test ./pkg/e2e/... -c -o build/xmtpd-e2e

FROM golang:1.17-alpine
ARG GIT_COMMIT=unknown
LABEL commit=$GIT_COMMIT
COPY --from=builder /code/build/xmtpd-e2e /usr/bin/
COPY --from=builder /code/dev/e2e/entrypoint.sh /opt/
ENV E2E_CONTINUOUS=true
CMD [ "go", "tool", "test2json", "-t", "xmtpd-e2e" ]
