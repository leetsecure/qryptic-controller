FROM golang:1.22-alpine as builder
WORKDIR /app
COPY . .
RUN go mod tidy
RUN go build -o /app/bin/qryptic-controller /app/cmd/controller/main.go

FROM golang:1.22-alpine
WORKDIR /app
COPY --from=builder /app/bin/qryptic-controller /app/qryptic-controller
RUN apk add wireguard-tools
CMD ["/app/qryptic-controller"]