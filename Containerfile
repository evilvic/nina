FROM docker.io/golang:1.22-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o nina .

FROM docker.io/alpine:latest
RUN apk add --no-cache ffmpeg
COPY --from=builder /app/nina /usr/local/bin/nina
EXPOSE 4533
CMD ["nina"]
