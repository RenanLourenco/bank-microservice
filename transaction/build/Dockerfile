FROM golang:1.21.5-alpine as builder

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN CGO_ENABLED=0 go build -o transactionApp ./cmd/api

RUN chmod +x /app/transactionApp

FROM alpine:latest

RUN mkdir /app

COPY --from=builder /app/transactionApp /app

CMD [ "/app/transactionApp" ]