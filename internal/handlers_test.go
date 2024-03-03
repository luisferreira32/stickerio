package internal

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/luisferreira32/stickerio"
)

type mockRepository struct {
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

func toCityInfo(v []byte) *stickerio.CityInfo {
	ci := &stickerio.CityInfo{}
	err := json.Unmarshal(v, ci)
	if err != nil {
		panic(err)
	}
	return ci
}

func Test_GetCityInfo(t *testing.T) {
	testcases := []struct {
		name     string
		request  *http.Request
		mockCity *city
		mockErr  error
		wantCode int
		wantResp *stickerio.CityInfo
	}{
		{
			name: "get city info works with a given cityID",
			request: httptest.NewRequest("GET", "/api", http.NoBody).WithContext(
				contextWithValues(context.Background(), map[ContextKey]string{
					CityIDKey: "foo",
				}),
			),
			mockCity: &city{
				id:   "foo",
				name: "foo",
			},
			wantCode: 200,
			wantResp: &stickerio.CityInfo{
				ID:   "foo",
				Name: "foo",
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			handler := NewServerHandler(&mockRepository{
				funcGetCityInfo: func(ctx context.Context, id string) (*city, error) {
					return testcase.mockCity, testcase.mockErr
				},
			})
			recorder := httptest.NewRecorder()
			http.HandlerFunc(handler.GetCityInfo).ServeHTTP(recorder, testcase.request)

			if recorder.Code != testcase.wantCode {
				t.Errorf("unexpected status code: %d, want: %d", recorder.Code, testcase.wantCode)
			}
			if diff := cmp.Diff(toCityInfo(recorder.Body.Bytes()), testcase.wantResp); diff != "" {
				t.Errorf("unexpected diff in response: %v", diff)
			}
		})
	}
}

func toCity(v []byte) *stickerio.City {
	ci := &stickerio.City{}
	err := json.Unmarshal(v, ci)
	if err != nil {
		panic(err)
	}
	return ci
}

func Test_GetCity(t *testing.T) {
	testcases := []struct {
		name     string
		request  *http.Request
		mockCity *city
		mockErr  error
		wantCode int
		wantResp *stickerio.City
	}{
		{
			name: "get city works with a given cityID and playerID",
			request: httptest.NewRequest("GET", "/api", http.NoBody).WithContext(
				contextWithValues(context.Background(), map[ContextKey]string{
					CityIDKey:   "foo",
					PlayerIDKey: "bar",
				}),
			),
			mockCity: &city{
				id:   "foo",
				name: "foo",
			},
			wantCode: 200,
			wantResp: &stickerio.City{
				CityInfo: &stickerio.CityInfo{
					ID:   "foo",
					Name: "foo",
				},
				CityBuildings: &stickerio.CityBuildings{},
				CityResources: &stickerio.CityResources{},
				UnitCount:     &stickerio.UnitCount{},
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.name, func(t *testing.T) {
			handler := NewServerHandler(&mockRepository{
				funcGetCity: func(ctx context.Context, id, playerID string) (*city, error) {
					return testcase.mockCity, testcase.mockErr
				},
			})
			recorder := httptest.NewRecorder()
			http.HandlerFunc(handler.GetCity).ServeHTTP(recorder, testcase.request)

			if recorder.Code != testcase.wantCode {
				t.Errorf("unexpected status code: %d, want: %d", recorder.Code, testcase.wantCode)
			}
			if diff := cmp.Diff(toCity(recorder.Body.Bytes()), testcase.wantResp); diff != "" {
				t.Errorf("unexpected diff in response: %v", diff)
			}
		})
	}
}
