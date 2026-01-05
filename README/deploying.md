# Deploying

The goal is to have our app running from a **Docker container**. We'll need:

- A `Dockerfile`.
- A `docker-compose.prod.yml` file.
- A slightly customized `.env.prod` file (copy/paste `.env`).

In this last `.env.prod` file, we have to change the values of a couple of variables:

- `PSQL_HOST=db`
- `PSQL_PORT=5432`

To test the **production** setup, we have to run:

```sh
docker compose -f docker-compose.yml -f docker-compose.prod.yml up --build
```

edit`PSQL_HOST` value to

> [!NOTE]  
>  We have to use `db` instead of `localhost`, because that's the value we're using in `docker-compose.prod.yml` (`db` is the name of the Docker container where we'll be running our database, used in the Docker network).During **development** we'll want to `PSQL_HOST` to be `localhost`.