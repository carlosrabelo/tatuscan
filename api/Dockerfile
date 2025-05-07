FROM golang:1.24.2-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/tatuscan-api ./api/cmd/tatuscan-api

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata wget \
  && rm -rf /var/lib/apt/lists/* \
  && useradd -m -u 1000 appuser \
  && mkdir -p /data && chown appuser:appuser /data
COPY --from=build /out/tatuscan-api /usr/local/bin/tatuscan-api
USER appuser
EXPOSE 8040
ENV TATUSCAN_PORT=8040 TATUSCAN_DB_DIR=/data TATUSCAN_DB_FILE=tatuscan.db TIMEZONE=America/Cuiaba
CMD ["tatuscan-api"]
