package session

import (
	"fmt"
	"golang.org/x/net/context"
	"reflect"
	"testing"
)

func TestSetSession(t *testing.T) {
	tests := []struct {
		name      string
		validator func(context.Context)
	}{
		{
			name: "SUCCESS:: Set Session",
			validator: func(ctx context.Context) {
				if !reflect.DeepEqual(ctx.Value(session{}), "123") {
					t.Errorf("Want: %v, Got: %v", "123", ctx.Value(session{}))
				}
			},
		},
	}

	// to execute the tests in the table
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := SetSession(context.Background(), "123")
			tt.validator(ctx)
		})
	}
}
func TestGetSession(t *testing.T) {
	tests := []struct {
		name      string
		validator func(any)
	}{
		{
			name: "SUCCESS:: Get Session",
			validator: func(val any) {
				if !reflect.DeepEqual(fmt.Sprint(val), "123") {
					t.Errorf("Want: %v, Got: %v", "123", val)
				}
			},
		},
	}

	// to execute the tests in the table
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), session{}, "123")
			val := GetSession(ctx)
			tt.validator(val)
		})
	}
}
