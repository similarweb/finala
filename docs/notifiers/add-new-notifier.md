## How To A new Notifier?


### Adding a new Notifier
This section describes how to add a new Notifier to Finala.

```go
// RegisterNotifiers registers existing notifier ctor to the ctor map we use to initiate all notifiers
func RegisterNotifiers() {
    notifiers.Register("slack", slack.NewManager)
    notifiers.Register("NewNotifierName",NewNotifierManager)
}
```

### Pay Attention

* Don't forget to add a new notifier to [RegisterNotifiers](../../notifiers/load/load.go#L10) function.
* Each new Notifier will have to implement the interface of [Notifier](../../notifiers/common/common.go#L7)
* Update the [Notifier Configuration](../../configuration/notifier.yaml) with the new Notifier configuration section.