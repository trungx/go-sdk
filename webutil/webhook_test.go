package webutil

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blend/go-sdk/assert"
)

func TestWebhookSend(t *testing.T) {
	assert := assert.New(t)

	var bodyCorrect, methodCorrect, headerCorrect bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		bodyCorrect = string(body) == `this is only a test`
		methodCorrect = r.Method == "POST"
		headerCorrect = r.Header.Get("X-Test-Value") == "foo"

		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK!\n")
	}))
	defer ts.Close()

	wh := Webhook{
		URL:    ts.URL,
		Method: "POST",
		Headers: map[string]string{
			"X-Test-Value": "foo",
		},
		Body: "this is only a test",
	}

	res, err := wh.Send()
	assert.Nil(err)
	assert.Equal(http.StatusOK, res.StatusCode)

	assert.True(bodyCorrect)
	assert.True(methodCorrect)
	assert.True(headerCorrect)
}
