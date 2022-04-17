import React, { PureComponent } from 'react';
import { pick } from 'lodash';
import { QueryEditorProps, ExploreMode } from '@grafana/data';
import { Segment } from '@grafana/ui';
import { CloudWatchJsonData, CloudWatchQuery } from '../types';
import { CloudWatchDatasource } from '../datasource';
import { QueryInlineField } from './';
import { MetricsQueryEditor } from './MetricsQueryEditor';
import LogsQueryEditor from './LogsQueryEditor';

export type Props = QueryEditorProps<CloudWatchDatasource, CloudWatchQuery, CloudWatchJsonData>;

const apiModes = {
  Metrics: { label: 'CloudWatch Metrics', value: 'Metrics' },
  Logs: { label: 'CloudWatch Logs', value: 'Logs' },
};

export class PanelQueryEditor extends PureComponent<Props> {
  render() {
    const { query } = this.props;
    const apiMode = query.queryMode ?? 'Metrics';

    return (
      <>
        <QueryInlineField label="Query Mode">
          <Segment
            value={apiModes[apiMode]}
            options={Object.values(apiModes)}
            onChange={({ value }) => {
              const newMode = (value as 'Metrics' | 'Logs') ?? 'Metrics';
              if (newMode !== apiModes[apiMode].value) {
                const commonProps = pick(
                  query,
                  'id',
                  'region',
                  'namespace',
                  'refId',
                  'hide',
                  'key',
                  'queryType',
                  'datasource'
                );

                this.props.onChange({
                  ...commonProps,
                  queryMode: newMode,
                } as CloudWatchQuery);
              }
            }}
          />
        </QueryInlineField>
        {apiMode === ExploreMode.Logs ? (
          <LogsQueryEditor {...this.props} allowCustomValue />
        ) : (
          <MetricsQueryEditor {...this.props} />
        )}
      </>
    );
  }
}
