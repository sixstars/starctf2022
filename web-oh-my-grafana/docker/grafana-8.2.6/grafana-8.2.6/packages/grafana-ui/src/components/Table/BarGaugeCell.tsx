import React, { FC } from 'react';
import { ThresholdsConfig, ThresholdsMode, VizOrientation, getFieldConfigWithMinMax } from '@grafana/data';
import { BarGauge, BarGaugeDisplayMode } from '../BarGauge/BarGauge';
import { TableCellProps, TableCellDisplayMode } from './types';

const defaultScale: ThresholdsConfig = {
  mode: ThresholdsMode.Absolute,
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
};

export const BarGaugeCell: FC<TableCellProps> = (props) => {
  const { field, innerWidth, tableStyles, cell, cellProps } = props;

  let config = getFieldConfigWithMinMax(field, false);
  if (!config.thresholds) {
    config = {
      ...config,
      thresholds: defaultScale,
    };
  }

  const displayValue = field.display!(cell.value);
  let barGaugeMode = BarGaugeDisplayMode.Gradient;

  if (field.config.custom && field.config.custom.displayMode === TableCellDisplayMode.LcdGauge) {
    barGaugeMode = BarGaugeDisplayMode.Lcd;
  } else if (field.config.custom && field.config.custom.displayMode === TableCellDisplayMode.BasicGauge) {
    barGaugeMode = BarGaugeDisplayMode.Basic;
  }

  return (
    <div {...cellProps} className={tableStyles.cellContainer}>
      <BarGauge
        width={innerWidth}
        height={tableStyles.cellHeightInner}
        field={config}
        display={field.display}
        text={{ valueSize: 14 }}
        value={displayValue}
        orientation={VizOrientation.Horizontal}
        theme={tableStyles.theme}
        itemSpacing={1}
        lcdCellWidth={8}
        displayMode={barGaugeMode}
      />
    </div>
  );
};
