# Finala [![codecov](https://codecov.io/gh/similarweb/finala/branch/master/graph/badge.svg)](https://codecov.io/gh/similarweb/finala) ![Lint](https://github.com/similarweb/finala/workflows/Lint/badge.svg) ![Fmt](https://github.com/similarweb/finala/workflows/Fmt/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/similarweb/finala)](https://goreportcard.com/report/github.com/similarweb/finala) [![Gitter](https://badges.gitter.im/similarweb-finala/community.svg)](https://gitter.im/similarweb-finala/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
<p align="center">
    <img src="https://raw.githubusercontent.com/similarweb/finala/docs/update-readme-md/docs/images/logo.png" width="400">
</p>

---
Finala is an open-source resource cloud scanner that analyzes, discloses, and notifies about wasteful and unused resources in your company's infrastructure.

Finala has 2 main objectives:

* Help your organization being cost-effective
* Find unused resources

## Features
* **YAML Definitions**: Resources definitions are described using a high-level YAML configuration syntax. This allows Finala consumers easily tweak the configuration to help it understand their infrastructure, spending habits and normal usage.
* **1 Click Deployment**: Finala can be deployed via Docker compose or a [Helm chart](https://github.com/similarweb/finala-helm).
* **Graphical user interface**: Users can easily explore and investigate unused or unutilized resources in your cloud provider.
* **Resource Filtering by Cloud Provider Tags**: Users can filter unused resources by just providing the tags you are using in your cloud provider.
* **Schedule Pro Active Notifications**: Finala has the ability to configure scheduled based notifications to a user or a group. 

## Supported Services
AWS:
Resource            | Sub Resources|Sub Resources|
--------------------| -------------|-------------|
DocumentDB          | -            |
DynamoDB            | -            |
EC2                 |              | Instances`|`ELB`|` NLB`|` ALB`|` EBS
ElasticCache        | -            |
IAM User            | -            |
Kinesis             | -            |
Lambda              | -            |
Neptune             | -            |
RDS                 | -            |
RedShift            | -            |

## **Screenshots**

### Dashboard
![alt Summary](https://raw.githubusercontent.com/similarweb/finala/docs/update-readme-md/docs/images/main-dashboard.png)

### Unused RDS report
![alt Resources](https://raw.githubusercontent.com/similarweb/finala/docs/update-readme-md/docs/images/resource.png)

### Notifications
![alt Slack](https://raw.githubusercontent.com/similarweb/finala/docs/update-readme-md/docs/images/slack.png)

## Getting Started

* [Configuration Examples](./docs/configuration_examples/README.md)

These instructions will get you a copy of the project up and running on your local machine for development and testing purposes. See deployment for notes on how to deploy the project on a live system.

### How To Use

Finala is built from 4 components:

* **API** - RESTful API server that receives events from the collector and serves the UI. See [example API configuration file](./configuration/api.yaml).

* **UI** - The User Interface, displays the data in a way that it'll look nice :).

* **Notifier** - Notifies notification groups with the support of multiple notifiers defined in [notifier.yaml](./configuration/notifier.yaml).
All resources that marked as "under utilized" are reported to the notification groups.
Currently we only support Slack notifier type [notifier.yaml](./configuration/notifier.yaml).
If you wish to contribute and add a new Notifier please read [How To add a new Notifier?](docs/notifiers/add-new-notifier.md)

* **Collector** - Collects and analyzes resources against their thresholds defined in [collector.yaml](./configuration/collector.yaml).
All resources that marked as "under utilized" are reported back to the API component.
You can define multiple accounts and regions in the [collector.yaml](./configuration/collector.yaml) file.

```yaml
providers:
  aws:
  - name: <ACCOUNT_NAME>
    # Environment variables will be used in case if these variables are absent
    access_key: <ACCESS_KEY>
    secret_key: <SECRET_KEY>
    session_token: "" # Optional variable, on default this variable not set
    regions:
      - <REGION>
```
We've already provided list of built-in cost-optimization `metrics`, you may modify the [collector.yaml](./configuration/collector.yaml) to suit your needs.
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

This example will mark RDS as under utilized if that RDS had **zero** connections in the last week.

### Deploy
You may use either approach in order to deploy Finala.

* Deploy on Kubernetes, see [Helm chart](https://github.com/similarweb/finala-helm) for more information.
* Run it locally with `docker-compose up`.

## Community, discussion, contribution, and support

You can reach the Finala community and developers via the following channels:
* [Gitter Community](https://gitter.im/similarweb-finala/community):
    * [finala-users](https://gitter.im/similarweb-finala/users)
    * [finala-developers](https://gitter.im/similarweb-finala/developers)


## Contributing to Finala
Thank you for your interest in contributing! Please refer to [Contribution guidelines](./CONTRIBUTING.md) for guidance.

