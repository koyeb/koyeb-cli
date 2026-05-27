FROM rust:1-alpine AS build
RUN apk --no-cache add musl-dev ca-certificates
WORKDIR /src
COPY . .
RUN cargo build --release --bin koyeb

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /src/target/release/koyeb /koyeb
ENTRYPOINT ["/koyeb"]
