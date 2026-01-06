FROM node:22 AS tailwind-builder
WORKDIR /app
COPY ./package.json ./package-lock.json ./
RUN ["sh","-lc","npm ci"]

COPY ./templates ./templates
COPY ./tailwind/styles.css ./tailwind/styles.css
RUN ["npm", "run", "build:css"]

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
COPY --from=tailwind-builder /app/assets/styles.css ./assets/styles.css
CMD ./server