# Finala

A resource cloud scanner that analyzes and reports about wasteful and unused resources to cut unwanted expenses.
The tool is based on yaml definitions (no code), by default configuration OR given yaml file and the report output will be saved in a given storage.

Currently it is implemented for AWS resources (RDS, EC2 instances, DynamoDB, ElasticCache, documentDB, ELB) and can be easily extended.

## Getting Started

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### How To Use

All the logic is contained inside [config.yaml](./config.yaml). 
1. Setup your Cloud provider (currently AWS only) credentials and accounts you want to analyze.

```yaml
providers:
  aws:
  - name: <ACCOUNT_NAME>
    access_key: <ACCESS_KEY>
    secret_key: <SECRET_KEY>
    regions:
      - <REGION>
```
2. Let it [run](#Installing)! 

*Optional:* There are defaults but, You can specify your own resources to analyze and change the metrics thresholds.

### For example: 

If you want to test RDS resources that had no zero connections in the last week: 

```yaml
rds:
    - description: Database connection count
        metrics:
        - name: DatabaseConnections
            statistic: Sum
        period: 24h 
        start_time: 168h # 24(h) * 7(d) = 168h
        constraint:
        operator: "=="
        value: 0
```

### Prerequisites

1. AWS access key and secret key (with readonly access) 
2. Optional: Docker

### Installing

1) Build from source

```
$ Git clone git@github.com:kaplanelad/finala.git
$ make build
```

To run:
```
$  ./finala aws -c ${PWD}/config.yaml
```

For config [example](./config.yaml)

2) Download the binary
https://github.com/similarweb/finala/releases


## Running the tests

```
$ make test

$ make test-html
```

## Built With

* [GO](https://golang.org/)
* [AWS SDK](https://aws.amazon.com/tools/) 

## Contributing

All pull requests and issues are more then welcome!

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

