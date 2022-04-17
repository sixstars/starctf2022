import React from 'react';
import { XYFieldMatchers } from '@grafana/ui/src/components/GraphNG/types';
import {
  ArrayVector,
  DataFrame,
  FALLBACK_COLOR,
  Field,
  FieldColorModeId,
  FieldConfig,
  FieldType,
  formattedValueToString,
  getFieldDisplayName,
  getValueFormat,
  GrafanaTheme2,
  getActiveThreshold,
  Threshold,
  getFieldConfigWithMinMax,
  outerJoinDataFrames,
  ThresholdsMode,
} from '@grafana/data';
import {
  FIXED_UNIT,
  SeriesVisibilityChangeMode,
  UPlotConfigBuilder,
  UPlotConfigPrepFn,
  VizLegendItem,
} from '@grafana/ui';
import { getConfig, TimelineCoreOptions } from './timeline';
import { VizLegendOptions, AxisPlacement, ScaleDirection, ScaleOrientation } from '@grafana/schema';
import { TimelineFieldConfig, TimelineOptions } from './types';
import { PlotTooltipInterpolator } from '@grafana/ui/src/components/uPlot/types';
import { preparePlotData } from '../../../../../packages/grafana-ui/src/components/uPlot/utils';

const defaultConfig: TimelineFieldConfig = {
  lineWidth: 0,
  fillOpacity: 80,
};

export function mapMouseEventToMode(event: React.MouseEvent): SeriesVisibilityChangeMode {
  if (event.ctrlKey || event.metaKey || event.shiftKey) {
    return SeriesVisibilityChangeMode.AppendToSelection;
  }
  return SeriesVisibilityChangeMode.ToggleSelection;
}

export function preparePlotFrame(data: DataFrame[], dimFields: XYFieldMatchers) {
  return outerJoinDataFrames({
    frames: data,
    joinBy: dimFields.x,
    keep: dimFields.y,
    keepOriginIndices: true,
  });
}

export const preparePlotConfigBuilder: UPlotConfigPrepFn<TimelineOptions> = ({
  frame,
  theme,
  timeZone,
  getTimeRange,
  mode,
  rowHeight,
  colWidth,
  showValue,
  alignValue,
}) => {
  const builder = new UPlotConfigBuilder(timeZone);

  const isDiscrete = (field: Field) => {
    const mode = field.config?.color?.mode;
    return !(mode && field.display && mode.startsWith('continuous-'));
  };

  const getValueColor = (seriesIdx: number, value: any) => {
    const field = frame.fields[seriesIdx];

    if (field.display) {
      const disp = field.display(value); // will apply color modes
      if (disp.color) {
        return disp.color;
      }
    }

    return FALLBACK_COLOR;
  };

  const opts: TimelineCoreOptions = {
    // should expose in panel config
    mode: mode!,
    numSeries: frame.fields.length - 1,
    isDiscrete: (seriesIdx) => isDiscrete(frame.fields[seriesIdx]),
    rowHeight: rowHeight!,
    colWidth: colWidth,
    showValue: showValue!,
    alignValue,
    theme,
    label: (seriesIdx) => getFieldDisplayName(frame.fields[seriesIdx], frame),
    getFieldConfig: (seriesIdx) => frame.fields[seriesIdx].config.custom,
    getValueColor,
    getTimeRange,
    // hardcoded formatter for state values
    formatValue: (seriesIdx, value) => formattedValueToString(frame.fields[seriesIdx].display!(value)),
    onHover: (seriesIndex, valueIndex) => {
      hoveredSeriesIdx = seriesIndex;
      hoveredDataIdx = valueIndex;
      shouldChangeHover = true;
    },
    onLeave: () => {
      hoveredSeriesIdx = null;
      hoveredDataIdx = null;
      shouldChangeHover = true;
    },
  };

  let shouldChangeHover = false;
  let hoveredSeriesIdx: number | null = null;
  let hoveredDataIdx: number | null = null;

  const coreConfig = getConfig(opts);

  builder.addHook('init', coreConfig.init);
  builder.addHook('drawClear', coreConfig.drawClear);
  builder.addHook('setCursor', coreConfig.setCursor);

  // in TooltipPlugin, this gets invoked and the result is bound to a setCursor hook
  // which fires after the above setCursor hook, so can take advantage of hoveringOver
  // already set by the above onHover/onLeave callbacks that fire from coreConfig.setCursor
  const interpolateTooltip: PlotTooltipInterpolator = (
    updateActiveSeriesIdx,
    updateActiveDatapointIdx,
    updateTooltipPosition
  ) => {
    if (shouldChangeHover) {
      if (hoveredSeriesIdx != null) {
        updateActiveSeriesIdx(hoveredSeriesIdx);
        updateActiveDatapointIdx(hoveredDataIdx);
      }

      shouldChangeHover = false;
    }

    updateTooltipPosition(hoveredSeriesIdx == null);
  };

  builder.setTooltipInterpolator(interpolateTooltip);

  builder.setPrepData(preparePlotData);

  builder.setCursor(coreConfig.cursor);

  builder.addScale({
    scaleKey: 'x',
    isTime: true,
    orientation: ScaleOrientation.Horizontal,
    direction: ScaleDirection.Right,
    range: coreConfig.xRange,
  });

  builder.addScale({
    scaleKey: FIXED_UNIT, // y
    isTime: false,
    orientation: ScaleOrientation.Vertical,
    direction: ScaleDirection.Up,
    range: coreConfig.yRange,
  });

  builder.addAxis({
    scaleKey: 'x',
    isTime: true,
    splits: coreConfig.xSplits!,
    placement: AxisPlacement.Bottom,
    timeZone,
    theme,
    grid: { show: true },
  });

  builder.addAxis({
    scaleKey: FIXED_UNIT, // y
    isTime: false,
    placement: AxisPlacement.Left,
    splits: coreConfig.ySplits,
    values: coreConfig.yValues,
    grid: { show: false },
    ticks: false,
    gap: 16,
    theme,
  });

  let seriesIndex = 0;

  for (let i = 0; i < frame.fields.length; i++) {
    if (i === 0) {
      continue;
    }

    const field = frame.fields[i];
    const config = field.config as FieldConfig<TimelineFieldConfig>;
    const customConfig: TimelineFieldConfig = {
      ...defaultConfig,
      ...config.custom,
    };

    field.state!.seriesIndex = seriesIndex++;

    // const scaleKey = config.unit || FIXED_UNIT;
    // const colorMode = getFieldColorModeForField(field);

    builder.addSeries({
      scaleKey: FIXED_UNIT,
      pathBuilder: coreConfig.drawPaths,
      pointsBuilder: coreConfig.drawPoints,
      //colorMode,
      lineWidth: customConfig.lineWidth,
      fillOpacity: customConfig.fillOpacity,
      theme,
      show: !customConfig.hideFrom?.viz,
      thresholds: config.thresholds,
      // The following properties are not used in the uPlot config, but are utilized as transport for legend config
      dataFrameFieldIndex: field.state?.origin,
    });
  }

  return builder;
};

