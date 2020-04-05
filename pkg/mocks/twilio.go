package mocks

import (
	"github.com/sfreiberg/gotwilio"
	"github.com/stretchr/testify/mock"
)

// TwilioClientMock is a mock for Twilio
type TwilioClientMock struct {
	mock.Mock
}

// SendSMS mocks sending Twilio SMS
func (m *TwilioClientMock) SendSMS(from, to, body, statusCallback, applicationSid string) (*gotwilio.SmsResponse, *gotwilio.Exception, error) {
	m.Called(from, to, body, statusCallback, applicationSid)
	return nil, nil, nil
}
