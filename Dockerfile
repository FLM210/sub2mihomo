FROM golang:1.25.5-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy 

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o sub2mihomo .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN apk --no-cache add tzdata

RUN addgroup -g 65532 nonroot &&\
    adduser -D -u 65532 -G nonroot nonroot

WORKDIR /app

COPY --from=builder /app/sub2mihomo .

EXPOSE 8080

COPY --from=builder /app/sub2mihomo /usr/local/bin/sub2mihomo

RUN chmod +x /usr/local/bin/sub2mihomo

USER nonroot

# 运行应用
CMD ["sub2mihomo"]