package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"Aicon-assignment/internal/domain/entity"
	domainErrors "Aicon-assignment/internal/domain/errors"
	"Aicon-assignment/internal/usecase"
)

// MockItemUsecase はtestify/mockを使用したモックユースケース
type MockItemUsecase struct {
	mock.Mock
}

func (m *MockItemUsecase) GetAllItems(ctx context.Context) ([]*entity.Item, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entity.Item), args.Error(1)
}

func (m *MockItemUsecase) GetItemByID(ctx context.Context, id int64) (*entity.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Item), args.Error(1)
}

func (m *MockItemUsecase) CreateItem(ctx context.Context, input usecase.CreateItemInput) (*entity.Item, error) {
	args := m.Called(ctx, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Item), args.Error(1)
}

func (m *MockItemUsecase) UpdateItem(ctx context.Context, id int64, input usecase.UpdateItemInput) (*entity.Item, error) {
	args := m.Called(ctx, id, input)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entity.Item), args.Error(1)
}

func (m *MockItemUsecase) DeleteItem(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockItemUsecase) GetCategorySummary(ctx context.Context) (*usecase.CategorySummary, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*usecase.CategorySummary), args.Error(1)
}

func TestItemHandler_UpdateItem(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		id             string
		requestBody    interface{}
		setupMock      func(*MockItemUsecase)
		expectedStatus int
	}{
		{
			name: "正常系: nameのみ更新",
			id:   "1",
			requestBody: map[string]interface{}{
				"name": "更新されたアイテム名",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				updatedItem, _ := entity.NewItem("更新されたアイテム名", "時計", "ROLEX", 1000000, "2023-01-01")
				updatedItem.ID = 1
				mockUsecase.On("UpdateItem", mock.Anything, int64(1), mock.AnythingOfType("usecase.UpdateItemInput")).Return(updatedItem, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "正常系: brandのみ更新",
			id:   "1",
			requestBody: map[string]interface{}{
				"brand": "更新されたブランド",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				updatedItem, _ := entity.NewItem("アイテム", "時計", "更新されたブランド", 1000000, "2023-01-01")
				updatedItem.ID = 1
				mockUsecase.On("UpdateItem", mock.Anything, int64(1), mock.AnythingOfType("usecase.UpdateItemInput")).Return(updatedItem, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "正常系: purchase_priceのみ更新",
			id:   "1",
			requestBody: map[string]interface{}{
				"purchase_price": 2000000,
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				updatedItem, _ := entity.NewItem("アイテム", "時計", "ROLEX", 2000000, "2023-01-01")
				updatedItem.ID = 1
				mockUsecase.On("UpdateItem", mock.Anything, int64(1), mock.AnythingOfType("usecase.UpdateItemInput")).Return(updatedItem, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "正常系: 複数フィールド更新",
			id:   "1",
			requestBody: map[string]interface{}{
				"name":           "更新されたアイテム名",
				"brand":          "更新されたブランド",
				"purchase_price": 2000000,
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				updatedItem, _ := entity.NewItem("更新されたアイテム名", "時計", "更新されたブランド", 2000000, "2023-01-01")
				updatedItem.ID = 1
				mockUsecase.On("UpdateItem", mock.Anything, int64(1), mock.AnythingOfType("usecase.UpdateItemInput")).Return(updatedItem, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "異常系: 無効なID",
			id:   "invalid",
			requestBody: map[string]interface{}{
				"name": "更新されたアイテム名",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				// UpdateItemは呼ばれない
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: 空のリクエストボディ",
			id:   "1",
			requestBody: map[string]interface{}{},
			setupMock: func(mockUsecase *MockItemUsecase) {
				// UpdateItemは呼ばれない
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: 存在しないアイテム",
			id:   "999",
			requestBody: map[string]interface{}{
				"name": "更新されたアイテム名",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				mockUsecase.On("UpdateItem", mock.Anything, int64(999), mock.AnythingOfType("usecase.UpdateItemInput")).Return((*entity.Item)(nil), domainErrors.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "異常系: バリデーションエラー（空のname）",
			id:   "1",
			requestBody: map[string]interface{}{
				"name": "",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				// UpdateItemは呼ばれない
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: バリデーションエラー（空のbrand）",
			id:   "1",
			requestBody: map[string]interface{}{
				"brand": "",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				// UpdateItemは呼ばれない
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: バリデーションエラー（負のpurchase_price）",
			id:   "1",
			requestBody: map[string]interface{}{
				"purchase_price": -1,
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				// UpdateItemは呼ばれない
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: 無効なJSON",
			id:   "1",
			requestBody: "invalid json",
			setupMock: func(mockUsecase *MockItemUsecase) {
				// UpdateItemは呼ばれない
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "異常系: サーバーエラー",
			id:   "1",
			requestBody: map[string]interface{}{
				"name": "更新されたアイテム名",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				mockUsecase.On("UpdateItem", mock.Anything, int64(1), mock.AnythingOfType("usecase.UpdateItemInput")).Return((*entity.Item)(nil), domainErrors.ErrDatabaseError)
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockItemUsecase)
			tt.setupMock(mockUsecase)
			handler := NewItemHandler(mockUsecase)

			var reqBody []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPatch, "/items/"+tt.id, bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/items/:id")
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			// ハンドラーを呼び出し、エラーは返されない（EchoのJSONレスポンスはエラーを返さない）
			err = handler.UpdateItem(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			mockUsecase.AssertExpectations(t)
		})
	}
}

func TestItemHandler_UpdateItem_HTTPResponse(t *testing.T) {
	e := echo.New()

	tests := []struct {
		name           string
		id             string
		requestBody    interface{}
		setupMock      func(*MockItemUsecase)
		expectedStatus int
	}{
		{
			name: "正常系: 成功レスポンス",
			id:   "1",
			requestBody: map[string]interface{}{
				"name": "更新されたアイテム名",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				updatedItem, _ := entity.NewItem("更新されたアイテム名", "時計", "ROLEX", 1000000, "2023-01-01")
				updatedItem.ID = 1
				mockUsecase.On("UpdateItem", mock.Anything, int64(1), mock.AnythingOfType("usecase.UpdateItemInput")).Return(updatedItem, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "異常系: 存在しないアイテム",
			id:   "999",
			requestBody: map[string]interface{}{
				"name": "更新されたアイテム名",
			},
			setupMock: func(mockUsecase *MockItemUsecase) {
				mockUsecase.On("UpdateItem", mock.Anything, int64(999), mock.AnythingOfType("usecase.UpdateItemInput")).Return((*entity.Item)(nil), domainErrors.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := new(MockItemUsecase)
			tt.setupMock(mockUsecase)
			handler := NewItemHandler(mockUsecase)

			reqBody, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPatch, "/items/"+tt.id, bytes.NewReader(reqBody))
			req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)
			c.SetPath("/items/:id")
			c.SetParamNames("id")
			c.SetParamValues(tt.id)

			err = handler.UpdateItem(c)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedStatus, rec.Code)

			mockUsecase.AssertExpectations(t)
		})
	}
}
