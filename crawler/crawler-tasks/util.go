package crawler_tasks

import "net/http"

var requiredCookies = []string{
	"ct0",
}

var requiredHeaders = []string{
	"authorization",
}

func setCookiesForRequest(req *http.Request, cookies map[string]string) {
	for name, value := range cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}
}

func checkRequiredCookiesExists(cookies map[string]string) bool {
	for _, reqName := range requiredCookies {
		if _, ok := cookies[reqName]; !ok {
			return false
		}
	}
	return true
}

func checkRequiredHeadersExists(headers map[string]string) bool {
	for _, reqName := range requiredHeaders {
		if _, ok := headers[reqName]; !ok {
			return false
		}
	}
	return true
}
