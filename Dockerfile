FROM nginx:stable-alpine

COPY packages/web/dist/spa /app

RUN apk --no-cache add nginx

RUN echo '  \
server {  \
	listen 80 default_server;  \
	root /app;  \
\
    location / { \
		add_header Cache-Control "no-store"; \
		try_files $uri $uri/index.html /index.html; \
	} \
\
	location /api/ {  \
        proxy_pass http://devbox-server:8080;  \
        proxy_set_header            Host $http_host;  \
        proxy_set_header            X-real-ip $remote_addr;  \
        proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;  \
        proxy_http_version 1.1;  \
        proxy_set_header Upgrade $http_upgrade;  \
        proxy_set_header Connection $http_connection;  \
        proxy_set_header Accept-Encoding gzip;  \
        proxy_read_timeout 180; \ 
	}  \
 \
}' > /etc/nginx/conf.d/default.conf

RUN ln -sf /dev/stdout /var/log/nginx/access.log && ln -sf /dev/stderr /var/log/nginx/error.log

EXPOSE 80

CMD [ "nginx", "-g", "daemon off;" ]