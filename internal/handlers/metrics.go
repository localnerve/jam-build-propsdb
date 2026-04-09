// metrics.go
//
// A scalable, high performance drop-in replacement for the jam-build nodejs data service
// Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC
//
// This file is part of jam-build-propsdb.
// jam-build-propsdb is free software: you can redistribute it and/or modify it
// under the terms of the GNU Affero General Public License as published by the Free Software
// Foundation, either version 3 of the License, or (at your option) any later version.
// jam-build-propsdb is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY;
// without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
// See the GNU Affero General Public License for more details.
// You should have received a copy of the GNU Affero General Public License along with jam-build-propsdb.
// If not, see <https://www.gnu.org/licenses/>.
// Additional terms under GNU AGPL version 3 section 7:
// a) The reasonable legal notice of original copyright and author attribution must be preserved
//    by including the string: "Copyright (c) 2026 Alex Grant <info@localnerve.com> (https://www.localnerve.com), LocalNerve LLC"
//    in this material, copies, or source code of derived works.

package handlers

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricDefinition defines the structure for a client metric
type MetricDefinition struct {
	Counter *prometheus.CounterVec
	Labels  []string
}

// Record extracts expected labels from the dynamic map to build the exact prometheus.Labels struct
func (m *MetricDefinition) Record(labels map[string]interface{}) {
	promLabels := prometheus.Labels{}
	for _, l := range m.Labels {
		if val, exists := labels[l]; exists {
			promLabels[l] = fmt.Sprintf("%v", val)
		} else {
			promLabels[l] = "" // Prometheus requires all registered labels to be present
		}
	}
	m.Counter.With(promLabels).Inc()
}

// metricsRegistry holds all the known client events mapping to their prometheus configurations.
// Extensibility: To add a new client event, simply add a new key/value entry here.
var metricsRegistry = map[string]*MetricDefinition{
	"version_conflict_backoff": {
		Labels: []string{"retryCount", "appVersion", "apiVersion", "schemaVersion"},
		Counter: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "propsdb_client_version_conflict_backoff_total",
				Help: "Total number of version conflict backoffs occurring on the client",
			},
			[]string{"retryCount", "appVersion", "apiVersion", "schemaVersion"},
		),
	},
	// Example of adding a future metric:
	// "auth_failure": {
	// 	Labels: []string{"reason"},
	// 	Counter: promauto.NewCounterVec(
	// 		prometheus.CounterOpts{
	// 			Name: "propsdb_client_auth_failures_total",
	// 			Help: "Total number of authentication failures on the client",
	// 		},
	// 		[]string{"reason"},
	// 	),
	// },
}

type metricsPayload struct {
	Event  string                 `json:"event"`
	Labels map[string]interface{} `json:"labels"`
}

// SetMetrics handles POST requests to log and increment a prometheus counter
// @Summary Set client metrics
// @Description Log client metrics events and increments prometheus counter
// @Tags Metrics
// @Accept json
// @Produce json
// @Param body body metricsPayload true "Metrics to record"
// @Success 204
// @Failure 400 {object} map[string]string
// @Router /metrics [post]
func SetMetrics(c *fiber.Ctx) error {
	var payload metricsPayload
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid metrics payload",
		})
	}

	if payload.Event != "" {
		log.Printf("Metrics event received: event=%s labels=%v", payload.Event, payload.Labels)

		if metricDef, exists := metricsRegistry[payload.Event]; exists {
			metricDef.Record(payload.Labels)
		} else {
			log.Printf("Warning: received unknown metrics event: %s", payload.Event)
		}
	}

	return c.SendStatus(fiber.StatusNoContent)
}
