FROM docker.io/golang:1.27-trixie AS builder
WORKDIR /goback
RUN apt update && apt -y dist-upgrade
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /bin/goback ./server

FROM scratch AS runtime
COPY --from=builder /bin/goback /bin/goback
EXPOSE 8000
ENTRYPOINT ["/bin/goback"]
