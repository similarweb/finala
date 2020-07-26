package testutils

import (
	notifier "github.com/similarweb/client-notifier"
)

type MockVersion struct {
}

func NewMockVersion() *MockVersion {
	return &MockVersion{}
}

func (v *MockVersion) Get() (*notifier.Response, error) {
	response := &notifier.Response{
		CurrentDownloadURL: "http://localhost",
		CurrentVersion:     "0.0.1",
		Outdated:           true,
	}
	return response, nil
}
