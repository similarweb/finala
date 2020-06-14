package notifiers

import (
	notifierCommon "finala/notifiers/common"
	"fmt"

	"github.com/pkg/errors"
)

// Register adds a notifier ctor to registeredNotifiers map
func Register(name notifierCommon.NotifierName, newFunc NotifierMaker) {
	registeredNotifiers[name] = newFunc
}

// Deregister removes a notifier ctor from the registeredNotifiers map
func Deregister(name notifierCommon.NotifierName) {
	delete(registeredNotifiers, name)
}

// GetNotifierMaker retrieves a notifier ctor from the registeredNotifiers map by name
func GetNotifierMaker(name notifierCommon.NotifierName) (notifierMaker NotifierMaker, err error) {
	var implemented bool

	if notifierMaker, implemented = registeredNotifiers[name]; !implemented {
		err = errors.New(fmt.Sprintf(notRegisteredTemplate, name))
		return notifierMaker, err
	}

	return notifierMaker, nil
}
