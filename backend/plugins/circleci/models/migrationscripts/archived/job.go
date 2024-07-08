/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package archived

import (
	"github.com/apache/incubator-devlake/core/models/migrationscripts/archived"
)

type CircleciJob struct {
	ConnectionId      uint64                `gorm:"primaryKey;type:BIGINT"`
	WorkflowId        string                `gorm:"primaryKey;type:varchar(100)" json:"workflow_id"`
	Id                string                `gorm:"primaryKey;type:varchar(100)" json:"id"`
	ProjectSlug       string                `gorm:"type:varchar(255)" json:"project_slug"`
	CanceledBy        string                `gorm:"type:varchar(100)" json:"canceled_by"`
	Dependencies      []string              `gorm:"serializer:json;type:text" json:"dependencies"`
	JobNumber         int64                 `json:"job_number"`
	StartedAt         *archived.Iso8601Time `json:"started_at"`
	Name              string                `gorm:"type:varchar(255)" json:"name"`
	ApprovedBy        string                `gorm:"type:varchar(100)" json:"approved_by"`
	Status            string                `gorm:"type:varchar(100)" json:"status"`
	Type              string                `gorm:"type:varchar(100)" json:"type"`
	ApprovalRequestId string                `gorm:"type:varchar(100)" json:"approval_request_id"`
	StoppedAt         *archived.Iso8601Time `json:"stopped_at"`
	DurationSec       uint64                `json:"duration_sec"`
	PipelineId        string                `gorm:"type:varchar(100)" json:"pipeline_id"`

	archived.NoPKModel
}

func (CircleciJob) TableName() string {
	return "_tool_circleci_jobs"
}
