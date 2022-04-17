import React, { Component } from 'react';

import { Select, Table } from '@grafana/ui';
import { DataFrame, FieldMatcherID, getFrameDisplayName, PanelProps, SelectableValue } from '@grafana/data';
import { PanelOptions } from './models.gen';
import { css } from '@emotion/css';
import { config } from 'app/core/config';
import { FilterItem, TableSortByFieldState } from '@grafana/ui/src/components/Table/types';
import { dispatch } from '../../../store/store';
import { applyFilterFromTable } from '../../../features/variables/adhoc/actions';
import { getDashboardSrv } from '../../../features/dashboard/services/DashboardSrv';

interface Props extends PanelProps<PanelOptions> {}

export class TablePanel extends Component<Props> {
  constructor(props: Props) {
    super(props);
  }

  onColumnResize = (fieldDisplayName: string, width: number) => {
    const { fieldConfig } = this.props;
    const { overrides } = fieldConfig;

    const matcherId = FieldMatcherID.byName;
    const propId = 'custom.width';

    // look for existing override
    const override = overrides.find((o) => o.matcher.id === matcherId && o.matcher.options === fieldDisplayName);

    if (override) {
      // look for existing property
      const property = override.properties.find((prop) => prop.id === propId);
      if (property) {
        property.value = width;
      } else {
        override.properties.push({ id: propId, value: width });
      }
    } else {
      overrides.push({
        matcher: { id: matcherId, options: fieldDisplayName },
        properties: [{ id: propId, value: width }],
      });
    }

    this.props.onFieldConfigChange({
      ...fieldConfig,
      overrides,
    });
  };

  onSortByChange = (sortBy: TableSortByFieldState[]) => {
    this.props.onOptionsChange({
      ...this.props.options,
      sortBy,
    });
  };

  onChangeTableSelection = (val: SelectableValue<number>) => {
    this.props.onOptionsChange({
      ...this.props.options,
      frameIndex: val.value || 0,
    });

    // Force a redraw -- but no need to re-query
    this.forceUpdate();
  };

  onCellFilterAdded = (filter: FilterItem) => {
    const { key, value, operator } = filter;
    const panelModel = getDashboardSrv().getCurrent()?.getPanelById(this.props.id);
    const datasource = panelModel?.datasource;

    if (!datasource) {
      return;
    }

    dispatch(applyFilterFromTable({ datasource, key, operator, value }));
  };

  renderTable(frame: DataFrame, width: number, height: number) {
    const { options } = this.props;

    return (
      <Table
        height={height}
        width={width}
        data={frame}
        noHeader={!options.showHeader}
        showTypeIcons={options.showTypeIcons}
        resizable={true}
        initialSortBy={options.sortBy}
        onSortByChange={this.onSortByChange}
        onColumnResize={this.onColumnResize}
        onCellFilterAdded={this.onCellFilterAdded}
      />
    );
  }

  getCurrentFrameIndex() {
    const { data, options } = this.props;
    const count = data.series?.length;
    return options.frameIndex > 0 && options.frameIndex < count ? options.frameIndex : 0;
  }

  render() {
    const { data, height, width } = this.props;

    const count = data.series?.length;
    const hasFields = data.series[0]?.fields.length;

    if (!count || !hasFields) {
      return <div className={tableStyles.noData}>No data</div>;
    }

    if (count > 1) {
      const inputHeight = config.theme.spacing.formInputHeight;
      const padding = 8 * 2;
      const currentIndex = this.getCurrentFrameIndex();
      const names = data.series.map((frame, index) => {
        return {
          label: getFrameDisplayName(frame),
          value: index,
        };
      });

      return (
        <div className={tableStyles.wrapper}>
          {this.renderTable(data.series[currentIndex], width, height - inputHeight - padding)}
          <div className={tableStyles.selectWrapper}>
            <Select
              menuShouldPortal
              options={names}
              value={names[currentIndex]}
              onChange={this.onChangeTableSelection}
            />
          </div>
        </div>
      );
    }

    return this.renderTable(data.series[0], width, height - 12);
  }
}

const tableStyles = {
  wrapper: css`
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    height: 100%;
  `,
  noData: css`
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
  `,
  selectWrapper: css`
    padding: 8px;
  `,
};
