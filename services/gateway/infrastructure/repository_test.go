package infrastructure

import (
	"context"
	"github.com/CoucouMonEcho/go-framework/micro/registry"
	"github.com/CoucouMonEcho/go-framework/micro/registry/memery"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestServiceDiscovery_resolveTarget(t *testing.T) {
	type fields struct {
		registry registry.Registry
	}
	type args struct {
		target string
	}

	r := memery.NewRegistry()
	_ = r.Register(context.Background(), registry.ServiceInstance{
		Name:    "user-service",
		Address: "localhost:8000",
	})

	tests := []struct {
		name   string
		fields fields
		args   args
		want   string
		want1  bool
	}{
		{
			name: "user-service",
			args: args{
				target: "service://user-service",
			},
			fields: fields{
				registry: r,
			},
			want: "localhost:8000",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &ServiceDiscovery{
				registry: tt.fields.registry,
			}
			addr, _ := s.resolveTarget(tt.args.target)
			require.Equal(t, tt.want, addr)
		})
	}
}
