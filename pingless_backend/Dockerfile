FROM golang:1.24.3-alpine
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache build-base libwebp-dev libwebp

COPY . .
RUN go build -o pingless .

EXPOSE 3000
CMD ["./pingless"]
