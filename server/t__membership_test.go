package server

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/mikeydub/go-gallery/persist"
	"github.com/mikeydub/go-gallery/util"
	"github.com/stretchr/testify/assert"
)

func TestMembership_Success(t *testing.T) {
	assert := setupTest(t)

	resp := membershipRequest(assert)
	defer resp.Body.Close()
	assertValidJSONResponse(assert, resp)

	type response struct {
		MembershipTiers []*persist.MembershipTier `json:"tiers"`
		Error           string                    `json:"error"`
	}

	membershipTiers := &response{}
	err := util.UnmarshallBody(&membershipTiers, resp.Body)
	assert.Nil(err)
	assert.Empty(membershipTiers.Error)
	assert.Greater(len(membershipTiers.MembershipTiers), 0)
}

func membershipRequest(assert *assert.Assertions) *http.Response {

	req, err := http.NewRequest("GET",
		fmt.Sprintf("%s/users/membership", tc.serverURL),
		nil)
	assert.Nil(err)
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Do(req)
	assert.Nil(err)
	return resp
}
