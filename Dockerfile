FROM golang:latest AS builder

WORKDIR /app
ADD main.go .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ./main main.go

FROM alpine:latest
COPY --from=builder /app/main .
CMD ["./main"]