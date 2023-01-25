package logic

import (
	"errors"
	"net/http"
	"testing"

	respModel "github.com/PereRohit/util/model"
	"github.com/PereRohit/util/testutil"
	"github.com/golang/mock/gomock"

	"github.com/vatsal278/TransactionManagementService/internal/model"
	"github.com/vatsal278/TransactionManagementService/internal/repo/datasource"
	"github.com/vatsal278/TransactionManagementService/pkg/mock"

)

func Test_transactionManagementServiceLogic_HealthCheck(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name  string
		setup func() datasource.DataSource
		want  bool
	}{
		{
			name: "Success",
			setup: func() datasource.DataSource {
				mockDs := mock.NewMockDataSource(mockCtrl)

				mockDs.EXPECT().HealthCheck().Times(1).
					Return(true)

				return mockDs
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := NewTransactionManagementServiceLogic(tt.setup())

			got := rec.HealthCheck()

			diff := testutil.Diff(got, tt.want)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})
	}
}

func Test_transactionManagementServiceLogic_Ping(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	tests := []struct {
		name  string
		setup func() datasource.DataSource
		give  *model.PingRequest
		want  *respModel.Response
	}{
		{
			name: "Success",
			setup: func() datasource.DataSource {
				mockDs := mock.NewMockDataSource(mockCtrl)

				mockDs.EXPECT().Ping(&model.PingDs{
					Data: "ping",
				}).Times(1).
					Return(&model.DsResponse{
						Data: "pong",
					}, nil)

				return mockDs
			},
			give: &model.PingRequest{
				Data: "ping",
			},
			want: &respModel.Response{
				Status:  http.StatusOK,
				Message: "Pong",
				Data: &model.DsResponse{
					Data: "pong",
				},
			},
		},
		{
			name: "Failure::datasource error",
			setup: func() datasource.DataSource {
				mockDs := mock.NewMockDataSource(mockCtrl)

				mockDs.EXPECT().Ping(&model.PingDs{
					Data: "ping",
				}).Times(1).
					Return(nil, errors.New("ds down"))

				return mockDs
			},
			give: &model.PingRequest{
				Data: "ping",
			},
			want: &respModel.Response{
				Status:  http.StatusInternalServerError,
				Message: "",
				Data:    nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := NewTransactionManagementServiceLogic(tt.setup())

			got := rec.Ping(tt.give)

			diff := testutil.Diff(got, tt.want)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})
	}
}
