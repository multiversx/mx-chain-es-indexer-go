package mock

import "net/http"

// TransportMock -
type TransportMock struct {
	Response *http.Response
	Err      error
}

// RoundTrip -
func (m *TransportMock) RoundTrip(_ *http.Request) (*http.Response, error) {
	return m.Response, m.Err
}
