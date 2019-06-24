package testing

import "time"

// MockNowService should be used for tests
type MockNowService struct {
	now time.Time
}

// Now returns now value
func (svc *MockNowService) Now() time.Time {
	return svc.now
}

// SetNow set current now value
func (svc *MockNowService) SetNow(val time.Time) {
	svc.now = val
}

// NewMockNowService returns an instance of a now service
func NewMockNowService(now time.Time) *MockNowService {
	return &MockNowService{now: now}
}
