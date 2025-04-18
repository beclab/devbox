ARG VARIANT=1.21-bookworm
FROM golang:${VARIANT}

RUN curl -fsSL https://code-server.dev/install.sh | sh -s -- --prefix=/usr/local --version=4.12.0

RUN apt-get update && apt-get install -y nginx gh

RUN mkdir -p /opt/html
RUN mkdir -p /etc/nginx/conf.d/dev
COPY containers/root/. /opt/html/.
COPY containers/conf/. /etc/nginx/conf.d/.

RUN ln -sf /dev/stdout /var/log/nginx/access.log && ln -sf /dev/stderr /var/log/nginx/error.log
RUN echo 'export PATH="/go/bin:/usr/local/go/bin:$PATH"' >> /etc/profile


EXPOSE 8080

CMD [ "/bin/sh", "-c", "nginx && exec /usr/bin/code-server --bind-addr \"0.0.0.0:3000\" --auth=none --log=debug" ]


