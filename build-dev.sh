docker build -t devbox-dev --rm -f Dockerfile.dev . && \
docker run --rm -d --entrypoint=sh --name devbox devbox-dev -c "sleep 10" && \
docker cp devbox:/devbox .