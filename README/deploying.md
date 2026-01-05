# Deploying

The goal is to have our app running from a **Docker container**. We'll need:

- A `Dockerfile`:

```dockerfile
FROM golang
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY ./ ./
RUN go build -v -o ./server ./cmd/server/
CMD ./server
```

- A slightly customized `.env.prod` file (copy/paste `.env`). In this last `.env.prod` file, we have to change the values of a couple of variables:

  - `PSQL_HOST=db`
  - `PSQL_PORT=5432`

> [!NOTE]  
>  We have to use `db` instead of `localhost`, because that's the value we're using in `docker-compose.prod.yml` (`db` is the name of the Docker container where we'll be running our database, used in the Docker network). During **development** we'll want to `PSQL_HOST` to be `localhost`.

- Finally, a `docker-compose.prod.yml` file:

```yml
services:
  server:
    build:
      context: ./
      dockerfile: Dockerfile
    restart: always
    volumes:
      # - <our-computer>:<container>
      - ./images:/app/images
      - ./.env.prod:/app/.env:ro
    # Testing (remove before deploying)
    ports:
      - 3000:3000
    depends_on:
      - db
```

> [!NOTE]
> Note how we're mounting our `.env.prod` as simply `.env` in our container, because that's the file that our `main` function loads.

To test the **production** setup, we have to run:

```sh
docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build
```

## Multi-stage Docker Builds

There's no need for a container with the `go` compiler, just to run our application. If we use a [multi-stage build](https://docs.docker.com/build/building/multi-stage/), we could use one container to build, then another one to run our binary:

```dockerfile
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
```

> [!NOTE]
> Note how we're copying over our `.env.prod` as simply `.env` in our container; thanks to this, we don't need to mount it in `docker-compose.prod.yml` file.

As a sanity test, it's a good idea to get rid of the volumes before spining up our app again:

```sh
docker compose -f docker-compose.yml -f docker-compose.prod.yml down -v
```

You can run the same command from the previous section to start **up** the app.