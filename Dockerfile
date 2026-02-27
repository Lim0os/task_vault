FROM golang:1.24-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -o /task_vault ./cmd/task_vault

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /task_vault /task_vault
COPY migrations /migrations

EXPOSE 8080

CMD ["/task_vault"]
