package e2e

import (
	"bytes"
	"path"
	"strings"
	"testing"

	"github.com/chancez/yamlforge/cmd"
	"github.com/stretchr/testify/require"
)

func TestExamples(t *testing.T) {
	trim := func(s string) string {
		return strings.TrimLeft(s, "\n")
	}
	tests := []struct {
		file     string
		expected string
	}{
		{
			file: "file.yfg.yaml",
			expected: trim(`
server {
    listen 8080;
    root /data/up1;

    location / {
    }
}
`),
		},
		{
			file: "exec.yfg.yaml",
			expected: trim(`
server {
    listen 443 ssl;
    root /data/up1;

    location / {
    }
}
`),
		},
		{
			file: "jq.yfg.yaml",
			expected: trim(`
metricsPort: 9999
`),
		},
		{
			file: "cel.yfg.yaml",
			expected: trim(`
MyApp`),
		},
		{
			file: "jsonpatch.yfg.yaml",
			expected: trim(`
apiVersion: v1
kind: Service
metadata:
    annotations:
        service.beta.kubernetes.io/aws-load-balancer-type: nlb
    name: my-service
spec:
    ports:
        - name: grpc
          port: 80
          protocol: TCP
          targetPort: 9376
        - name: metrics
          port: 9999
          protocol: TCP
          targetPort: 9999
    selector:
        app.kubernetes.io/name: MyApp
`),
		},
		{
			file: "jsonmergepatch.yfg.yaml",
			expected: trim(`
apiVersion: v1
kind: Service
metadata:
    annotations:
        service.beta.kubernetes.io/aws-load-balancer-type: nlb
    name: my-service
spec:
    ports:
        - name: grpc
          port: 80
          protocol: TCP
          targetPort: 9376
        - name: metrics
          port: 9999
          protocol: TCP
          targetPort: 9999
    selector:
        app.kubernetes.io/name: MyApp
`),
		},
		{
			file: "merge.yfg.yaml",
			expected: trim(`
apiVersion: v1
kind: Service
metadata:
    name: my-service
    namespace: example
spec:
    ports:
        - name: grpc
          port: 80
          protocol: TCP
          targetPort: 9376
        - name: metrics
          port: 9999
          protocol: TCP
          targetPort: 9999
    selector:
        app.kubernetes.io/name: MyApp
`),
		},
		{
			file: "template.yfg.yaml",
			expected: trim(`
app:
  version: 'v1.2.3'
  environment: 'production'
`),
		},
		{
			file: "template-literal.yfg.yaml",
			expected: trim(`
server {
    listen 8080;
    root /var/www/data;

    location / {
    }
}
`),
		},
		{
			file: "single-generator.yfg.yaml",
			expected: trim(`
app:
  version: 'v2.0.0'
  environment: 'dev'
`),
		},
		{
			file: "advanced/reusable-transformer.yfg.yaml",
			expected: trim(`
apiVersion: v1
kind: Service
metadata:
    annotations:
        service.beta.kubernetes.io/aws-load-balancer-type: nlb
    name: my-service
spec:
    ports:
        - name: grpc
          port: 80
          protocol: TCP
          targetPort: 9376
        - name: metrics
          port: 9999
          protocol: TCP
          targetPort: 9999
    selector:
        app.kubernetes.io/name: MyApp
`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			p := path.Join("../examples", tt.file)
			require.FileExists(t, p, "example file must exist")
			var buf bytes.Buffer
			c := cmd.RootCmd
			c.SetArgs([]string{"generate", p})
			c.SetOut(&buf)
			err := c.Execute()
			require.NoError(t, err, "yfg generate should succeed on examples")
			require.Equal(t, tt.expected, buf.String(), "example output should match expected")
		})
	}
}
