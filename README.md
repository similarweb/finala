# Finala [![codecov](https://codecov.io/gh/similarweb/finala/branch/master/graph/badge.svg)](https://codecov.io/gh/similarweb/finala) ![Lint](https://github.com/similarweb/finala/workflows/Lint/badge.svg) ![Fmt](https://github.com/similarweb/finala/workflows/Fmt/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/similarweb/finala)](https://goreportcard.com/report/github.com/similarweb/finala) [![Gitter](https://badges.gitter.im/similarweb-finala/community.svg)](https://gitter.im/similarweb-finala/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)
<p align="center">
    <img src="https://raw.githubusercontent.com/similarweb/finala/docs/update-readme-md/docs/images/logo.png" width="400">
</p>

---
## What is Finala?
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
Resource            | Potential Cost Optimization| Unused resource         |Config File                                          |
--------------------| ---------------------------|-------------------------|-----------------------------------------------------|
DocumentDB          | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L28), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L130)
DynamoDB            | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L69), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L180)
EC2 Instances       | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L78), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L170)
EC2 ELB             | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L48), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L150)
EC2 ALB,NLB         | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L58), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L160)
EC2 EBS             | :heavy_minus_sign:         | :ballot_box_with_check: |
ElasticCache        | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L38), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L140)
IAM User            | :heavy_minus_sign:         | :ballot_box_with_check: | [Collector Config](./configuration/collector.yaml#L153), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L249)
Kinesis             | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L126), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L228)
Lambda              | :heavy_minus_sign:         | :ballot_box_with_check: | [Collector Config](./configuration/collector.yaml#L103), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L205)
Neptune             | :ballot_box_with_check:    | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L113), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L216)
RDS                 | :ballot_box_with_check:    | :heavy_minus_sign:      |[Collector Config](./configuration/collector.yaml#L18), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L120)
RedShift            | :heavy_minus_sign:         | :heavy_minus_sign:      | [Collector Config](./configuration/collector.yaml#L142), [Helm Config](https://github.com/similarweb/finala-helm/blob/master/values.yaml#L244)

## **Screenshots**

### Dashboard
![alt Summary](https://raw.githubusercontent.com/similarweb/finala/docs/update-readme-md/docs/images/main-dashboard.png)

### Unused RDS report
![alt Resources](https://raw.githubusercontent.com/similarweb/finala/docs/update-readme-md/docs/images/resource.png)

### Notifications
![alt Slack](https://raw.githubusercontent.com/similarweb/finala/docs/update-readme-md/docs/images/slack.png)

## Getting Started
* The quickest way to get started with Finala is by using K8S. Get started with [Finala Helm chart](https://github.com/similarweb/finala-helm).


### Documentation & Guides
* [How to use](./docs/how-to-use.md): 
* [Developer Guide](./docs/developers/README.md):
* [Configuration Examples](./docs/configuration_examples/README.md):


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

