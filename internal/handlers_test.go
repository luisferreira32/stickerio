package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type mockRepository struct {
	funcInsertEvent            func(ctx context.Context, e *event) error
	funcGetCity                func(ctx context.Context, id, playerID string) (*city, error)
	funcGetCityInfo            func(ctx context.Context, id string) (*city, error)
	funcGetMovement            func(ctx context.Context, id, playerID string) (*movement, error)
	funcGetUnitQueueItem       func(ctx context.Context, id, cityID string) (*unitQueueItem, error)
	funcGetBuildingQueueItem   func(ctx context.Context, id, cityID string) (*buildingQueueItem, error)
	funcListCityInfo           func(ctx context.Context, lastID string, pageSize int, filters ...listCityInfoFilterOpt) ([]*city, error)
	funcListMovements          func(ctx context.Context, lastID, playerID string, pageSize int, filters ...listMovementsFilterOpt) ([]*movement, error)
	funcListUnitQueueItems     func(ctx context.Context, cityID, lastID string, pageSize int) ([]*unitQueueItem, error)
	funcListBuildingQueueItems func(ctx context.Context, cityID, lastID string, pageSize int) ([]*buildingQueueItem, error)
}

func (m *mockRepository) InsertEvent(ctx context.Context, e *event) error {
	return m.funcInsertEvent(ctx, e)
}
func (m *mockRepository) GetCity(ctx context.Context, id, playerID string) (*city, error) {
	return m.funcGetCity(ctx, id, playerID)
}
func (m *mockRepository) GetCityInfo(ctx context.Context, id string) (*city, error) {
	return m.funcGetCityInfo(ctx, id)
}
func (m *mockRepository) ListCityInfo(ctx context.Context, lastID string, pageSize int, filters ...listCityInfoFilterOpt) ([]*city, error) {
	return m.funcListCityInfo(ctx, lastID, pageSize, filters...)
}
func (m *mockRepository) GetMovement(ctx context.Context, id, playerID string) (*movement, error) {
	return m.funcGetMovement(ctx, id, playerID)
}
func (m *mockRepository) ListMovements(ctx context.Context, lastID, playerID string, pageSize int, filters ...listMovementsFilterOpt) ([]*movement, error) {
	return m.funcListMovements(ctx, lastID, playerID, pageSize, filters...)
}
func (m *mockRepository) GetUnitQueueItem(ctx context.Context, id, cityID string) (*unitQueueItem, error) {
	return m.funcGetUnitQueueItem(ctx, id, cityID)
}
func (m *mockRepository) ListUnitQueueItems(ctx context.Context, cityID, lastID string, pageSize int) ([]*unitQueueItem, error) {
	return m.funcListUnitQueueItems(ctx, cityID, lastID, pageSize)
}
func (m *mockRepository) GetBuildingQueueItem(ctx context.Context, id, cityID string) (*buildingQueueItem, error) {
	return m.funcGetBuildingQueueItem(ctx, id, cityID)
}
func (m *mockRepository) ListBuildingQueueItems(ctx context.Context, cityID, lastID string, pageSize int) ([]*buildingQueueItem, error) {
	return m.funcListBuildingQueueItems(ctx, cityID, lastID, pageSize)
}

func unsafeJSONMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}

func contextWithValues(base context.Context, values map[ContextKey]string) context.Context {
	for k, v := range values {
		base = context.WithValue(base, k, v)
	}
	return base
}

func Test_GetWelcome(t *testing.T) {
	testcases := []struct {
		name     string
		request  *http.Request
		wantCode int
		wantResp []byte
	}{
		{
			name:     "get welcome works",
			request:  httptest.NewRequest("GET", "/api", http.NoBody),
			wantCode: 200,
			wantResp: []byte(`Welcome to Stickerio API.`),
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			handler := NewServerHandler(&mockRepository{})
			recorder := httptest.NewRecorder()
			http.HandlerFunc(handler.GetWelcome).ServeHTTP(recorder, testcase.request)

			if recorder.Code != testcase.wantCode {
				t.Errorf("unexpected status code: %d, want: %d", recorder.Code, testcase.wantCode)
			}
			if diff := cmp.Diff(recorder.Body.Bytes(), testcase.wantResp); diff != "" {
				t.Errorf("unexpected diff in response: %v", diff)
			}
		})
	}
}
