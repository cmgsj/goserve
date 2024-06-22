package files

import (
	"net/url"
	"strconv"
)

func parseQueryBool(u *url.URL, key string) bool {
	values, ok := u.Query()[key]
	if !ok {
		return false
	}

	if len(values) == 0 || values[0] == "" {
		return true
	}

	b, err := strconv.ParseBool(values[0])
	if err != nil {
		return false
	}

	return b
}
