package main

import (
	"testing"
)

// Test nginx config generation for path-based routing
func TestGenerateNginxConfigPathRouting(t *testing.T) {
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
				RoutingMode: RoutingModePath,
			},
			expected: `upstream backend_upstream {
    server backend:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    resolver 127.0.0.11 valid=30s;

    server_name _;

    location / {
      proxy_pass http://backend_upstream;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
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
				},
				RoutingMode: RoutingModePath,
			},
			expected: `upstream api_upstream {
    server api:8080 max_fails=3 fail_timeout=30s;
}

upstream web_upstream {
    server web:3000 max_fails=3 fail_timeout=30s;
}

server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    resolver 127.0.0.11 valid=30s;

    server_name _;

    location /api/ {
      proxy_pass http://api_upstream/;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    location /web/ {
      proxy_pass http://web_upstream/;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    location / {
      proxy_pass http://api_upstream;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    server_tokens off;
}
`,
		},
		{
			name: "no services",
			config: ServicesConfig{
				Services:    []Service{},
				RoutingMode: RoutingModePath,
			},
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

// Test nginx config generation for subdomain-based routing
func TestGenerateNginxConfigSubdomainRouting(t *testing.T) {
	tests := []struct {
		name     string
		config   ServicesConfig
		expected string
	}{
		{
			name: "subdomain routing single service",
			config: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
				},
				RoutingMode: RoutingModeSubdomain,
				Domain:      "example.com",
			},
			expected: `upstream api_upstream {
    server api:8080 max_fails=3 fail_timeout=30s;
}

server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    resolver 127.0.0.11 valid=30s;

    server_name api.example.com;

    location / {
      proxy_pass http://api_upstream;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    server_tokens off;
}

server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    resolver 127.0.0.11 valid=30s;

    server_name example.com;

    location / {
      proxy_pass http://api_upstream;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    server_tokens off;
}
`,
		},
		{
			name: "subdomain routing multiple services",
			config: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
					{Name: "web", Port: "3000"},
					{Name: "admin", Port: "9000"},
				},
				RoutingMode: RoutingModeSubdomain,
				Domain:      "example.com",
			},
			expected: `upstream api_upstream {
    server api:8080 max_fails=3 fail_timeout=30s;
}

upstream web_upstream {
    server web:3000 max_fails=3 fail_timeout=30s;
}

upstream admin_upstream {
    server admin:9000 max_fails=3 fail_timeout=30s;
}

server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    resolver 127.0.0.11 valid=30s;

    server_name api.example.com;

    location / {
      proxy_pass http://api_upstream;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    server_tokens off;
}

server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    resolver 127.0.0.11 valid=30s;

    server_name web.example.com;

    location / {
      proxy_pass http://web_upstream;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    server_tokens off;
}

server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    resolver 127.0.0.11 valid=30s;

    server_name admin.example.com;

    location / {
      proxy_pass http://admin_upstream;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    server_tokens off;
}

server {
    listen 443 ssl;
    http2 on;
    include /etc/nginx/conf.d/ssl.conf;
    resolver 127.0.0.11 valid=30s;

    server_name example.com;

    location / {
      proxy_pass http://api_upstream;
      proxy_set_header Host $host;
      proxy_set_header X-Real-IP $remote_addr;
      proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
      proxy_next_upstream error timeout invalid_header http_500 http_502 http_503;
    }

    server_tokens off;
}
`,
		},
		{
			name: "subdomain routing no services",
			config: ServicesConfig{
				Services:    []Service{},
				RoutingMode: RoutingModeSubdomain,
				Domain:      "example.com",
			},
			expected: "",
		},
		{
			name: "subdomain routing empty domain",
			config: ServicesConfig{
				Services: []Service{
					{Name: "api", Port: "8080"},
				},
				RoutingMode: RoutingModeSubdomain,
				Domain:      "",
			},
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