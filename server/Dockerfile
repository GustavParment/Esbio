FROM golang:1.24-alpine AS builder

WORKDIR /build
COPY cmd/ ./cmd/
WORKDIR /build/cmd
RUN GOFLAGS=-mod=mod go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /server api/main.go

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
COPY --from=builder /server /server

EXPOSE 8080
CMD ["/server"]
