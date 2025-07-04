package main

import (
	"testing"
)

func TestGenerateNginxConfig(t *testing.T) {
	tests := []struct {
		name     string
		config   ServicesConfig
		expected string
	}{
		{
			name: "single service",
			config: ServicesConfig{
				Services: []Service{
					{Name: "backend", Port: "8080"},
				},
			},
			expected: `server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;

    server_name _;

    location / {
      proxy_pass http://backend:8080;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    server_tokens off;
}
`,
		},
		{
			name: "multiple services with path-based routing",
			config: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
					{Name: "web", Port: "3000"},
					{Name: "admin", Port: "9000"},
				},
			},
			expected: `server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;

    server_name _;

    location /api/ {
      proxy_pass http://api:8080/;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    location /web/ {
      proxy_pass http://web:3000/;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    location /admin/ {
      proxy_pass http://admin:9000/;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    location / {
      proxy_pass http://api:8080;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }

    server_tokens off;
}
`,
		},
		{
			name:     "no services",
			config:   ServicesConfig{Services: []Service{}},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateNginxConfig(tt.config)
			if result != tt.expected {
				t.Errorf("generateNginxConfig() = %q, want %q", result, tt.expected)
			}
		})
	}
}