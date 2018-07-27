// Copyright 2018 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package clickhouse

import (
	"database/sql"
	"fmt"
	"net/url"
	"sort"
	"strings"
	"sync"
	"time"

	clickhouse_common "k8s.io/heapster/common/clickhouse"
	"k8s.io/heapster/metrics/core"

	"github.com/golang/glog"
	"github.com/kshvakov/clickhouse"
)

var insertSQL = `INSERT INTO %s.%s
(date, name, tags, val, ts)
VALUES	(?, ?, ?, ?, ?)`

var clickhouseBlacklistLabels = map[string]struct{}{
	core.LabelPodNamespaceUID.Key: {},
	core.LabelPodId.Key:           {},
	core.LabelHostname.Key:        {},
	core.LabelHostID.Key:          {},
}

type point struct {
	name string
	tags []string
	val  float64
	ts   time.Time
}

type clickhouseSink struct {
	sync.RWMutex
	client *sql.DB
	c      clickhouse_common.ClickhouseConfig

	// wg and conChan will work together to limit concurrent clickhouse sink goroutines.
	wg      sync.WaitGroup
	conChan chan struct{}
}

func (sink *clickhouseSink) Name() string {
	return "clickhouse"
}

func (tsdbSink *clickhouseSink) Stop() {
	// Do nothing
}

func (sink *clickhouseSink) ExportData(dataBatch *core.DataBatch) {
	sink.Lock()
	defer sink.Unlock()

	if err := sink.client.Ping(); err != nil {
		glog.Warningf("Failed to ping clickhouse: %v", err)
		return
	}

	dataPoints := make([]point, 0, 0)
	for _, metricSet := range dataBatch.MetricSets {
		for metricName, metricValue := range metricSet.MetricValues {
			var value float64
			if core.ValueInt64 == metricValue.ValueType {
				value = float64(metricValue.IntValue)
			} else if core.ValueFloat == metricValue.ValueType {
				value = float64(metricValue.FloatValue)
			} else {
				continue
			}

			pt := point{
				name: metricName,
				val:  value,
				ts:   dataBatch.Timestamp,
			}

			for key, value := range metricSet.Labels {
				if _, exists := clickhouseBlacklistLabels[key]; !exists {
					if value != "" {
						if key == "labels" {
							lbs := strings.Split(value, ",")
							for _, lb := range lbs {
								ts := strings.Split(lb, ":")
								if len(ts) == 2 && ts[0] != "" && ts[1] != "" {
									pt.tags = append(pt.tags, fmt.Sprintf("%s=%s", ts[0], ts[1]))
								}
							}
						} else {
							pt.tags = append(pt.tags, fmt.Sprintf("%s=%s", key, value))
						}

					}
				}
			}

			pt.tags = append(pt.tags, "cluster_name="+sink.c.ClusterName)

			dataPoints = append(dataPoints, pt)
			if len(dataPoints) >= sink.c.BatchSize {
				sink.concurrentSendData(dataPoints)
				dataPoints = make([]point, 0, 0)
			}
		}
	}

	if len(dataPoints) >= 0 {
		sink.concurrentSendData(dataPoints)
	}

	sink.wg.Wait()
}

func (sink *clickhouseSink) concurrentSendData(dataPoints []point) {
	sink.wg.Add(1)
	// use the channel to block until there's less than the maximum number of concurrent requests running
	sink.conChan <- struct{}{}
	go func(dataPoints []point) {
		sink.sendData(dataPoints)
	}(dataPoints)
}

func (sink *clickhouseSink) sendData(dataPoints []point) {
	defer func() {
		// empty an item from the channel so the next waiting request can run
		<-sink.conChan
		sink.wg.Done()
	}()

	sql := fmt.Sprintf(insertSQL, sink.c.Database, sink.c.Table)

	start := time.Now()
	// post them to db all at once
	tx, err := sink.client.Begin()
	if err != nil {
		glog.Errorf("begin transaction failure: %s", err.Error())
		return
	}

	// build statements
	smt, err := tx.Prepare(sql)
	if err != nil {
		glog.Errorf("prepare statement failure: %s", err.Error())
		return
	}
	for _, pts := range dataPoints {
		// ensure tags are inserted in the same order each time
		// possibly/probably impacts indexing?
		sort.Strings(pts.tags)
		_, err = smt.Exec(pts.ts, pts.name, clickhouse.Array(pts.tags),
			pts.val, pts.ts)

		if err != nil {
			glog.Errorf("statement exec failure: %s", err.Error())
			return
		}
	}

	// commit and record metrics
	if err = tx.Commit(); err != nil {
		glog.Errorf("commit failed failure: %s", err.Error())
		return
	}

	end := time.Now()
	glog.Infof("Exported %d data to clickhouse in %s", len(dataPoints), end.Sub(start))
}

func NewClickhouseSink(uri *url.URL) (core.DataSink, error) {
	config, err := clickhouse_common.BuildConfig(uri)
	if err != nil {
		return nil, err
	}

	client, err := sql.Open("clickhouse", config.DSN)
	if err != nil {
		glog.Errorf("connecting to clickhouse: %v", err)
		return nil, err
	}

	sink := &clickhouseSink{
		c:       *config,
		client:  client,
		conChan: make(chan struct{}, config.Concurrency),
	}

	glog.Infof("created clickhouse sink with options: host:%s user:%s db:%s", config.Host, config.UserName, config.Database)
	return sink, nil
}
