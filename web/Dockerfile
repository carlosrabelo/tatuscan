FROM golang:1.24.2-bookworm AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /out/tatuscan-web ./web/cmd/tatuscan-web

FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates wget \
  && rm -rf /var/lib/apt/lists/* \
  && useradd -m -u 1000 appuser
COPY --from=build /out/tatuscan-web /usr/local/bin/tatuscan-web
USER appuser
EXPOSE 8050
ENV TATUSCAN_PORT=8050 TATUSCAN_API_URL=http://api:8040
CMD ["tatuscan-web"]
