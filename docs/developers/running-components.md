**Running the different components**:

#### Collector

```shell
go run main.go collector -c ./configuration/collector.yaml
```

#### Notifier

```shell
go run main.go notifier -c ./configuration/notifier.yaml
```

#### API
```shell
go run main.go api -c ./configuration/api.yaml
```

#### UI

```shell
cd ui
npm run dev
```

*OR*

```shell
make build-ui
go run main.go ui -c ./configuration/ui.yaml
```

### Docker
Running all components using `docker-compose`:

```
docker-compose up
```
UI is exposed on port 8080 ([quick link](http://127.0.0.1:8080)).
