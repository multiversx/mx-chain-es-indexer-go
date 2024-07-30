FROM golang:1.20.7 as builder

RUN apt-get update && apt-get install -y

WORKDIR /multiversx
COPY . .

WORKDIR /multiversx/cmd/elasticindexer

RUN go build -o elasticindexer

# ===== SECOND STAGE ======
FROM ubuntu:22.04
RUN apt-get update && apt-get install -y

COPY --from=builder /multiversx/cmd/elasticindexer /multiversx

EXPOSE 22112

WORKDIR /multiversx

ENTRYPOINT ["./elasticindexer"]
CMD ["--log-level", "*:DEBUG"]
