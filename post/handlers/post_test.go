package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	commonmw "smapp/common/middleware"
	validation "smapp/common/validation"
	"smapp/post/handlers"
	"smapp/post/model"
	"smapp/post/service"
	"testing"

	"smapp/post/service/mocks"

	ozzo "github.com/go-ozzo/ozzo-validation/v4"

	"github.com/google/uuid"
	"github.com/matryer/is"
	"go.uber.org/mock/gomock"
)

type validateSuccess struct{}

func (validateSuccess) Validate(_ validation.Validatable) error {
	return nil
}

type validateFailure struct{}

func (validateFailure) Validate(_ validation.Validatable) error {
	return ozzo.Errors{"field1": errors.New("error1"), "field2": errors.New("error2"), "valid_field": nil}
}

func TestCreatePost(t *testing.T) {
	reqBody := handlers.CreatePostRequestBody{
		Body: "Post body!",
		Images: []model.ImageLocation{
			{Bucket: "bucket1", Key: "key1"},
			{Bucket: "bucket2", Key: "key2"},
		},
	}

	var userID uuid.UUID
	for i := 0; i < 16; i++ {
		userID[i] = byte(i)
	}

	var returnedPostID uuid.UUID
	for i := 15; i >= 0; i-- {
		returnedPostID[i] = byte(i)
	}

	checkRespSuccess := func(is *is.I, body map[string]interface{}) {
		is.Equal(body["status"], "success")
		is.Equal(body["id"], returnedPostID.String())
	}

	getMockReturnsError := func(err error) func(*gomock.Controller) *mocks.MockPost {
		return func(ctrl *gomock.Controller) *mocks.MockPost {
			m := mocks.NewMockPost(ctrl)
			m.
				EXPECT().
				Create(gomock.Any(), reqBody.Body, userID, gomock.Len(len(reqBody.Images))).
				Return(uuid.Nil, err)
			return m
		}
	}

	tests := []struct {
		name        string
		reqBody     handlers.CreatePostRequestBody
		code        int
		getPostMock func(*gomock.Controller) *mocks.MockPost
		validator   validation.Validator
		checkResp   func(*is.I, map[string]interface{})
	}{
		{
			name:      "creates a new post",
			reqBody:   reqBody,
			code:      http.StatusCreated,
			validator: validateSuccess{},
			getPostMock: func(ctrl *gomock.Controller) *mocks.MockPost {
				m := mocks.NewMockPost(ctrl)
				m.
					EXPECT().
					Create(gomock.Any(), reqBody.Body, userID, gomock.Len(len(reqBody.Images))).
					Return(returnedPostID, nil)
				return m
			},
			checkResp: checkRespSuccess,
		},
		{
			name:      "fails on validation failure",
			reqBody:   reqBody,
			code:      http.StatusBadRequest,
			validator: validateFailure{},
			getPostMock: func(ctrl *gomock.Controller) *mocks.MockPost {
				m := mocks.NewMockPost(ctrl)
				m.
					EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				return m
			},
			checkResp: func(is *is.I, body map[string]interface{}) {
				is.Equal(body["status"], "error")
				is.Equal(body["message"], "Validation failed")
				is.True(
					reflect.DeepEqual(
						body["errors"],
						interface{}(map[string]interface{}{"field1": "error1", "field2": "error2"}),
					),
				)
			},
		},
		{
			name:        "fails on invalid image",
			reqBody:     reqBody,
			code:        http.StatusBadRequest,
			validator:   validateSuccess{},
			getPostMock: getMockReturnsError(fmt.Errorf("%w", service.ErrInvalidImage)),
			checkResp: func(is *is.I, body map[string]interface{}) {
				is.Equal(body["status"], "error")
				is.Equal(body["message"], "One or more provided image locations are invalid or inaccessible")
			},
		},
		{
			name:        "fails on timeout",
			reqBody:     reqBody,
			code:        http.StatusRequestTimeout,
			validator:   validateSuccess{},
			getPostMock: getMockReturnsError(fmt.Errorf("%w", context.DeadlineExceeded)),
			checkResp: func(is *is.I, body map[string]interface{}) {
				is.Equal(body["status"], "error")
				is.Equal(body["message"], "Request Timeout")
			},
		},
		{
			name:        "fails on unknown error",
			reqBody:     reqBody,
			code:        http.StatusInternalServerError,
			validator:   validateSuccess{},
			getPostMock: getMockReturnsError(errors.New("unknown error")),
			checkResp: func(is *is.I, body map[string]interface{}) {
				is.Equal(body["status"], "error")
				is.Equal(body["message"], "Internal Server Error")
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is := is.New(t)
			ctrl := gomock.NewController(t)
			handler := commonmw.ParseUserID(handlers.CreatePost(test.validator, test.getPostMock(ctrl)))
			body, err := json.Marshal(test.reqBody)
			is.NoErr(err)
			req := httptest.NewRequest(http.MethodPost, "/posts", bytes.NewReader(body))
			req.Header.Set("X-User-Id", userID.String())
			resp := httptest.NewRecorder()
			handler.ServeHTTP(resp, req)

			is.Equal(resp.Code, test.code)
			var respBody map[string]interface{}
			err = json.NewDecoder(resp.Body).Decode(&respBody)
			is.NoErr(err)
			test.checkResp(is, respBody)
		})
	}
}
