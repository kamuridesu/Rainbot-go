FROM alpine:3.21 AS download-ffmpeg

ARG TARGETARCH

RUN apk add --no-cache curl tar xz

RUN case "$(uname -m)" in \
      x86_64) ARCH=amd64 ;; \
      aarch64) ARCH=arm64 ;; \
      *) echo "Unsupported architecture: $(uname -m)"; exit 1 ;; \
    esac && \
    echo "Detected system architecture: $(uname -m) -> Downloading FFmpeg for: $ARCH" && \
    curl -L "https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-${ARCH}-static.tar.xz" | \
    tar -xJ --no-same-owner --no-same-permissions && \
    mv ffmpeg-*-static/ffmpeg /usr/local/bin/ffmpeg && \
    mv ffmpeg-*-static/ffprobe /usr/local/bin/ffprobe && \
    rm -rf ffmpeg-*-static

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
RUN go build -ldflags='-s -w -extldflags "-static"'  -o "default-app"
RUN upx --best --lzma /workspace/default-app

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata libwebp-tools
WORKDIR /app

COPY ./migrations /app/migrations
COPY --from=download-ffmpeg /usr/local/bin/ffmpeg /usr/local/bin/ffmpeg
COPY --from=download-ffmpeg /usr/local/bin/ffprobe /usr/local/bin/ffprobe
COPY --from=build /workspace/default-app /usr/local/bin/default-app

ENTRYPOINT [ "default-app" ]

