We use `docker-compose` to run all Finala's components locally.  
This section describes how to use docker-compose in order to get a setup of Finala up and running.  

You must run the commands in the directory in which `docker-compose.yml` is located.
See the [Docker Compose command-line reference](https://docs.docker.com/compose/reference/) for more information about `docker-compose`.

## Create Finala

* Before creating a new setup of Finala, the only thing you need to configure is your AWS keys per account in the [collector's configuration file](../../../finala/configuration/collector.yaml#L10)
* We've already provided list of built-in cost-optimization `metrics`, you may modify the [collector.yaml](../../../finala/configuration/collector.yaml#L17) to suit your needs.

To create Finala, run the following command.

```sh
sudo docker-compose up -d                                                             Defaulting to a blank string.
Starting finala_elasticsearch_1 ... done
Starting finala_api_1           ... done
Starting finala_collector_1     ... done
Starting finala_ui_1            ... done
```

* UI is exposed on port 8080 ([Quick Link](http://127.0.0.1:8080))

## Stop Finala

To stop Finala, run the following command.

```sh
sudo docker-compose stop
Stopping finala_collector_1     ... done
Stopping finala_ui_1            ... done
Stopping finala_api_1           ... done
Stopping finala_elasticsearch_1 ... done
```

## Restart Finala

To restart Finala, run the following command.

```sh
docker-compose restart
Restarting finala_collector_1     ... done
Restarting finala_ui_1            ... done
Restarting finala_api_1           ... done
Restarting finala_elasticsearch_1 ... done
```


## Remove Finala

To remove Finala containers, run the following command.

```sh
 docker-compose down
Stopping finala_ui_1            ... done
Stopping finala_api_1           ... done
Stopping finala_elasticsearch_1 ... done
Removing finala_collector_1     ... done
Removing network finala_backend
```