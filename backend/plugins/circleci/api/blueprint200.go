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

package api

import (
	"fmt"

	"github.com/apache/incubator-devlake/core/models/domainlayer/devops"
	"github.com/apache/incubator-devlake/plugins/circleci/models"

	"github.com/apache/incubator-devlake/core/dal"
	"github.com/apache/incubator-devlake/core/errors"
	coreModels "github.com/apache/incubator-devlake/core/models"
	"github.com/apache/incubator-devlake/core/models/domainlayer"
	"github.com/apache/incubator-devlake/core/models/domainlayer/didgen"
	"github.com/apache/incubator-devlake/core/plugin"
	"github.com/apache/incubator-devlake/core/utils"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
)

func MakeDataSourcePipelinePlanV200(
	subtaskMetas []plugin.SubTaskMeta,
	connectionId uint64,
	bpScopes []*coreModels.BlueprintScope,
) (coreModels.PipelinePlan, []plugin.Scope, errors.Error) {
	plan := make(coreModels.PipelinePlan, len(bpScopes))
	plan, err := makeDataSourcePipelinePlanV200(subtaskMetas, plan, bpScopes, connectionId)
	if err != nil {
		return nil, nil, err
	}
	scopes, err := makeScopesV200(bpScopes, connectionId)
	if err != nil {
		return nil, nil, err
	}

	return plan, scopes, nil
}

func makeDataSourcePipelinePlanV200(
	subtaskMetas []plugin.SubTaskMeta,
	plan coreModels.PipelinePlan,
	bpScopes []*coreModels.BlueprintScope,
	connectionId uint64,
) (coreModels.PipelinePlan, errors.Error) {
	for i, bpScope := range bpScopes {
		stage := plan[i]
		if stage == nil {
			stage = coreModels.PipelineStage{}
		}
		// construct task options for circleci
		options := make(map[string]interface{})
		project, err := scopeHelper.DbHelper().GetScope(connectionId, bpScope.ScopeId)
		if err != nil {
			return nil, err
		}
		options["projectSlug"] = project.Slug
		options["connectionId"] = connectionId

		// get scope config from db
		_, scopeConfig, err := scopeHelper.DbHelper().GetScopeAndConfig(connectionId, bpScope.ScopeId)
		if err != nil {
			return nil, err
		}

		subtasks, err := helper.MakePipelinePlanSubtasks(subtaskMetas, scopeConfig.Entities)
		if err != nil {
			return nil, err
		}
		stage = append(stage, &coreModels.PipelineTask{
			Plugin:   "circleci",
			Subtasks: subtasks,
			Options:  options,
		})
		plan[i] = stage
	}

	return plan, nil
}

func makeScopesV200(bpScopes []*coreModels.BlueprintScope, connectionId uint64) ([]plugin.Scope, errors.Error) {
	scopes := make([]plugin.Scope, 0)
	for _, bpScope := range bpScopes {
		project := &models.CircleciProject{}
		// get project from db
		err := basicRes.GetDal().First(project,
			dal.Where(`connection_id = ? and id = ?`,
				connectionId, bpScope.ScopeId))
		if err != nil {
			return nil, errors.Default.Wrap(err, fmt.Sprintf("fail to find project %s", bpScope.ScopeId))
		}
		// get scope config from db
		_, scopeConfig, err := scopeHelper.DbHelper().GetScopeAndConfig(connectionId, bpScope.ScopeId)
		if err != nil {
			return nil, err
		}

		// add board to scopes
		if utils.StringsContains(scopeConfig.Entities, plugin.DOMAIN_TYPE_CICD) {
			domainBoard := &devops.CicdScope{
				DomainEntity: domainlayer.DomainEntity{
					Id: didgen.NewDomainIdGenerator(&models.CircleciProject{}).Generate(project.ConnectionId, project.Id),
				},
				Name: project.Name,
			}
			scopes = append(scopes, domainBoard)
		}
	}
	return scopes, nil
}
