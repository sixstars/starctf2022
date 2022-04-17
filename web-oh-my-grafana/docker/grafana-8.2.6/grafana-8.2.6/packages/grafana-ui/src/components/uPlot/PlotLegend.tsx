import React from 'react';
import {
  DataFrame,
  DisplayValue,
  fieldReducers,
  getFieldDisplayName,
  getFieldSeriesColor,
  reduceField,
} from '@grafana/data';
import { UPlotConfigBuilder } from './config/UPlotConfigBuilder';
import { VizLegendItem } from '../VizLegend/types';
import { VizLegendOptions, AxisPlacement } from '@grafana/schema';
import { VizLayout, VizLayoutLegendProps } from '../VizLayout/VizLayout';
import { VizLegend } from '../VizLegend/VizLegend';
import { useTheme2 } from '../../themes';

const defaultFormatter = (v: any) => (v == null ? '-' : v.toFixed(1));

interface PlotLegendProps extends VizLegendOptions, Omit<VizLayoutLegendProps, 'children'> {
  data: DataFrame[];
  config: UPlotConfigBuilder;
}

export const PlotLegend: React.FC<PlotLegendProps> = ({
  data,
  config,
  placement,
  calcs,
  displayMode,
  ...vizLayoutLegendProps
}) => {
  const theme = useTheme2();
  const legendItems = config
    .getSeries()
    .map<VizLegendItem | undefined>((s) => {
      const seriesConfig = s.props;
      const fieldIndex = seriesConfig.dataFrameFieldIndex;
      const axisPlacement = config.getAxisPlacement(s.props.scaleKey);

      if (!fieldIndex) {
        return undefined;
      }

      const field = data[fieldIndex.frameIndex]?.fields[fieldIndex.fieldIndex];

      if (!field || field.config.custom?.hideFrom?.legend) {
        return undefined;
      }

      const label = getFieldDisplayName(field, data[fieldIndex.frameIndex]!, data);
      const scaleColor = getFieldSeriesColor(field, theme);
      const seriesColor = scaleColor.color;

      return {
        disabled: !(seriesConfig.show ?? true),
        fieldIndex,
        color: seriesColor,
        label,
        yAxis: axisPlacement === AxisPlacement.Left ? 1 : 2,
        getDisplayValues: () => {
          if (!calcs?.length) {
            return [];
          }

          const fmt = field.display ?? defaultFormatter;
          const fieldCalcs = reduceField({
            field,
            reducers: calcs,
          });

          return calcs.map<DisplayValue>((reducerId) => {
            const fieldReducer = fieldReducers.get(reducerId);

            return {
              ...fmt(fieldCalcs[reducerId]),
              title: fieldReducer.name,
              description: fieldReducer.description,
            };
          });
        },
        getItemKey: () => `${label}-${fieldIndex.frameIndex}-${fieldIndex.fieldIndex}`,
      };
    })
    .filter((i) => i !== undefined) as VizLegendItem[];

  return (
    <VizLayout.Legend placement={placement} {...vizLayoutLegendProps}>
      <VizLegend placement={placement} items={legendItems} displayMode={displayMode} />
    </VizLayout.Legend>
  );
};

PlotLegend.displayName = 'PlotLegend';
