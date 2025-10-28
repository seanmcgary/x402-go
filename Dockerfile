FROM golang:1.24-bookworm as builder

RUN apt-get update

RUN mkdir /app

WORKDIR /app
COPY . .

RUN make all

FROM debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    && rm -rf /var/lib/apt/lists/*

COPY --from=builder /app/bin/facilitator /usr/local/bin/facilitator

ENTRYPOINT ["/usr/local/bin/facilitator"]
CMD ["serve"]
