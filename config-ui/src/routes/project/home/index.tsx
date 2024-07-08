/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

import { useState, useMemo } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { PlusOutlined, SettingOutlined } from '@ant-design/icons';
import { Flex, Table, Button, Modal, Input, Checkbox, message } from 'antd';
import dayjs from 'dayjs';

import API from '@/api';
import { PageHeader, Block, ExternalLink } from '@/components';
import { getCron, cronPresets, PATHS } from '@/config';
import { ConnectionName } from '@/features';
import { useRefreshData } from '@/hooks';
import { DOC_URL } from '@/release';
import { formatTime, operator } from '@/utils';
import { PipelineStatus } from '@/routes/pipeline';
import { IBlueprint, IBPMode } from '@/types';

import { validName } from '../utils';

import * as S from './styled';

export const ProjectHomePage = () => {
  const [version, setVersion] = useState(1);
  const [page, setPage] = useState(1);
  const [pageSize] = useState(20);
  const [open, setOpen] = useState(false);
  const [name, setName] = useState('');
  const [enableDora, setEnableDora] = useState(true);
  const [saving, setSaving] = useState(false);

  const { ready, data } = useRefreshData(() => API.project.list({ page, pageSize }), [version, page, pageSize]);

  const navigate = useNavigate();

  const presets = useMemo(() => cronPresets.map((preset) => preset.config), []);
  const [dataSource, total] = useMemo(
    () => [
      (data?.projects ?? []).map((it) => {
        return {
          name: it.name,
          connections: it.blueprint?.connections,
          isManual: it.blueprint?.isManual,
          cronConfig: it.blueprint?.cronConfig,
          createdAt: it.createdAt,
          lastRunCompletedAt: it.lastPipeline?.finishedAt,
          lastRunStatus: it.lastPipeline?.status,
        };
      }),
      data?.count ?? 0,
    ],
    [data],
  );

  const handleShowDialog = () => setOpen(true);
  const handleHideDialog = () => {
    setOpen(false);
    setName('');
    setEnableDora(true);
  };

  const handleCreate = async () => {
    if (!validName(name)) {
      message.error('Please enter alphanumeric or underscore');
      return;
    }

    const [success] = await operator(
      async () => {
        await API.project.create({
          name,
          description: '',
          metrics: [
            {
              pluginName: 'dora',
              pluginOption: '',
              enable: enableDora,
            },
          ],
        });
        return API.blueprint.create({
          name: `${name}-Blueprint`,
          projectName: name,
          mode: IBPMode.NORMAL,
          enable: true,
          cronConfig: presets[0],
          isManual: false,
          skipOnFail: true,
          timeAfter: formatTime(dayjs().subtract(6, 'month').startOf('day').toDate(), 'YYYY-MM-DD[T]HH:mm:ssZ'),
          connections: [],
        });
      },
      {
        setOperating: setSaving,
      },
    );

    if (success) {
      handleHideDialog();
      setVersion((v) => v + 1);
    }
  };

  return (
    <PageHeader breadcrumbs={[{ name: 'Projects', path: PATHS.PROJECTS() }]}>
      <Flex style={{ marginBottom: 16 }} justify="flex-end">
        <Button type="primary" icon={<PlusOutlined />} onClick={handleShowDialog}>
          New Project
        </Button>
      </Flex>
      <Table
        rowKey="name"
        size="middle"
        loading={!ready}
        columns={[
          {
            title: 'Project Name',
            dataIndex: 'name',
            key: 'name',
            render: (name: string) => (
              <Link to={PATHS.PROJECT(name, 'configuration')} style={{ color: '#292b3f' }}>
                {name}
              </Link>
            ),
          },
          {
            title: 'Data Connections',
            dataIndex: 'connections',
            key: 'connections',
            render: (val: IBlueprint['connections']) =>
              !val || !val.length ? (
                'N/A'
              ) : (
                <ul>
                  {val.map((it) => (
                    <li key={`${it.pluginName}-${it.connectionId}`}>
                      <ConnectionName plugin={it.pluginName} connectionId={it.connectionId} />
                    </li>
                  ))}
                </ul>
              ),
          },
          {
            title: 'Sync Frequency',
            key: 'frequency',
            render: (_, { isManual, cronConfig }) => {
              const cron = getCron(isManual, cronConfig);
              return cron.label;
            },
          },
          {
            title: 'Created at',
            dataIndex: 'createdAt',
            key: 'createdAt',
            render: (val) => formatTime(val),
          },
          {
            title: 'Last Run Completed at',
            dataIndex: 'lastRunCompletedAt',
            key: 'lastRunCompletedAt',
            render: (val) => (val ? formatTime(val) : '-'),
          },
          {
            title: 'Last Run Status',
            dataIndex: 'lastRunStatus',
            key: 'lastRunStatus',
            render: (val) => (val ? <PipelineStatus status={val} /> : '-'),
          },
          {
            title: '',
            dataIndex: 'name',
            key: 'action',
            width: 100,
            align: 'center',
            render: (name: any) => (
              <Button
                type="primary"
                icon={<SettingOutlined />}
                onClick={() => navigate(PATHS.PROJECT(name, 'configuration'))}
              />
            ),
          },
        ]}
        dataSource={dataSource}
        pagination={{
          current: page,
          pageSize,
          total,
          onChange: setPage,
        }}
      />
      <Modal
        open={open}
        width={820}
        centered
        title="Create a New Project"
        okText="Save"
        okButtonProps={{
          disabled: !name,
          loading: saving,
        }}
        onOk={handleCreate}
        onCancel={handleHideDialog}
      >
        <S.DialogWrapper>
          <Block
            title="Project Name"
            description="Give your project a unique name with letters, numbers, -, _ or /"
            required
          >
            <Input
              style={{ width: 386 }}
              placeholder="Your Project Name"
              value={name}
              onChange={(e) => setName(e.target.value)}
            />
          </Block>
          <Block
            title="Project Settings"
            description={
              <>
                <ExternalLink link={DOC_URL.DORA}>DORA metrics</ExternalLink>
                <span style={{ marginLeft: 4 }}>
                  are four widely-adopted metrics for measuring software delivery performance.
                </span>
              </>
            }
          >
            <Checkbox checked={enableDora} onChange={(e) => setEnableDora(e.target.checked)}>
              Enable DORA Metrics
            </Checkbox>
          </Block>
        </S.DialogWrapper>
      </Modal>
    </PageHeader>
  );
};
