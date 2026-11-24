# syntax=docker/dockerfile:1
# ---------- build ----------
FROM golang:1.25-bookworm AS build
WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download || true
COPY . .
# go mod tidy gera/atualiza o go.sum com verificação no sum.golang.org.
RUN go mod tidy \
 && CGO_ENABLED=1 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

# ---------- runtime ----------
FROM debian:bookworm-slim
RUN apt-get update \
 && apt-get install -y --no-install-recommends ca-certificates tzdata \
 && rm -rf /var/lib/apt/lists/* \
 && useradd --system --uid 10001 --no-create-home app \
 && mkdir -p /data && chown app:app /data
COPY --from=build /out/server /usr/local/bin/server
USER app
ENV APP_ENV=production \
    PORT=8080 \
    DATABASE_PATH=/data/app.db \
    TZ=America/Sao_Paulo
VOLUME ["/data"]
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/usr/local/bin/server", "healthcheck"]
ENTRYPOINT ["/usr/local/bin/server"]
CMD ["serve"]
