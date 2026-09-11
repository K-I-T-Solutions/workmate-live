FROM golang:1.25-alpine AS builder

RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=1 go build -o workmate-live-portal ./cmd/workmate-live-portal

FROM alpine:3.21

RUN apk add --no-cache ca-certificates sqlite-libs

WORKDIR /app

COPY --from=builder /app/workmate-live-portal .

EXPOSE 8080

ENTRYPOINT ["./workmate-live-portal"]
CMD ["--config", "/app/config/portal.yaml"]
