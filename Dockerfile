FROM golang:1.25.5-alpine AS build
ARG TARGETARCH
ENV CGO_ENABLED=1
RUN apk add --no-cache gcc musl-dev upx 

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY ./*.go /workspace/
COPY ./internal /workspace/internal
COPY ./core /workspace/core
COPY ./commands/ /workspace/commands
RUN go build -ldflags='-s -w -extldflags "-static"'  -o "default-app"
RUN upx --best --lzma /workspace/default-app

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata libwebp-tools ffmpeg
WORKDIR /app

COPY ./migrations /app/migrations
COPY --from=build /workspace/default-app /usr/local/bin/default-app

ENTRYPOINT [ "default-app" ]
