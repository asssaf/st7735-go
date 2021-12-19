FROM golang:1.16.3-alpine3.13 AS builder

ENV APP_ROOT=/root
WORKDIR /$APP_ROOT

ENV GOARM=6
ENV GOARCH=arm
ENV GOOS=linux
ENV GOPROXY=proxy.golang.org

COPY go.mod $APP_ROOT/
RUN go mod download

COPY . $APP_ROOT/
RUN go build -o /st7735 ./cli/


FROM --platform=linux/arm/v6 arm32v6/alpine:3.13.5

ENV APP_ROOT=/root
WORKDIR /$APP_ROOT

COPY --from=builder /st7735 /root/st7735

ENTRYPOINT [ "/root/st7735" ]
