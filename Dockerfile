# Build stage: compile Go binary
FROM golang:1.26-bookworm AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ cmd/
COPY internal/ internal/
COPY config.yaml ./
RUN go build -o ai-report ./cmd/ai-report

# Final stage: Python + Go binary + gh CLI
FROM python:3.12-slim-bookworm
WORKDIR /app

RUN apt-get update && \
    apt-get install -y --no-install-recommends git curl && \
    curl -fsSL https://cli.github.com/packages/githubcli-archive-keyring.gpg \
      | dd of=/usr/share/keyrings/githubcli-archive-keyring.gpg && \
    echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/githubcli-archive-keyring.gpg] https://cli.github.com/packages stable main" \
      > /etc/apt/sources.list.d/github-cli.list && \
    apt-get update && apt-get install -y --no-install-recommends gh && \
    rm -rf /var/lib/apt/lists/*

# Go runtime (needed by go test, not just the binary)
COPY --from=golang:1.26-bookworm /usr/local/go /usr/local/go
ENV PATH="/usr/local/go/bin:${PATH}"

RUN pip install --no-cache-dir notebooklm-py pyyaml

COPY --from=builder /build/ai-report ./ai-report
COPY scripts/ scripts/
COPY config.yaml podcast-config.yaml ./
COPY reports/ reports/
COPY go.mod go.sum ./
COPY cmd/ cmd/
COPY internal/ internal/

CMD ["bash"]
