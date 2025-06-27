FROM golang:1.24.3-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache build-base libwebp-dev

COPY . .
RUN go build -o pingless .

FROM alpine:latest
WORKDIR /app
RUN apk add --no-cache libwebp
COPY --from=builder /app/pingless .
# COPY --from=builder /app/static ./static
COPY --from=builder /app/uploads ./uploads
EXPOSE 3000
CMD ["./pingless"]
