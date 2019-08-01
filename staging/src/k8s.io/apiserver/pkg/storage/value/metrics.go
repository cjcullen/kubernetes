/*
Copyright 2019 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package value

import (
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "apiserver"
	subsystem = "storage"
)

var (
	transformerLatencies = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "transformation_latencies_microseconds",
			Help:      "Latencies in microseconds of value transformation operations.",
			// In-process transformations (ex. AES CBC) complete on the order of 20 microseconds. However, when
			// external KMS is involved latencies may climb into milliseconds.
			Buckets: prometheus.ExponentialBuckets(5, 2, 14),
		},
		[]string{"transformation_type"},
	)

	transformerOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "transformation_operations_total",
			Help:      "Total number of transformations.",
		},
		[]string{"transformation_type", "transformer_prefix", "status"},
	)

	deprecatedTransformerFailuresTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "transformation_failures_total",
			Help:      "(Deprecated) Total number of failed transformation operations.",
		},
		[]string{"transformation_type"},
	)

	envelopeTransformationCacheMissTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "envelope_transformation_cache_misses_total",
			Help:      "Total number of cache misses while accessing key decryption key(KEK).",
		},
	)

	dataKeyGenerationLatencies = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "data_key_generation_latencies_microseconds",
			Help:      "Latencies in microseconds of data encryption key(DEK) generation operations.",
			Buckets:   prometheus.ExponentialBuckets(5, 2, 14),
		},
	)
	dataKeyGenerationFailuresTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "data_key_generation_failures_total",
			Help:      "Total number of failed data encryption key(DEK) generation operations.",
		},
	)
)

var registerMetrics sync.Once

func RegisterMetrics() {
	registerMetrics.Do(func() {
		prometheus.MustRegister(transformerLatencies)
		prometheus.MustRegister(transformerOperationsTotal)
		prometheus.MustRegister(deprecatedTransformerFailuresTotal)
		prometheus.MustRegister(envelopeTransformationCacheMissTotal)
		prometheus.MustRegister(dataKeyGenerationLatencies)
		prometheus.MustRegister(dataKeyGenerationFailuresTotal)
	})
}

// RecordTransformation records latencies and count of TransformFromStorage and TransformToStorage operations.
// Note that transformation_failures_total metric is deprecated, use transformation_operations_total instead.
func RecordTransformation(transformationType, transformerPrefix string, start time.Time, err error) {
	transformerOperationsTotal.WithLabelValues(transformationType, transformerPrefix, Code(err).String()).Inc()

	switch {
	case err == nil:
		transformerLatencies.WithLabelValues(transformationType).Observe(sinceInMicroseconds(start))
	default:
		deprecatedTransformerFailuresTotal.WithLabelValues(transformationType).Inc()
	}

	since := sinceInMicroseconds(start)
	transformerLatencies.WithLabelValues(transformationType).Observe(float64(since))
}

// RecordCacheMiss records a miss on Key Encryption Key(KEK) - call to KMS was required to decrypt KEK.
func RecordCacheMiss() {
	envelopeTransformationCacheMissTotal.Inc()
}

// RecordDataKeyGeneration records latencies and count of Data Encryption Key generation operations.
func RecordDataKeyGeneration(start time.Time, err error) {
	if err != nil {
		dataKeyGenerationFailuresTotal.Inc()
		return
	}

	since := sinceInMicroseconds(start)
	dataKeyGenerationLatencies.Observe(float64(since))
}

func sinceInMicroseconds(start time.Time) float64 {
	return float64(time.Since(start).Nanoseconds() / time.Microsecond.Nanoseconds())
}

// Code returns the Code of the error if it is a Status error, codes.OK if err
// is nil, or codes.Unknown otherwise.
func Code(err error) codes.Code {
	// Don't use FromError to avoid allocation of OK status.
	if err == nil {
		return codes.OK
	}
	st, ok := status.FromError(err)
	if !ok {
		return codes.Unknown
	}
	return st.Code()
}