export function getNamesToFieldIndex(frame: DataFrame): Map<string, number> {
  const names = new Map<string, number>();
  for (let i = 0; i < frame.fields.length; i++) {
    names.set(getFieldDisplayName(frame.fields[i], frame), i);
  }
  return names;
}

/**
 * If any sequential duplicate values exist, this will return a new array
 * with the future values set to undefined.
 *
 * in:  1,        1,undefined,        1,2,        2,null,2,3
 * out: 1,undefined,undefined,undefined,2,undefined,null,2,3
 */
export function unsetSameFutureValues(values: any[]): any[] | undefined {
  let prevVal = values[0];
  let clone: any[] | undefined = undefined;

  for (let i = 1; i < values.length; i++) {
    let value = values[i];

    if (value === null) {
      prevVal = null;
    } else {
      if (value === prevVal) {
        if (!clone) {
          clone = [...values];
        }
        clone[i] = undefined;
      } else if (value != null) {
        prevVal = value;
      }
    }
  }
  return clone;
}

/**
 * Merge values by the threshold
 */
export function mergeThresholdValues(field: Field, theme: GrafanaTheme2): Field | undefined {
  const thresholds = field.config.thresholds;
  if (field.type !== FieldType.number || !thresholds || !thresholds.steps.length) {
    return undefined;
  }

  const items = getThresholdItems(field.config, theme);
  if (items.length !== thresholds.steps.length) {
    return undefined; // should not happen
  }

  const thresholdToText = new Map<Threshold, string>();
  const textToColor = new Map<string, string>();
  for (let i = 0; i < items.length; i++) {
    thresholdToText.set(thresholds.steps[i], items[i].label);
    textToColor.set(items[i].label, items[i].color!);
  }

  let prev: Threshold | undefined = undefined;
  let input = field.values.toArray();
  const vals = new Array<String | undefined>(field.values.length);
  if (thresholds.mode === ThresholdsMode.Percentage) {
    const { min, max } = getFieldConfigWithMinMax(field);
    const delta = max! - min!;
    input = input.map((v) => {
      if (v == null) {
        return v;
      }
      return ((v - min!) / delta) * 100;
    });
  }

  for (let i = 0; i < vals.length; i++) {
    const v = input[i];
    if (v == null) {
      vals[i] = v;
      prev = undefined;
    }
    const active = getActiveThreshold(v, thresholds.steps);
    if (active === prev) {
      vals[i] = undefined;
    } else {
      vals[i] = thresholdToText.get(active);
    }
    prev = active;
  }

  return {
    ...field,
    type: FieldType.string,
    values: new ArrayVector(vals),
    display: (value: string) => ({
      text: value,
      color: textToColor.get(value),
      numeric: NaN,
    }),
  };
}

