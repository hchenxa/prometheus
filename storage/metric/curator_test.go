// Copyright 2013 Prometheus Team
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package metric

import (
	"code.google.com/p/goprotobuf/proto"
	"github.com/prometheus/prometheus/coding"
	"github.com/prometheus/prometheus/coding/indexable"
	"github.com/prometheus/prometheus/model"
	dto "github.com/prometheus/prometheus/model/generated"
	fixture "github.com/prometheus/prometheus/storage/raw/leveldb/test"
	"testing"
	"time"
)

type (
	curationState struct {
		fingerprint string
		groupSize   int
		olderThan   time.Duration
		lastCurated time.Time
	}

	watermarkState struct {
		fingerprint  string
		lastAppended time.Time
	}

	sample struct {
		time  time.Time
		value float32
	}

	sampleGroup struct {
		fingerprint string
		values      []sample
	}

	context struct {
		curationStates  fixture.Pairs
		watermarkStates fixture.Pairs
		sampleGroups    fixture.Pairs
	}
)

func (c curationState) Get() (key, value coding.Encoder) {
	key = coding.NewProtocolBufferEncoder(&dto.CurationKey{
		Fingerprint:      model.NewFingerprintFromRowKey(c.fingerprint).ToDTO(),
		MinimumGroupSize: proto.Uint32(uint32(c.groupSize)),
		OlderThan:        proto.Int64(int64(c.olderThan)),
	})

	value = coding.NewProtocolBufferEncoder(&dto.CurationValue{
		LastCompletionTimestamp: proto.Int64(c.lastCurated.Unix()),
	})

	return
}

func (w watermarkState) Get() (key, value coding.Encoder) {
	key = coding.NewProtocolBufferEncoder(model.NewFingerprintFromRowKey(w.fingerprint).ToDTO())
	value = coding.NewProtocolBufferEncoder(model.NewWatermarkFromTime(w.lastAppended).ToMetricHighWatermarkDTO())
	return
}

func (s sampleGroup) Get() (key, value coding.Encoder) {
	key = coding.NewProtocolBufferEncoder(&dto.SampleKey{
		Fingerprint:   model.NewFingerprintFromRowKey(s.fingerprint).ToDTO(),
		Timestamp:     indexable.EncodeTime(s.values[0].time),
		LastTimestamp: proto.Int64(s.values[len(s.values)-1].time.Unix()),
		SampleCount:   proto.Uint32(uint32(len(s.values))),
	})

	series := &dto.SampleValueSeries{}

	for _, value := range s.values {
		series.Value = append(series.Value, &dto.SampleValueSeries_Value{
			Timestamp: proto.Int64(value.time.Unix()),
			Value:     proto.Float32(float32(value.value)),
		})
	}

	value = coding.NewProtocolBufferEncoder(series)

	return
}

