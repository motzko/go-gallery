package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mikeydub/go-gallery/runtime"
	"github.com/stretchr/testify/assert"
)

func TestHealthcheck(t *testing.T) {
	t.Cleanup(clearDB)
	assert := assert.New(t)

	resp, err := http.Get(fmt.Sprintf("%s/health", tc.serverURL))
	assert.Nil(err)
	assertValidJSONResponse(assert, resp)

	body := healthcheckResponse{}
	runtime.UnmarshallBody(&body, resp.Body, tc.r)
	assert.Equal("gallery operational", body.Message)
	assert.Equal("local", body.Env)
}
