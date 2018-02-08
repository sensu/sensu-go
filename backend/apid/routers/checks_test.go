package routers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sensu/sensu-go/backend/apid/actions"
	"github.com/sensu/sensu-go/testing/mockqueue"
	"github.com/sensu/sensu-go/testing/mockstore"
	"github.com/sensu/sensu-go/testing/testutil"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/mock"
)

func TestHttpApiEventsGet(t *testing.T) {

}

func TestHttpApiChecksAdhocRequest(t *testing.T) {
	defaultCtx := testutil.NewContext(
		testutil.ContextWithOrgEnv("default", "default"),
		testutil.ContextWithRules(
			types.FixtureRuleWithPerms(types.RuleTypeCheck, types.RulePermCreate, types.RulePermRead),
		),
	)

	store := &mockstore.MockStore{}
	queue := &mockqueue.MockQueue{}
	adhocRequest := types.FixtureAdhocRequest("check1", []string{"subscription1", "subscription2"})
	checkConfig := types.FixtureCheckConfig("check1")
	store.On("NewQueue", mock.Anything, mock.Anything).Return(queue)
	store.On("GetCheckConfigByName", mock.Anything, mock.Anything).Return(checkConfig, nil)
	queue.On("Enqueue", mock.Anything, mock.Anything).Return(nil)
	checkController := actions.NewCheckController(store)
	c := &ChecksRouter{controller: checkController}
	payload, _ := json.Marshal(adhocRequest)
	req, err := http.NewRequest(http.MethodPost, "/checks/check1/execute", bytes.NewBuffer(payload))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(c.adhocRequest)
	handler.ServeHTTP(rr, req.WithContext(defaultCtx))

	if status := rr.Code; status != http.StatusAccepted {
		t.Errorf("handler returned incorrect status code: %v want %v", status, http.StatusAccepted)
	}
}