// This will return a set of frames with only graphable values included
export function prepareTimelineFields(
  series: DataFrame[] | undefined,
  mergeValues: boolean,
  theme: GrafanaTheme2
): { frames?: DataFrame[]; warn?: string } {
  if (!series?.length) {
    return { warn: 'No data in response' };
  }
  let hasTimeseries = false;
  const frames: DataFrame[] = [];
  for (let frame of series) {
    let isTimeseries = false;
    let changed = false;
    const fields: Field[] = [];
    for (let field of frame.fields) {
      switch (field.type) {
        case FieldType.time:
          isTimeseries = true;
          hasTimeseries = true;
          fields.push(field);
          break;
        case FieldType.number:
          if (mergeValues && field.config.color?.mode === FieldColorModeId.Thresholds) {
            const f = mergeThresholdValues(field, theme);
            if (f) {
              fields.push(f);
              changed = true;
              continue;
            }
          }

        case FieldType.boolean:
        case FieldType.string:
          field = {
            ...field,
            config: {
              ...field.config,
              custom: {
                ...field.config.custom,
                // magic value for join() to leave nulls alone
                spanNulls: -1,
              },
            },
          };

          if (mergeValues) {
            let merged = unsetSameFutureValues(field.values.toArray());
            if (merged) {
              fields.push({
                ...field,
                values: new ArrayVector(merged),
              });
              changed = true;
              continue;
            }
          }
          fields.push(field);
          break;
        default:
          changed = true;
      }
    }
    if (isTimeseries && fields.length > 1) {
      hasTimeseries = true;
      if (changed) {
        frames.push({
          ...frame,
          fields,
        });
      } else {
        frames.push(frame);
      }
    }
  }

  if (!hasTimeseries) {
    return { warn: 'Data does not have a time field' };
  }
  if (!frames.length) {
    return { warn: 'No graphable fields' };
  }
  return { frames };
}

export function getThresholdItems(fieldConfig: FieldConfig, theme: GrafanaTheme2): VizLegendItem[] {
  const items: VizLegendItem[] = [];
  const thresholds = fieldConfig.thresholds;
  if (!thresholds || !thresholds.steps.length) {
    return items;
  }

  const steps = thresholds.steps;
  const disp = getValueFormat(thresholds.mode === ThresholdsMode.Percentage ? 'percent' : fieldConfig.unit ?? '');

  const fmt = (v: number) => formattedValueToString(disp(v));

  for (let i = 1; i <= steps.length; i++) {
    const step = steps[i - 1];
    items.push({
      label: i === 1 ? `< ${fmt(step.value)}` : `${fmt(step.value)}+`,
      color: theme.visualization.getColorByName(step.color),
      yAxis: 1,
    });
  }

  return items;
}

export function prepareTimelineLegendItems(
  frames: DataFrame[] | undefined,
  options: VizLegendOptions,
  theme: GrafanaTheme2
): VizLegendItem[] | undefined {
  if (!frames || options.displayMode === 'hidden') {
    return undefined;
  }

  const fields = allNonTimeFields(frames);
  if (!fields.length) {
    return undefined;
  }

  const items: VizLegendItem[] = [];
  const fieldConfig = fields[0].config;
  const colorMode = fieldConfig.color?.mode ?? FieldColorModeId.Fixed;
  const thresholds = fieldConfig.thresholds;

  // If thresholds are enabled show each step in the legend
  if (colorMode === FieldColorModeId.Thresholds && thresholds?.steps && thresholds.steps.length > 1) {
    return getThresholdItems(fieldConfig, theme);
  }

  // If thresholds are enabled show each step in the legend
  if (colorMode.startsWith('continuous')) {
    return undefined; // eventually a color bar
  }

  let stateColors: Map<string, string | undefined> = new Map();

  fields.forEach((field) => {
    field.values.toArray().forEach((v) => {
      let state = field.display!(v);
      if (state.color) {
        stateColors.set(state.text, state.color!);
      }
    });
  });

  stateColors.forEach((color, label) => {
    if (label.length > 0) {
      items.push({
        label: label!,
        color: theme.visualization.getColorByName(color ?? FALLBACK_COLOR),
        yAxis: 1,
      });
    }
  });

  return items;
}

function allNonTimeFields(frames: DataFrame[]): Field[] {
  const fields: Field[] = [];
  for (const frame of frames) {
    for (const field of frame.fields) {
      if (field.type !== FieldType.time) {
        fields.push(field);
      }
    }
  }
  return fields;
}

export function findNextStateIndex(field: Field, datapointIdx: number) {
  let end;
  let rightPointer = datapointIdx + 1;

  if (rightPointer >= field.values.length) {
    return null;
  }

  while (end === undefined) {
    if (rightPointer >= field.values.length) {
      return null;
    }
    const rightValue = field.values.get(rightPointer);

    if (rightValue !== undefined) {
      end = rightPointer;
    } else {
      rightPointer++;
    }
  }

  return end;
}
