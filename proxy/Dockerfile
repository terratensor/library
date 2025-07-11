FROM golang:1.24 as builder

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /proxy .

FROM alpine:latest
WORKDIR /
COPY --from=builder /proxy /proxy
CMD ["/proxy"]