func TestCurator(t *testing.T) {
	var (
		scenarios = []struct {
			context context
		}{
			{
				context: context{
					curationStates: fixture.Pairs{
						curationState{
							fingerprint: "0001-A-1-Z",
							groupSize:   5,
							olderThan:   1 * time.Hour,
							lastCurated: testInstant.Add(-1 * 30 * time.Minute),
						},
						curationState{
							fingerprint: "0002-A-2-Z",
							groupSize:   5,
							olderThan:   1 * time.Hour,
							lastCurated: testInstant.Add(-1 * 90 * time.Minute),
						},
					},
					watermarkStates: fixture.Pairs{
						watermarkState{
							fingerprint:  "0001-A-1-Z",
							lastAppended: testInstant.Add(-1 * 15 * time.Minute),
						},
						watermarkState{
							fingerprint:  "0002-A-1-Z",
							lastAppended: testInstant.Add(-1 * 15 * time.Minute),
						},
					},
					sampleGroups: fixture.Pairs{
						sampleGroup{
							fingerprint: "0001-A-1-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 90 * time.Minute),
									value: 0,
								},
								{
									time:  testInstant.Add(-1 * 85 * time.Minute),
									value: 1,
								},
								{
									time:  testInstant.Add(-1 * 80 * time.Minute),
									value: 2,
								},
								{
									time:  testInstant.Add(-1 * 75 * time.Minute),
									value: 3,
								},
								{
									time:  testInstant.Add(-1 * 70 * time.Minute),
									value: 4,
								},
							},
						},
						sampleGroup{
							fingerprint: "0001-A-1-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 65 * time.Minute),
									value: 0,
								},
								{
									time:  testInstant.Add(-1 * 60 * time.Minute),
									value: 1,
								},
								{
									time:  testInstant.Add(-1 * 55 * time.Minute),
									value: 2,
								},
								{
									time:  testInstant.Add(-1 * 50 * time.Minute),
									value: 3,
								},
								{
									time:  testInstant.Add(-1 * 45 * time.Minute),
									value: 4,
								},
							},
						},
						sampleGroup{
							fingerprint: "0001-A-1-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 40 * time.Minute),
									value: 0,
								},
								{
									time:  testInstant.Add(-1 * 35 * time.Minute),
									value: 1,
								},
								{
									time:  testInstant.Add(-1 * 30 * time.Minute),
									value: 2,
								},
							},
						},
						sampleGroup{
							fingerprint: "0001-A-1-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 25 * time.Minute),
									value: 0,
								},
							},
						},
						sampleGroup{
							fingerprint: "0001-A-1-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 35 * time.Minute),
									value: 1,
								},
							},
						},
						sampleGroup{
							fingerprint: "0001-A-1-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 30 * time.Minute),
									value: 2,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 90 * time.Minute),
									value: 0,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 89 * time.Minute),
									value: 1,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 88 * time.Minute),
									value: 2,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 87 * time.Minute),
									value: 3,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 86 * time.Minute),
									value: 4,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 85 * time.Minute),
									value: 5,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 84 * time.Minute),
									value: 6,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 83 * time.Minute),
									value: 7,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 82 * time.Minute),
									value: 8,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 81 * time.Minute),
									value: 9,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 80 * time.Minute),
									value: 10,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 79 * time.Minute),
									value: 11,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 78 * time.Minute),
									value: 12,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 77 * time.Minute),
									value: 13,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 76 * time.Minute),
									value: 14,
								},
								{
									time:  testInstant.Add(-1 * 75 * time.Minute),
									value: 15,
								},
								{
									time:  testInstant.Add(-1 * 74 * time.Minute),
									value: 16,
								},
								{
									time:  testInstant.Add(-1 * 73 * time.Minute),
									value: 17,
								},
								{
									time:  testInstant.Add(-1 * 72 * time.Minute),
									value: 18,
								},
								{
									time:  testInstant.Add(-1 * 71 * time.Minute),
									value: 19,
								},
								{
									time:  testInstant.Add(-1 * 70 * time.Minute),
									value: 20,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 69 * time.Minute),
									value: 21,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 68 * time.Minute),
									value: 22,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 67 * time.Minute),
									value: 23,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 66 * time.Minute),
									value: 24,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 65 * time.Minute),
									value: 25,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 64 * time.Minute),
									value: 26,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 63 * time.Minute),
									value: 27,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 62 * time.Minute),
									value: 28,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 61 * time.Minute),
									value: 29,
								},
							},
						},
						sampleGroup{
							fingerprint: "0002-A-2-Z",
							values: []sample{
								{
									time:  testInstant.Add(-1 * 60 * time.Minute),
									value: 30,
								},
							},
						},
					},
				},
			},
		}
	)

	for _, scenario := range scenarios {
		var (
			curatorDirectory   = fixture.NewPreparer(t).Prepare("curator", fixture.NewCassetteFactory(scenario.context.curationStates))
			watermarkDirectory = fixture.NewPreparer(t).Prepare("watermark", fixture.NewCassetteFactory(scenario.context.watermarkStates))
			sampleDirectory    = fixture.NewPreparer(t).Prepare("sample", fixture.NewCassetteFactory(scenario.context.sampleGroups))
		)
		defer curatorDirectory.Close()
		defer watermarkDirectory.Close()
		defer sampleDirectory.Close()

	}
}