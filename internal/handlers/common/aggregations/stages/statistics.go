// Copyright 2021 FerretDB Inc.
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

package stages

// Statistic represents a statistic that can be fetched from the DB.
type Statistic int32

// List of statistics that can be fetched from the DB.
const (
	StatisticCount Statistic = iota
	StatisticLatency
	StatisticQueryExec
	StatisticStorage
)

// GetStatistics has the same idea as GetPushdownQuery: it returns a list of statistics that need
// to be fetched from the DB, because they are needed for one or more stages.
func GetStatistics(stages []Stage) map[Statistic]struct{} {
	stats := make(map[Statistic]struct{}, len(stages))

	for _, stage := range stages {
		switch st := stage.(type) {
		case *collStatsStage:
			if st.count {
				stats[StatisticCount] = struct{}{}
			}

			if st.latencyStats {
				stats[StatisticLatency] = struct{}{}
			}

			if st.queryExecStats {
				stats[StatisticQueryExec] = struct{}{}
			}

			if st.storageStats != nil {
				stats[StatisticStorage] = struct{}{}
			}
		}
	}

	return stats
}
