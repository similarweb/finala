# Configuration Examples
This section explains how some of the Finala's collector configuration work.

The full working example can be found here [collector.yaml](./../../configuration/collector.yaml).
<hr>

1. Find EC2 instances which have less that 5% CPU usage in the last week.
```yaml
ec2:
  - description: CPU utilization
    enable: true
    metrics:
      - name: CPUUtilization
        statistic: Maximum
    period: 24h
    start_time: 168h # 24h * 7d
    constraint:
      operator: "<"
      value: 5
```

2. Find RDS DB's that had zero connections in the last week.

```yaml
rds:
  - description: Connection count
    enable: true
    metrics:
      - name: DatabaseConnections
        statistic: Sum
    period: 24h 
    start_time: 168h # 24h * 7d
    constraint:
      operator: "=="
      value: 0
```

3. Find ELB's that had zero traffic (requests) in the last week.

```yaml
elb:
  - description: Request count
    enable: true
    metrics:
      - name: RequestCount
        statistic: Sum
    period: 24h 
    start_time: 168h # 24h * 7d
    constraint:
      operator: "=="
      value: 0
```

4. Find Kinesis streams which don't have put records requests in the last week.
```yaml
kinesis:
  - description: Total put records
    enable: true
    metrics:
      - name: "PutRecords.Bytes"
        statistic: Sum
      - name: "PutRecord.Bytes"
        statistic: Sum
    period: 24h
    start_time: 168h # 24h * 7d
    constraint:
      # The go module Knetic/govaluate has a built in escaping
      # https://github.com/Knetic/govaluate#escaping-characters
      # [PutRecord.Bytes] will escape the parameter name
      formula: "[PutRecord.Bytes] + [PutRecords.Bytes]"
      operator: "=="
      value: 0
```