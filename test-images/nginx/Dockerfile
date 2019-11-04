FROM docker.io/nginx:latest

COPY teapot.conf /etc/nginx/teapot.conf
COPY docker-entrypoint.sh /usr/local/bin/docker-entrypoint.sh
COPY 418.html /usr/share/nginx/html/418.html

RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["/usr/local/bin/docker-entrypoint.sh"]
