# Studio

Olares App development management tools

## Build backend
### Build binary
```shell
make studio-server
```

### Build image
```shell
make docker-build-server IMG=<some-registry>/studio-server:tag
```

### Push image
```shell
make docker-push-server IMG=<some-registry>/studio-server:tag
```


## Build frontend
### Install the dependencies
```shell
npm install
```
### Start the app in development mode (hot-code reloading, error reporting, etc.)

```bash
npm run dev
```

### Build the app for production

```bash
npm run build
```

### Build image
```shell
make docker-build-frontend IMG=<some-registry>/studio:tag
```

### Push image
```shell
make docker-build-frontend IMG=<some-registry>/studio:tag
```