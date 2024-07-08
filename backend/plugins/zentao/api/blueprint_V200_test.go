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
	"database/sql"
	"github.com/apache/incubator-devlake/core/dal"
	"gorm.io/gorm/migrator"
	"testing"

	coreModels "github.com/apache/incubator-devlake/core/models"

	"github.com/apache/incubator-devlake/core/models/common"
	"github.com/apache/incubator-devlake/core/models/domainlayer"
	"github.com/apache/incubator-devlake/core/models/domainlayer/ticket"
	"github.com/apache/incubator-devlake/core/plugin"
	helper "github.com/apache/incubator-devlake/helpers/pluginhelper/api"
	"github.com/apache/incubator-devlake/helpers/unithelper"
	mockdal "github.com/apache/incubator-devlake/mocks/core/dal"
	mockplugin "github.com/apache/incubator-devlake/mocks/core/plugin"
	"github.com/apache/incubator-devlake/plugins/zentao/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMakeDataSourcePipelinePlanV200(t *testing.T) {
	connection := &models.ZentaoConnection{
		BaseConnection: helper.BaseConnection{
			Name: "zentao-test",
			Model: common.Model{
				ID: 1,
			},
		},
		ZentaoConn: models.ZentaoConn{
			RestConnection: helper.RestConnection{
				Endpoint:         "https://zentao.example.org/api.php/v1/",
				Proxy:            "",
				RateLimitPerHour: 0,
			},
			BasicAuth: helper.BasicAuth{
				Username: "Username",
				Password: "Password",
			},
		},
	}
	mockMeta := mockplugin.NewPluginMeta(t)
	mockMeta.On("RootPkgPath").Return("github.com/apache/incubator-devlake/plugins/zentao")
	mockMeta.On("Name").Return("zentao").Maybe()
	err := plugin.RegisterPlugin("zentao", mockMeta)
	assert.Nil(t, err)
	// Refresh Global Variables and set the sql mock
	mockBasicRes(t)

	bs := &coreModels.BlueprintScope{
		ScopeId: "1",
	}
	/*bs2 := &coreModels.BlueprintScope{
		Id: "product/1",
	}*/
	bpScopes := make([]*coreModels.BlueprintScope, 0)
	bpScopes = append(bpScopes, bs)

	plan := make(coreModels.PipelinePlan, len(bpScopes))
	plan, scopes, err := makePipelinePlanV200(nil, plan, bpScopes, connection)
	assert.Nil(t, err)

	expectPlan := coreModels.PipelinePlan{
		coreModels.PipelineStage{
			{
				Plugin:   "zentao",
				Subtasks: []string{},
				Options: map[string]interface{}{
					"ConnectionId": uint64(1),
					"projectId":    int64(1),
					"timeAfter":    "",
				},
			},
		},
		/*coreModels.PipelineStage{
			{
				Plugin:   "zentao",
				Subtasks: []string{},
				Options: map[string]interface{}{
					"ConnectionId": uint64(1),
					"productId":    int64(1),
					"projectId":    int64(0),
				},
			},
		},*/
	}
	assert.Equal(t, expectPlan, plan)
	expectScopes := make([]plugin.Scope, 0)
	scopeTicket1 := &ticket.Board{
		DomainEntity: domainlayer.DomainEntity{
			Id: "zentao:ZentaoProject:1:1",
		},
		Name:        "test/testRepo",
		Description: "",
		Url:         "",
		CreatedDate: nil,
		Type:        `project`,
	}
	/*scopeTicket2 := &ticket.Board{
		DomainEntity: domainlayer.DomainEntity{
			Id: "zentao:ZentaoProduct:1:1",
		},
		Name:        "test/testRepo",
		Description: "",
		Url:         "",
		CreatedDate: nil,
		Type:        `product/normal`,
	}*/

	expectScopes = append(expectScopes, scopeTicket1)
	assert.Equal(t, expectScopes, scopes)
}

// mockBasicRes FIXME ...
func mockBasicRes(t *testing.T) {
	/*testZentaoProduct := &models.ZentaoProduct{
		ConnectionId:  1,
		Id:            1,
		Name:          "test/testRepo",
		Type:          `product/normal`,
		ScopeConfigId: 0,
	}*/
	testZentaoProject := &models.ZentaoProject{
		Scope: common.Scope{
			ConnectionId:  1,
			ScopeConfigId: 0,
		},
		Id:   1,
		Name: "test/testRepo",
		Type: `project`,
	}
	var testColumTypes = []dal.ColumnMeta{
		migrator.ColumnType{
			NameValue: sql.NullString{
				String: "abc",
				Valid:  true,
			},
		},
	}

	mockRes := unithelper.DummyBasicRes(func(mockDal *mockdal.Dal) {
		mockDal.On("First", mock.AnythingOfType("*models.ZentaoProject"), mock.Anything).Run(func(args mock.Arguments) {
			dst := args.Get(0).(*models.ZentaoProject)
			*dst = *testZentaoProject
		}).Return(nil)

		/*mockDal.On("First", mock.AnythingOfType("*models.ZentaoProduct"), mock.Anything).Run(func(args mock.Arguments) {
			dst := args.Get(0).(*models.ZentaoProduct)
			*dst = *testZentaoProduct
		}).Return(nil)*/

		mockDal.On("First", mock.AnythingOfType("*models.ZentaoScopeConfig"), mock.Anything).Run(func(args mock.Arguments) {
			panic("The empty scope should not call First() for ZentaoScopeConfig")
		}).Return(nil)
		mockDal.On("GetColumns", mock.AnythingOfType("models.ZentaoConnection"), mock.Anything).Run(nil).Return(
			testColumTypes, nil)
		mockDal.On("GetColumns", mock.AnythingOfType("models.ZentaoProject"), mock.Anything).Run(nil).Return(
			testColumTypes, nil)
		mockDal.On("GetColumns", mock.AnythingOfType("models.ZentaoScopeConfig"), mock.Anything).Run(nil).Return(
			testColumTypes, nil)
	})
	p := mockplugin.NewPluginMeta(t)
	p.On("Name").Return("dummy").Maybe()
	Init(mockRes, p)
}
