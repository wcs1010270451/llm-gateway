FROM golang:1.24.5-alpine AS build

WORKDIR /src/backend

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/llm-gateway ./cmd/server

FROM alpine:3.22

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=build /out/llm-gateway /app/llm-gateway

USER 65532:65532
EXPOSE 3212

ENTRYPOINT ["/app/llm-gateway"]
