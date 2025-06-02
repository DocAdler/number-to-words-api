FROM golang:1.22 as builder
WORKDIR /build
COPY . .

# Додаємо потрібний пакет CLI прямо тут, якщо не можеш зробити go get локально:
RUN cd cmd/ntw-web && go get gopkg.in/urfave/cli.v1 && go build -o /build/ntw-web .

FROM gcr.io/distroless/base-debian12
COPY --from=builder /build/ntw-web /ntw-web
EXPOSE 8080
ENTRYPOINT ["/ntw-web"]

