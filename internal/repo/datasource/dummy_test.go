package datasource

import (
	"testing"

	"github.com/PereRohit/util/testutil"

	"github.com/vatsal278/TransactionManagementService/internal/config"
	"github.com/vatsal278/TransactionManagementService/internal/model"
)

func Test_dummyDs_HealthCheck(t *testing.T) {
	type fields struct {
		dummySvc *config.DummyInternalSvc
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "Success",
			fields: fields{
				dummySvc: &config.DummyInternalSvc{
					Data: "hello world",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dummyDs{
				dummySvc: tt.fields.dummySvc,
			}

			got := d.HealthCheck()

			diff := testutil.Diff(got, tt.want)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})
	}
}

func Test_dummyDs_Ping(t *testing.T) {
	type fields struct {
		dummySvc *config.DummyInternalSvc
	}
	type args struct {
		req *model.PingDs
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *model.DsResponse
		wantErr error
	}{
		{
			name: "Success",
			fields: fields{
				dummySvc: &config.DummyInternalSvc{
					Data: "hello world",
				},
			},
			args: args{
				req: &model.PingDs{
					Data: "ping",
				},
			},
			want: &model.DsResponse{
				Data: "ping",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := dummyDs{
				dummySvc: tt.fields.dummySvc,
			}
			got, err := d.Ping(tt.args.req)

			diff := testutil.Diff(got, tt.want)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}

			diff = testutil.Diff(err, tt.wantErr)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})
	}
}

func TestNewDummyDs(t *testing.T) {
	type args struct {
		dummySvc *config.DummyInternalSvc
	}
	tests := []struct {
		name string
		args args
		want DataSource
	}{
		{
			name: "Success",
			args: args{
				dummySvc: &config.DummyInternalSvc{
					Data: "hello world",
				},
			},
			want: &dummyDs{
				dummySvc: &config.DummyInternalSvc{
					Data: "hello world",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDummyDs(tt.args.dummySvc)

			diff := testutil.Diff(got, tt.want)
			if diff != "" {
				t.Error(testutil.Callers(), diff)
			}
		})
	}
}
