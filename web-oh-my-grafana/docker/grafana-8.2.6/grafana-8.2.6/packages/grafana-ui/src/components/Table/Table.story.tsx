import React from 'react';
import { merge } from 'lodash';
import { Table } from '@grafana/ui';
import { withCenteredStory } from '../../utils/storybook/withCenteredStory';
import { Meta, Story } from '@storybook/react';
import { useTheme2 } from '../../themes';
import mdx from './Table.mdx';
import {
  DataFrame,
  FieldType,
  GrafanaTheme2,
  MutableDataFrame,
  ThresholdsConfig,
  ThresholdsMode,
  FieldConfig,
  formattedValueToString,
} from '@grafana/data';
import { prepDataForStorybook } from '../../utils/storybook/data';
import { FooterItem } from './types';

export default {
  title: 'Visualizations/Table',
  component: Table,
  decorators: [withCenteredStory],
  parameters: {
    controls: {
      exclude: ['onColumnResize', 'onSortByChange', 'onCellFilterAdded', 'ariaLabel', 'data', 'initialSortBy'],
    },
    docs: {
      page: mdx,
    },
  },
  args: {
    width: 700,
    height: 500,
    columnMinWidth: 150,
  },
} as Meta;

function buildData(theme: GrafanaTheme2, config: Record<string, FieldConfig>): DataFrame {
  const data = new MutableDataFrame({
    fields: [
      { name: 'Time', type: FieldType.time, values: [] }, // The time field
      {
        name: 'Quantity',
        type: FieldType.number,
        values: [],
        config: {
          decimals: 0,
          custom: {
            align: 'center',
            width: 80,
          },
        },
      },
      { name: 'Status', type: FieldType.string, values: [] }, // The time field
      {
        name: 'Value',
        type: FieldType.number,
        values: [],
        config: {
          decimals: 2,
        },
      },
      {
        name: 'Progress',
        type: FieldType.number,
        values: [],
        config: {
          unit: 'percent',
          min: 0,
          max: 100,
          custom: {
            width: 150,
          },
        },
      },
    ],
  });

  for (const field of data.fields) {
    field.config = merge(field.config, config[field.name]);
  }

  for (let i = 0; i < 1000; i++) {
    data.appendRow([
      new Date().getTime(),
      Math.random() * 2,
      Math.random() > 0.7 ? 'Active' : 'Cancelled',
      Math.random() * 100,
      Math.random() * 100,
    ]);
  }

  return prepDataForStorybook([data], theme)[0];
}

function buildFooterData(data: DataFrame): FooterItem[] {
  const values = data.fields[3].values.toArray();
  const valueSum = values.reduce((prev, curr) => {
    return prev + curr;
  }, 0);

  const valueField = data.fields[3];
  const displayValue = valueField.display ? valueField.display(valueSum) : valueSum;
  const val = valueField.display ? formattedValueToString(displayValue) : displayValue;

  const sum = { sum: val };
  const min = { min: String(5.2) };
  const valCell = [sum, min];

  return ['Totals', '10', undefined, valCell, '100%'];
}

const defaultThresholds: ThresholdsConfig = {
  steps: [
    {
      color: 'blue',
      value: -Infinity,
    },
    {
      color: 'green',
      value: 20,
    },
  ],
  mode: ThresholdsMode.Absolute,
};

export const Basic: Story = (args) => {
  const theme = useTheme2();
  const data = buildData(theme, {});

  return (
    <div className="panel-container" style={{ width: 'auto' }}>
      <Table data={data} height={args.height} width={args.width} {...args} />
    </div>
  );
};

export const BarGaugeCell: Story = (args) => {
  const theme = useTheme2();
  const data = buildData(theme, {
    Progress: {
      custom: {
        width: 200,
        displayMode: 'gradient-gauge',
      },
      thresholds: defaultThresholds,
    },
  });

  return (
    <div className="panel-container" style={{ width: 'auto' }}>
      <Table data={data} height={args.height} width={args.width} {...args} />
    </div>
  );
};

export const ColoredCells: Story = (args) => {
  const theme = useTheme2();
  const data = buildData(theme, {
    Progress: {
      custom: {
        width: 80,
        displayMode: 'color-background',
      },
      thresholds: defaultThresholds,
    },
  });

  return (
    <div className="panel-container" style={{ width: 'auto' }}>
      <Table data={data} height={args.height} width={args.width} {...args} />
    </div>
  );
};

export const Footer: Story = (args) => {
  const theme = useTheme2();
  const data = buildData(theme, {});
  const footer = buildFooterData(data);

  return (
    <div className="panel-container" style={{ width: 'auto', height: 'unset' }}>
      <Table data={data} height={args.height} width={args.width} footerValues={footer} {...args} />
    </div>
  );
};
