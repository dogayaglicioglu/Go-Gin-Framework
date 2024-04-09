FROM golang:1.21.4-alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    DB_HOST=my-postgres \
    DB_USER=dogayag1 \
    DB_NAME=assignmentkonzek \
    DB_PORT=5432

WORKDIR /app

COPY go.mod go.sum ./

RUN go mod tidy && go mod download

COPY . .

COPY .env .  

RUN go build -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/main .

COPY --from=builder /app/.env .  
EXPOSE 8000

CMD ["./main"]
