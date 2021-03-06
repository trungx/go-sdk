package web

import (
	"testing"

	"net/http"

	"github.com/blend/go-sdk/assert"
	"github.com/blend/go-sdk/exception"
)

func TestMockRequestBuilderWithPathf(t *testing.T) {
	assert := assert.New(t)
	rb := NewMockRequestBuilder(nil)
	rb.WithPathf("/test/%s", "foo")
	assert.Equal("/test/foo", rb.path)
}

func TestMockRequestBuilderWithVerb(t *testing.T) {
	assert := assert.New(t)
	rb := NewMockRequestBuilder(nil)
	rb.WithVerb("get")
	assert.Equal("GET", rb.verb)
}

func TestMockRequestBuilderWithQueryString(t *testing.T) {
	assert := assert.New(t)
	rb := NewMockRequestBuilder(nil)
	rb.WithQueryString("foo", "bar")
	assert.Equal("bar", rb.queryString.Get("foo"))
}

func TestMockRequestBuilderFetchResponseAsBytes(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.GET("/test_path", func(r *Ctx) Result {
		return r.Raw([]byte("test"))
	})
	resBody, err := app.Mock().WithPathf("/test_path").Bytes()
	assert.Nil(err)
	assert.NotEmpty(resBody)
	assert.Equal("test", string(resBody))
}

func TestMockRequestBuilderJSON(t *testing.T) {
	assert := assert.New(t)
	app := New()
	app.GET("/test_path", func(r *Ctx) Result {
		return r.JSON().Result([]string{"foo", "bar"})
	})
	var res []string
	err := app.Mock().WithPathf("/test_path").JSON(&res)
	assert.Nil(err)
	assert.NotEmpty(res)
	assert.Equal("foo", res[0])
	assert.Equal("bar", res[1])
}

func TestMockRequestBuilderPanicHandler(t *testing.T) {
	assert := assert.New(t)

	var didPanic bool

	app := New()
	app.WithPanicAction(func(r *Ctx, err interface{}) Result {
		didPanic = true
		return r.Text().InternalError(exception.New(err))
	})
	app.GET("/test_path", func(r *Ctx) Result {
		panic("this is only a test")
	})

	meta, err := app.Mock().Get("/test_path").ExecuteWithMeta()
	assert.Nil(err)
	assert.True(didPanic)
	assert.Equal(http.StatusInternalServerError, meta.StatusCode)
}
