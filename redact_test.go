package queryplan

import "testing"

func Test_redactSQL(t *testing.T) {
	tests := []struct {
		name      string
		statement string
		want      string
		wantErr   bool
	}{
		{
			name:      "simple",
			statement: "SELECT * FROM users WHERE id = 1",
			want:      "select * from users where id = :redacted1",
			wantErr:   false,
		},
		{
			name:      "simple with comment",
			statement: "SELECT * FROM users WHERE id = 1 /* comment */",
			want:      "select * from users where id = :redacted1 /* comment */",
			wantErr:   false,
		},
		{
			name:      "simple with comment and newline",
			statement: "SELECT * FROM users WHERE id = 1 /* comment */\n",
			want:      "select * from users where id = :redacted1 /* comment */",
			wantErr:   false,
		},
		{
			name: "simple with subquery",
			statement: `SELECT * FROM users WHERE id IN (
				SELECT id FROM users WHERE id = 1
			)`,
			want:    "select * from users where id in (select id from users where id = :redacted1)",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := redactSQL(tt.statement)
			if (err != nil) != tt.wantErr {
				t.Errorf("redactSQL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("redactSQL() got: %v, want: %v", got, tt.want)
			}
		})
	}
}
