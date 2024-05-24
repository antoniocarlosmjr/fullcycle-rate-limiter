# fullcycle-rate-limiter

### How to run

1) First of all, run the command for start the containers:

```
docker-compose up -d --build
```

The command above will start the Redis container and the API container. The API container will be available on port `8080` and the Redis container will be available on port `6379`.

2) After that you can test the API using the following command:

```
curl -X GET http://localhost:8080
```

or requests with token

```
curl -X GET http://localhost:8080 -
H "API_KEY: TOKEN"
```

You can see examples of use in `/api/requests`.

3) The variables for the rate limiter can be configured in the `.env` file in the root folder. They are:
   - `MAX_REQUESTS_WITHOUT_TOKEN_PER_SECOND`: The maximum number of requests per second without the `API_KEY` header.
   - `MAX_REQUESTS_WITH_TOKEN_PER_SECOND`: The maximum number of requests per second with the `API_KEY` header.
   - `TIME_BLOCK_IN_SECOND`: The time in seconds when the IP or token will be blocked.

* Obs.: After changing the `.env` file, you need to restart the API container to apply the changes using the command:
* `docker-compose up -d`

4) The automatized tests can be run using the command:

```
go test ./...
```