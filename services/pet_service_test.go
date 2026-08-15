package services

import (
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestIsForeignKeyViolation(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{name: "postgres FK violation", err: &pgconn.PgError{Code: "23503"}, want: true},
		{name: "postgres duplicate key", err: &pgconn.PgError{Code: "23505"}, want: false},
		{name: "generic error", err: errors.New("boom"), want: false},
		{name: "nil", err: nil, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isForeignKeyViolation(tt.err); got != tt.want {
				t.Errorf("isForeignKeyViolation() = %v, want %v", got, tt.want)
			}
		})
	}
}
