package rabbitmq

import (
	"fmt"
	"strings"

	amqp "github.com/rabbitmq/amqp091-go"
)

// redactURL returns the provided url without the credentials (it can be used in logs) and the password from the url
func redactURL(url string) (string, string, error) {
	uri, err := amqp.ParseURI(url)
	if err != nil {
		// the parse error can contain the url, so it is not returned
		return "", "", ErrInvalidURL
	}

	return fmt.Sprintf("%s://%s:%d/%s", uri.Scheme, uri.Host, uri.Port, strings.TrimPrefix(uri.Vhost, "/")), uri.Password, nil
}
