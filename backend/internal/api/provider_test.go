package api

import (
	"testing"
)

func TestValidateJSONSchema(t *testing.T) {
	tests := []struct {
		name    string
		schema  string
		wantErr bool
	}{
		{
			name:    "Empty Schema",
			schema:  "",
			wantErr: true,
		},
		{
			name:    "Invalid JSON",
			schema:  "{invalid json}",
			wantErr: true,
		},
		{
			name: "Missing Type",
			schema: `{
				"properties": {
					"name": {"type": "string"}
				}
			}`,
			wantErr: true,
		},
		{
			name: "Valid Schema",
			schema: `{
				"type": "object",
				"properties": {
					"name": {"type": "string"},
					"age": {"type": "integer"}
				}
			}`,
			wantErr: false,
		},
		{
			name: "Invalid Property Format",
			schema: `{
				"type": "object",
				"properties": {
					"name": "invalid"
				}
			}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateJSONSchema(tt.schema)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateJSONSchema() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
} 