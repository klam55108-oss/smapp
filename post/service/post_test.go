package service_test

import (
	"context"
	"errors"
	"smapp/post/model"
	"smapp/post/service"
	"testing"

	imagemocks "smapp/common/grpc/image/mocks"
	repomocks "smapp/post/repository/mocks"

	"github.com/google/uuid"
	"github.com/matryer/is"
	"go.uber.org/mock/gomock"
)

func TestDefaultPostCreate(t *testing.T) {
	body := "Post body!"

	var authorID uuid.UUID
	for i := 0; i < 16; i++ {
		authorID[i] = byte(i)
	}

	validImage1 := model.ImageLocation{Bucket: "bucket1", Key: "images/post/something"}
	validImage2 := model.ImageLocation{Bucket: "bucket2", Key: "images/post/something"}
	invalidImage := model.ImageLocation{Bucket: "bucket", Key: "images/profile/something"}

	var returnedPostID uuid.UUID
	for i := 15; i >= 0; i-- {
		returnedPostID[i] = byte(i)
	}

	getPostMockReturnsError := func(err error) func(ctrl *gomock.Controller) *repomocks.MockPost {
		return func(ctrl *gomock.Controller) *repomocks.MockPost {
			m := repomocks.NewMockPost(ctrl)
			m.EXPECT().
				Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
				Return(uuid.Nil, err)
			return m
		}
	}

	checkResultError := func(targetErr error) func(*is.I, uuid.UUID, error) {
		return func(is *is.I, id uuid.UUID, err error) {
			is.Equal(id, uuid.Nil)
			is.True(errors.Is(err, targetErr))
		}
	}

	unknownError := errors.New("unknown error")

	tests := []struct {
		name         string
		images       []model.ImageLocation
		getPostMock  func(*gomock.Controller) *repomocks.MockPost
		getImageMock func(*gomock.Controller) *imagemocks.MockImageClient
		checkResult  func(*is.I, uuid.UUID, error)
	}{
		{
			name:   "returns a UUID and no error",
			images: []model.ImageLocation{validImage1, validImage2},
			getPostMock: func(ctrl *gomock.Controller) *repomocks.MockPost {
				m := repomocks.NewMockPost(ctrl)
				m.EXPECT().
					Create(gomock.Any(), body, authorID, gomock.Len(2)).
					Return(returnedPostID, nil)
				return m
			},
			getImageMock: func(ctrl *gomock.Controller) *imagemocks.MockImageClient {
				m := imagemocks.NewMockImageClient(ctrl)
				m.EXPECT().
					CheckObjectExists(gomock.Any(), gomock.Any()).
					Return(nil, nil).
					Times(2)
				return m
			},
			checkResult: func(is *is.I, id uuid.UUID, err error) {
				is.Equal(id, returnedPostID)
				is.NoErr(err)
			},
		},
		{
			name:   "returns an error when image prefix is not 'images/post'",
			images: []model.ImageLocation{validImage1, invalidImage, validImage2},
			getPostMock: func(ctrl *gomock.Controller) *repomocks.MockPost {
				m := repomocks.NewMockPost(ctrl)
				m.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				return m
			},
			getImageMock: func(ctrl *gomock.Controller) *imagemocks.MockImageClient {
				m := imagemocks.NewMockImageClient(ctrl)
				m.EXPECT().
					CheckObjectExists(gomock.Any(), gomock.Any()).
					Return(nil, nil).
					AnyTimes()
				return m
			},
			checkResult: checkResultError(service.ErrInvalidImage),
		},
		{
			name:   "returns an error when image is not accessible",
			images: []model.ImageLocation{validImage1, validImage2},
			getPostMock: func(ctrl *gomock.Controller) *repomocks.MockPost {
				m := repomocks.NewMockPost(ctrl)
				m.EXPECT().
					Create(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
				return m
			},
			getImageMock: func(ctrl *gomock.Controller) *imagemocks.MockImageClient {
				m := imagemocks.NewMockImageClient(ctrl)
				gomock.InOrder(
					m.EXPECT().
						CheckObjectExists(gomock.Any(), gomock.Any()).
						Return(nil, nil),
					m.EXPECT().
						CheckObjectExists(gomock.Any(), gomock.Any()).
						Return(nil, errors.New("error while checking image existence")),
				)
				return m
			},
			checkResult: checkResultError(service.ErrInvalidImage),
		},
		{
			name:        "returns an context.DeadlineExceeded when repository returns context.DeadlineExceeded",
			images:      []model.ImageLocation{validImage1, validImage2},
			getPostMock: getPostMockReturnsError(context.DeadlineExceeded),
			getImageMock: func(ctrl *gomock.Controller) *imagemocks.MockImageClient {
				m := imagemocks.NewMockImageClient(ctrl)
				m.EXPECT().
					CheckObjectExists(gomock.Any(), gomock.Any()).
					Return(nil, nil).
					Times(2)
				return m
			},
			checkResult: checkResultError(context.DeadlineExceeded),
		},
		{
			name:        "returns the error that the repository returns if it is unknown",
			images:      []model.ImageLocation{validImage1, validImage2},
			getPostMock: getPostMockReturnsError(unknownError),
			getImageMock: func(ctrl *gomock.Controller) *imagemocks.MockImageClient {
				m := imagemocks.NewMockImageClient(ctrl)
				m.EXPECT().
					CheckObjectExists(gomock.Any(), gomock.Any()).
					Return(nil, nil).
					Times(2)
				return m
			},
			checkResult: checkResultError(unknownError),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			is := is.New(t)
			ctrl := gomock.NewController(t)
			post := service.NewDefaultPost(test.getPostMock(ctrl), nil, nil, nil, test.getImageMock(ctrl))
			id, err := post.Create(context.TODO(), body, authorID, test.images)
			test.checkResult(is, id, err)
		})
	}
}
