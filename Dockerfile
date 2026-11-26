FROM golang:1.22-bookworm AS build
WORKDIR /src
RUN apt-get update && apt-get install -y --no-install-recommends \
    tesseract-ocr tesseract-ocr-eng poppler-utils \
    && rm -rf /var/lib/apt/lists/*
COPY go.mod ./
COPY . .
ENV DIGITIZER_ALLOW_STUB=1
RUN go test ./... && CGO_ENABLED=0 go build -o /out/digitizer ./cmd/digitizer

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends \
    tesseract-ocr tesseract-ocr-eng poppler-utils ca-certificates \
    && rm -rf /var/lib/apt/lists/*
COPY --from=build /out/digitizer /usr/local/bin/digitizer
WORKDIR /data
ENV DIGITIZER_LLAMA_URL=http://llama:8080
ENV DIGITIZER_MODEL=qwen
ENTRYPOINT ["digitizer"]
CMD ["help"]
