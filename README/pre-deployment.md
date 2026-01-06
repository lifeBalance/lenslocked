# Pre-Deployment

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

### Adding Tailwind to the Multi-Stage Build

Finally, let's add an optimized version of our `styles.css` to the **production image**:

```dockerfile
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
```

In the **first stage** we're:

- Copying over our dependencies, and installing them.
- Building a **minified** version of our styles.

> [!NOTE]
> Note how in the **last stage**, we're copying over the minified styles to the proper folder. You can verify this, if you check in your browser `http://localhost:3000/assets/styles.css`.

## Caddy

[Caddy](https://caddyserver.com/) is a web server + reverse proxy, that we'll use in front of your Go app. Our long term goal using it, it's so that we can:

- Run server on an **internal port** (`:3000`), and use caddy to expose it on public ports `:80` and `:443`.
- Let Caddy manage TLS certificates.

Let's start by creating a very simple [caddyfile](https://caddyserver.com/docs/caddyfile) at the root of our project:

```Caddyfile
:80

reverse_proxy server:3000
```

At this point we're using Caddy as a **reverse proxy**, meaning it will receive requests on port `80` and redirect them to port `3000`, which it's the port that the Docker container running our app exposes.

We'll be using  inside a Docker container, so we'll need to edit our `docker-compose.prod.yml` file:
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
    depends_on:
      - db
    # Testing (remove before deploying)
    # ports:
    #   - 3000:3000
    
  caddy:
    image: caddy
    restart: always
    ports:
      - 80:80
      - 443:443
    volumes:
      - ./Caddyfile:/etc/caddy/Caddyfile
```

> [!NOTE]
> Note how we're also removing `ports` from our `server`, since Caddy now will receive requests on `80/443` and route to the `3000` in our `server` (both services run in the Docker network).

To test this out, we'll bring **down** our running containers:

```sh
docker compose -f docker-compose.yml -f docker-compose.prod.yml down -v
```

And **up** again:

```sh
docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build
```

Now, you won't be able to access your app in `localhost:3000`, but if you point your browser to `localhost:80` or simply `localhost` voila!

> [!CAUTION]
> If you try to **signup** or submit any form, we'll probably get a CSRF error. The solution is simple: add `localhost` (next to `localhost:3000`) in the list of `TrustedOrigins` (in the `cmd/server/server.go` file).