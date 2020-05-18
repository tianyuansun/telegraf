package portname

import (
	"strings"
	"testing"
	"time"

	"github.com/influxdata/telegraf"
	"github.com/influxdata/telegraf/testutil"
	"github.com/stretchr/testify/require"
)

var fakeServices = `
http		80/tcp		www		# WorldWideWeb HTTP
https		443/tcp				# http protocol over TLS/SSL
tftp		69/udp`

func TestReadServicesFile(t *testing.T) {
	readServicesFile()
	require.NotZero(t, len(services))
}

func TestFakeServices(t *testing.T) {
	r := strings.NewReader(fakeServices)
	m := readServices(r)
	require.Equal(t, sMap{"tcp": {80: "http", 443: "https"}, "udp": {69: "tftp"}}, m)
}

func TestTable(t *testing.T) {
	var tests = []struct {
		name     string
		tag      string
		dest     string
		prot     string
		input    []telegraf.Metric
		expected []telegraf.Metric
	}{
		{
			name: "ordinary tcp default",
			tag:  "port",
			dest: "service",
			prot: "tcp",
			input: []telegraf.Metric{
				testutil.MustMetric(
					"meas",
					map[string]string{
						"port": "443",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"meas",
					map[string]string{
						"port":    "443",
						"service": "https",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "force udp default",
			tag:  "port",
			dest: "service",
			prot: "udp",
			input: []telegraf.Metric{
				testutil.MustMetric(
					"meas",
					map[string]string{
						"port": "69",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"meas",
					map[string]string{
						"port":    "69",
						"service": "tftp",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "override default protocol",
			tag:  "port",
			dest: "service",
			prot: "foobar",
			input: []telegraf.Metric{
				testutil.MustMetric(
					"meas",
					map[string]string{
						"port": "80/tcp",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"meas",
					map[string]string{
						"port":    "80/tcp",
						"service": "http",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "multiple metrics, multiple protocols",
			tag:  "port",
			dest: "service",
			prot: "tcp",
			input: []telegraf.Metric{
				testutil.MustMetric(
					"meas",
					map[string]string{
						"port": "80",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"meas",
					map[string]string{
						"port": "69/udp",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"meas",
					map[string]string{
						"port":    "80",
						"service": "http",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
				testutil.MustMetric(
					"meas",
					map[string]string{
						"port":    "69/udp",
						"service": "tftp",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "rename source and destination tags",
			tag:  "foo",
			dest: "bar",
			prot: "tcp",
			input: []telegraf.Metric{
				testutil.MustMetric(
					"meas",
					map[string]string{
						"foo": "80",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"meas",
					map[string]string{
						"foo": "80",
						"bar": "http",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
		},
		{
			name: "don't mix up protocols",
			tag:  "port",
			dest: "service",
			prot: "udp",
			input: []telegraf.Metric{
				testutil.MustMetric(
					"meas",
					map[string]string{
						"port": "80",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
			expected: []telegraf.Metric{
				testutil.MustMetric(
					"meas",
					map[string]string{
						"port": "80",
					},
					map[string]interface{}{},
					time.Unix(0, 0),
				),
			},
		},
	}

	r := strings.NewReader(fakeServices)
	services = readServices(r)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PortName{
				SourceTag:       tt.tag,
				DestTag:         tt.dest,
				DefaultProtocol: tt.prot,
			}

			actual := p.Apply(tt.input...)

			testutil.RequireMetricsEqual(t, tt.expected, actual)
		})
	}
}
