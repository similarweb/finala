# Finala [![codecov](https://codecov.io/gh/similarweb/finala/branch/master/graph/badge.svg)](https://codecov.io/gh/similarweb/finala) ![Lint](https://github.com/similarweb/finala/workflows/Lint/badge.svg) ![Fmt](https://github.com/similarweb/finala/workflows/Fmt/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/similarweb/finala)](https://goreportcard.com/report/github.com/similarweb/finala) [![Gitter](https://badges.gitter.im/similarweb-finala/community.svg)](https://gitter.im/similarweb-finala/community?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge)

![alt Logo](https://raw.githubusercontent.com/similarweb/finala/master/docs/images/main-logo.png)
![Finala Processing](docs/images/finala.png)

----

## Overview

Finala is an open-source resource cloud scanner that analyzes, discloses, presents and notifies about wasteful and unused resources.

With Finala you can achieve 2 main objectives: **Cost saving & Unused resources detection**.

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
Resource            | Potential Cost Optimization| Unused Resource         |
--------------------| ---------------------------|-------------------------|
DocumentDB          | :ballot_box_with_check:    | :heavy_minus_sign:
DynamoDB            | :ballot_box_with_check:    | :heavy_minus_sign:
EC2 Instances       | :ballot_box_with_check:    | :heavy_minus_sign:
EC2 ELB             | :ballot_box_with_check:    | :heavy_minus_sign:
EC2 ALB,NLB         | :ballot_box_with_check:    | :heavy_minus_sign:
EC2 Volumes         | :ballot_box_with_check:    | :heavy_minus_sign:
EC2 Elastic IPs     | :ballot_box_with_check:    | :heavy_minus_sign:
ElasticCache        | :ballot_box_with_check:    | :heavy_minus_sign:
IAM User            | :heavy_minus_sign:         | :ballot_box_with_check:
Kinesis             | :ballot_box_with_check:    | :heavy_minus_sign:
Lambda              | :heavy_minus_sign:         | :ballot_box_with_check:
Neptune             | :ballot_box_with_check:    | :heavy_minus_sign:
RDS                 | :ballot_box_with_check:    | :heavy_minus_sign:
RedShift            | :ballot_box_with_check:    | :heavy_minus_sign:
ElasticSearch       | :ballot_box_with_check:    | :heavy_minus_sign:
API Gateway         | :heavy_minus_sign:         | :ballot_box_with_check:

## Installation

Please refer to [Installation instructions](https://finala.io/docs/installation/getting-started).

## Documentation & Guides

Documentation is available on the Finala website [here](https://finala.io/).

## Community, discussion, contribution, and support

You can reach the Finala community and developers via the following channels:

* [Gitter Community](https://gitter.im/similarweb-finala/community):
  * [finala-users](https://gitter.im/similarweb-finala/users)
  * [finala-developers](https://gitter.im/similarweb-finala/developers)

## Contributing

Thank you for your interest in contributing! Please refer to [Contribution guidelines](https://finala.io/docs/contributing/submitting-pr) for guidance.
