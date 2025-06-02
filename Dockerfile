FROM golang:1.22 as builder
WORKDIR /build
COPY . .
RUN cd cmd/ntw-web && go build -o /build/ntw-web .

FROM gcr.io/distroless/base-debian12
COPY --from=builder /build/ntw-web /ntw-web
EXPOSE 8080
ENTRYPOINT ["/ntw-web"]
