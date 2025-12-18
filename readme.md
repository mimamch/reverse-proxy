# Reverse Proxy

A reverse proxy implementation for routing and load balancing HTTP requests.

## Features

- Request routing
- Load balancing: Round Robin
- SSL termination
- SSL Generation using Let's Encrypt
- Zero downtime reloads

# Installation

1. Ensure you have Docker and Docker Compose installed on your system.
2. Create a `docker-compose.yml` file with the following content:

```yaml
services:
  reverse-proxy:
    container_name: reverse-proxy
    image: ghcr.io/mimamch/reverse-proxy:latest
    restart: unless-stopped
    ports:
      - "80:80"
      - "443:443"
    environment:
      - DATABASE_URL=postgres://user:password@db:5432/reverse-proxy
      - EMAIL=youremail@mail.com
```

3. Run the following command to start the reverse proxy:

```bash
docker-compose up -d
```

# TODO:

- Make GUI for easier management of routes and settings.
