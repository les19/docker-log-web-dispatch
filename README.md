# Container logs web dispatcher.

Simple app that would listen logs of your docker compose containers and send them via http to the endpoint you define.

## How to use

### Step 1.
Add logger container into your compose.yml.

```yml
services:
    # your awesome application that would produce logs
    application:
        image: some-image
        # probably you're gonna spin up multiple nodes, otherwise why would you need this app at all
        deploy:
            replicas: 4
        networks:
            - main

    logger:
        image: 'ghcr.io/les19/docker-log-web-dispatch:main'
        volumes:
            - '/var/run/docker.sock:/var/run/docker.sock'
        networks:
            - main
        deploy:
            mode: global
        depends_on:
            - application

networks:
    main:
        driver: bridge
```

### Step 2.
Define your endpoint and optionally auth header via ENV vars.
```yml
    logger:
        image: 'ghcr.io/les19/docker-log-web-dispatch:main'
        volumes:
            - '/var/run/docker.sock:/var/run/docker.sock'
        networks:
            - main
        deploy:
            mode: global
        depends_on:
            - application
        environment:
            LOGGER_SERVICE_URL: 'logger.my-awesome-application.com'
            LOGGER_AUTH_HEADER_NAME: 'X-Auth-Token'
            LOGGER_AUTH_HEADER_VALUE: 'some-random-token'
            # you can filter which container's logs you care about, by defining name filter variable
            CONTAINER_NAME_FILTERS: 'application'
```

## Check example

### Step 1.
Go to [example directory](https://github.com/les19/docker-log-web-dispatch/tree/main/example).
```bash
cd example
```

### Step 2.
Put into .env your LOGGER_SERVICE_URL variable.
You can use https://webhook.site for testing to send your logs in case you do not have an endpoint for now.
```bash
webhook_site_url='https://webhook.site/' && webhook_view_url='https://webhook.site/#!/view/' \
    && webhook_site_uuid=$(curl -X POST https://webhook.site/token | jq -r '.uuid') \
    && echo "LOGGER_SERVICE_URL=$webhook_site_url$webhook_site_uuid" > .env \
    && echo "LOGGER_WEBHOOK_VIEW_URL=\"$webhook_view_url$webhook_site_uuid\"" >> .env
```

### Step 3.
Run example docker compose stack
```bash
docker compose up -d
```

### Step 4.
Navigate to URL you'll see at LOGGER_WEBHOOK_VIEW_URL env variable. You should see your service log messages appearing.

> NOTE: variable *LOGGER_WEBHOOK_VIEW_URL* does not required by service to run.
> Added to the example excript for observability purpose only
