This section describes how to perform a new installation of Finala on a local machine or Kubernetes.

## Docker Compose
These instructions will get you a copy of the project up and running on your local.
Please refer [Local Installation](./local-install.md)


## Deploy Finala on Kubernetes
You can use Helm to install Finala on a Kubernetes cluster see [Deploying Finala on Kubernetes]((https://github.com/similarweb/finala-helm)).

## Finala Components

The table below lists the some of the components of Finala.

|Component    |Description                                            | Kind      |
|-------------|-------------------------------------------------------|-----------|
|Elasticsearch| Current Storage for Finala                            | Deployment|
|Collector    | Collects the data for all the cloud provider resources| CronJob   |
|Notifier     | Notifies a user or a group for unutilized resources   | CronJob   |
|API          | Works with the Storage to save/get Data               | Deployment|
|UserInterface| Queries the API and shows the Finala Dashboard        | Deployment|

