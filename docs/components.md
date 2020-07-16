### Components

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
