# Finala [![codecov](https://codecov.io/gh/similarweb/finala/branch/master/graph/badge.svg)](https://codecov.io/gh/similarweb/finala) ![Lint](https://github.com/similarweb/finala/workflows/Lint/badge.svg) ![Fmt](https://github.com/similarweb/finala/workflows/Fmt/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/similarweb/finala)](https://goreportcard.com/report/github.com/similarweb/finala) [![Gitter](https://badges.gitter.im/similarweb-finala/community.svg)](https://gitter.im/similarweb-finala/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

![alt Logo](https://raw.githubusercontent.com/similarweb/finala/master/docs/images/main-logo.png)

----

## What is Finala?

Finala is an open-source resource cloud scanner that analyzes, discloses, and notifies about wasteful and unused resources in your company's infrastructure.

Finala has 2 main objectives:

* Cost saving
* Unused resources detection

## Features

* **YAML Definitions**: Resources definitions are described using a high-level YAML configuration syntax. This allows Finala consumers easily tweak the configuration to help it understand their infrastructure, spending habits and normal usage.
* **1 Click Deployment**: Finala can be deployed via Docker compose or a [Helm chart](https://github.com/similarweb/finala-helm).
* **Graphical user interface**: Users can easily explore and investigate unused or unutilized resources in your cloud provider.
* **Resource Filtering by Cloud Provider Tags**: Users can filter unused resources by just providing the tags you are using in your cloud provider.
* **Schedule Pro Active Notifications**: Finala has the ability to configure scheduled based notifications to a user or a group.


## Supported Services
* **Potential Cost Optimization** - is the price you can save for untilized resources in your infrastructure
* **Unused Resource** - are resources which don't necessarily cost money and can be removed.

AWS:
Resource            | Potential Cost Optimization| Unused Resource         |Config File                                          |
--------------------| ---------------------------|-------------------------|-----------------------------------------------------|
DocumentDB          | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L29), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L131)
DynamoDB            | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L84), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L186)
EC2 Instances       | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L73), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L175)
EC2 ELB             | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L51), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L154)
EC2 ALB,NLB         | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L62), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L164)
EC2 Volumes         | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L189), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L291)
EC2 Elastic IPs     | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L186), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L288)
ElasticCache        | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L40), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L142)
IAM User            | :heavy_minus_sign:         | :ballot_box_with_check: | [Collector Config](./configuration/collector.yaml#L179), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L281)
Kinesis             | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L136), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L238)
Lambda              | :heavy_minus_sign:         | :ballot_box_with_check: | [Collector Config](./configuration/collector.yaml#L111), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L213)
Neptune             | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L111), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L224)
RDS                 | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L18), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L120)
RedShift            | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L153), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L255)
ElasticSearch       | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L164), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L266)
API Gateway         | :heavy_minus_sign:         | :ballot_box_with_check: | [Collector Config](./configuration/collector.yaml#L190)

## Screenshots

### Dashboard

![alt Summary](https://raw.githubusercontent.com/similarweb/finala/master/docs/images/dashboard.png)

### Unused RDS report

![alt Resources](https://raw.githubusercontent.com/similarweb/finala/master/docs/images/resource.jpg)

### Notifications

![alt Slack](https://raw.githubusercontent.com/similarweb/finala/master/docs/images/slack.png)

### Installation

* Please refer to [Installation instructions](docs/install/index.md).

### Documentation & Guides

* [Components](./docs/components.md): List of Finala components.
* [Configuration Examples](./docs/configuration_examples/README.md): See configuration examples and explanations.
* [Developer Guide](./docs/developers/index.md):  If you are interested in contributing, read the developer guide.

## Community, discussion, contribution, and support

You can reach the Finala community and developers via the following channels:

* [Gitter Community](https://gitter.im/similarweb-finala/community):
  * [finala-users](https://gitter.im/similarweb-finala/users)
  * [finala-developers](https://gitter.im/similarweb-finala/developers)

## Contributing to Finala

Thank you for your interest in contributing! Please refer to [Contribution guidelines](./CONTRIBUTING.md) for guidance.