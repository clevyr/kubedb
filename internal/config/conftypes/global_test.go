package conftypes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGlobal_PasswordArg(t *testing.T) {
	tests := []struct {
		name string
		conf Global
		want string
	}{
		{"literal", Global{Password: "p a$s"}, `PGPASSWORD='p a$s'`},
		{"ref", Global{Password: "p a$s", PasswordRef: `"${KUBEDB_PASSWORD}"`}, `PGPASSWORD="${KUBEDB_PASSWORD}"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.conf.PasswordArg("PGPASSWORD").Quote())
		})
	}
}
