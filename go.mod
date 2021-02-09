module github.com/MauveSoftware/cni-health-probe

go 1.14

require (
	github.com/alecthomas/units v0.0.0-20210208195552-ff826a37aa15 // indirect
	github.com/go-ping/ping v0.0.0-20210207230027-ab39f29b51f8
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.6.0
	go.opentelemetry.io/otel v0.16.0
	go.opentelemetry.io/otel/exporters/metric/prometheus v0.16.0
	go.opentelemetry.io/otel/sdk v0.16.0
	golang.org/x/net v0.0.0-20201209123823-ac852fbbde11 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/go-playground/assert.v1 v1.2.1
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/apimachinery v0.16.7
	k8s.io/client-go v0.16.7
)
