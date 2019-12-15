package notifiers_test
import (
    "finala/notifiers"
    "os"
    "testing"
)

func TestSendMsg(t *testing.T) {
    webhookUrl := os.Getenv("FINALA_SLACK")
    err := notifiers.SendSlackNotification(webhookUrl, "Test Message from finala")
    if err != nil {
        t.Fatalf(err.Error())
    }
}