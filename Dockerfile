FROM golang AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN go build -v -o ./server ./cmd/server/

FROM ubuntu
WORKDIR /app
COPY ./assets ./assets
COPY .env.prod .env
COPY --from=builder /app/server ./server
CMD ./server