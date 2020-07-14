# Getting started

This document describes how to setup your local development environment for Finala.

## Preparation

Make sure the following tools are installed:

* Docker
* Golang 1.13.0+ ([installation manual](https://golang.org/dl/))
* ElasticSearch

Fork [Finala project](https://github.com/similarweb/finala)

### ElasticSearch

```shell
docker run -p 9200:9200 -p 5601:5601 nshou/elasticsearch-kibana
```

**Running the different Finala components**:

### Collector

Please refer to [configuration example](../../configuration/collector.yaml) file to see additional configurations.

```shell
go run main.go collector -c ./configuration/collector.yaml
```

#### Notifier

Please refer to [configuration example](../../configuration/notifier.yaml) file to see additional configurations.

```shell
go run main.go notifier -c ./configuration/notifier.yaml
```

#### API

Please refer to [configuration example](../../configuration/api.yaml) file to see additional configurations.

```shell
go run main.go api -c ./configuration/api.yaml
```

#### UI

Please refer to [configuration example](../../configuration/ui.yaml) file to see additional configurations.

```shell
cd ui
npm run dev
```

#### OR

```shell
make build-ui
go run main.go ui -c ./configuration/ui.yaml
```

### Docker

Running all components using `docker-compose`:

```shell
docker-compose up
```

UI is exposed on port 8080 ([quick link](http://127.0.0.1:8080))
