import { isNumber } from 'lodash';
import {
  DashboardCursorSync,
  DataFrame,
  DataHoverClearEvent,
  DataHoverEvent,
  DataHoverPayload,
  FieldConfig,
  FieldType,
  formattedValueToString,
  getFieldColorModeForField,
  getFieldSeriesColor,
  getFieldDisplayName,
} from '@grafana/data';

import { UPlotConfigBuilder, UPlotConfigPrepFn } from '../uPlot/config/UPlotConfigBuilder';
import { FIXED_UNIT } from '../GraphNG/GraphNG';
import {
  AxisPlacement,
  GraphDrawStyle,
  GraphFieldConfig,
  GraphTresholdsStyleMode,
  VisibilityMode,
  ScaleDirection,
  ScaleOrientation,
} from '@grafana/schema';
import { collectStackingGroups, preparePlotData } from '../uPlot/utils';
import uPlot from 'uplot';

const defaultFormatter = (v: any) => (v == null ? '-' : v.toFixed(1));

const defaultConfig: GraphFieldConfig = {
  drawStyle: GraphDrawStyle.Line,
  showPoints: VisibilityMode.Auto,
  axisPlacement: AxisPlacement.Auto,
};

export const preparePlotConfigBuilder: UPlotConfigPrepFn<{ sync: DashboardCursorSync }> = ({
  frame,
  theme,
  timeZone,
  getTimeRange,
  eventBus,
  sync,
  allFrames,
}) => {
  const builder = new UPlotConfigBuilder(timeZone);

  builder.setPrepData(preparePlotData);

  // X is the first field in the aligned frame
  const xField = frame.fields[0];
  if (!xField) {
    return builder; // empty frame with no options
  }

  let seriesIndex = 0;

  const xScaleKey = 'x';
  let xScaleUnit = '_x';
  let yScaleKey = '';

  if (xField.type === FieldType.time) {
    xScaleUnit = 'time';
    builder.addScale({
      scaleKey: xScaleKey,
      orientation: ScaleOrientation.Horizontal,
      direction: ScaleDirection.Right,
      isTime: true,
      range: () => {
        const r = getTimeRange();
        return [r.from.valueOf(), r.to.valueOf()];
      },
    });

    builder.addAxis({
      scaleKey: xScaleKey,
      isTime: true,
      placement: AxisPlacement.Bottom,
      timeZone,
      theme,
      grid: { show: xField.config.custom?.axisGridShow },
    });
  } else {
    // Not time!
    if (xField.config.unit) {
      xScaleUnit = xField.config.unit;
    }

    builder.addScale({
      scaleKey: xScaleKey,
      orientation: ScaleOrientation.Horizontal,
      direction: ScaleDirection.Right,
    });

    builder.addAxis({
      scaleKey: xScaleKey,
      placement: AxisPlacement.Bottom,
      theme,
      grid: { show: xField.config.custom?.axisGridShow },
    });
  }

  const stackingGroups: Map<string, number[]> = new Map();

  let indexByName: Map<string, number> | undefined;

  for (let i = 1; i < frame.fields.length; i++) {
    const field = frame.fields[i];
    const config = field.config as FieldConfig<GraphFieldConfig>;
    const customConfig: GraphFieldConfig = {
      ...defaultConfig,
      ...config.custom,
    };

    if (field === xField || field.type !== FieldType.number) {
      continue;
    }
    field.state!.seriesIndex = seriesIndex++;

    const fmt = field.display ?? defaultFormatter;
    const scaleKey = config.unit || FIXED_UNIT;
    const colorMode = getFieldColorModeForField(field);
    const scaleColor = getFieldSeriesColor(field, theme);
    const seriesColor = scaleColor.color;

    // The builder will manage unique scaleKeys and combine where appropriate
    builder.addScale({
      scaleKey,
      orientation: ScaleOrientation.Vertical,
      direction: ScaleDirection.Up,
      distribution: customConfig.scaleDistribution?.type,
      log: customConfig.scaleDistribution?.log,
      min: field.config.min,
      max: field.config.max,
      softMin: customConfig.axisSoftMin,
      softMax: customConfig.axisSoftMax,
    });

    if (!yScaleKey) {
      yScaleKey = scaleKey;
    }

    if (customConfig.axisPlacement !== AxisPlacement.Hidden) {
      builder.addAxis({
        scaleKey,
        label: customConfig.axisLabel,
        size: customConfig.axisWidth,
        placement: customConfig.axisPlacement ?? AxisPlacement.Auto,
        formatValue: (v) => formattedValueToString(fmt(v)),
        theme,
        grid: { show: customConfig.axisGridShow },
      });
    }

    const showPoints =
      customConfig.drawStyle === GraphDrawStyle.Points ? VisibilityMode.Always : customConfig.showPoints;

    let pointsFilter: uPlot.Series.Points.Filter = () => null;

    if (customConfig.spanNulls !== true) {
      pointsFilter = (u, seriesIdx, show, gaps) => {
        let filtered = [];

        let series = u.series[seriesIdx];

        if (!show && gaps && gaps.length) {
          const [firstIdx, lastIdx] = series.idxs!;
          const xData = u.data[0];
          const firstPos = Math.round(u.valToPos(xData[firstIdx], 'x', true));
          const lastPos = Math.round(u.valToPos(xData[lastIdx], 'x', true));

          if (gaps[0][0] === firstPos) {
            filtered.push(firstIdx);
          }

          // show single points between consecutive gaps that share end/start
          for (let i = 0; i < gaps.length; i++) {
            let thisGap = gaps[i];
            let nextGap = gaps[i + 1];

            if (nextGap && thisGap[1] === nextGap[0]) {
              filtered.push(u.posToIdx(thisGap[1], true));
            }
          }

          if (gaps[gaps.length - 1][1] === lastPos) {
            filtered.push(lastIdx);
          }
        }

        return filtered.length ? filtered : null;
      };
    }

    let { fillOpacity } = customConfig;

    if (customConfig.fillBelowTo && field.state?.origin) {
      if (!indexByName) {
        indexByName = getNamesToFieldIndex(frame, allFrames);
      }

      const originFrame = allFrames[field.state.origin.frameIndex];
      const originField = originFrame.fields[field.state.origin.fieldIndex];

      const t = indexByName.get(getFieldDisplayName(originField, originFrame, allFrames));
      const b = indexByName.get(customConfig.fillBelowTo);
      if (isNumber(b) && isNumber(t)) {
        builder.addBand({
          series: [t, b],
          fill: null as any, // using null will have the band use fill options from `t`
        });
      }
      if (!fillOpacity) {
        fillOpacity = 35; // default from flot
      }
    }

    builder.addSeries({
      scaleKey,
      showPoints,
      pointsFilter,
      colorMode,
      fillOpacity,
      theme,
      drawStyle: customConfig.drawStyle!,
      lineColor: customConfig.lineColor ?? seriesColor,
      lineWidth: customConfig.lineWidth,
      lineInterpolation: customConfig.lineInterpolation,
      lineStyle: customConfig.lineStyle,
      barAlignment: customConfig.barAlignment,
      barWidthFactor: customConfig.barWidthFactor,
      barMaxWidth: customConfig.barMaxWidth,
      pointSize: customConfig.pointSize,
      spanNulls: customConfig.spanNulls || false,
      show: !customConfig.hideFrom?.viz,
      gradientMode: customConfig.gradientMode,
      thresholds: config.thresholds,
      hardMin: field.config.min,
      hardMax: field.config.max,
      softMin: customConfig.axisSoftMin,
      softMax: customConfig.axisSoftMax,
      // The following properties are not used in the uPlot config, but are utilized as transport for legend config
      dataFrameFieldIndex: field.state?.origin,
    });

    // Render thresholds in graph
    if (customConfig.thresholdsStyle && config.thresholds) {
      const thresholdDisplay = customConfig.thresholdsStyle.mode ?? GraphTresholdsStyleMode.Off;
      if (thresholdDisplay !== GraphTresholdsStyleMode.Off) {
        builder.addThresholds({
          config: customConfig.thresholdsStyle,
          thresholds: config.thresholds,
          scaleKey,
          theme,
          hardMin: field.config.min,
          hardMax: field.config.max,
          softMin: customConfig.axisSoftMin,
          softMax: customConfig.axisSoftMax,
        });
      }
    }
    collectStackingGroups(field, stackingGroups, seriesIndex);
  }

  if (stackingGroups.size !== 0) {
    builder.setStacking(true);
    for (const [_, seriesIdxs] of stackingGroups.entries()) {
      for (let j = seriesIdxs.length - 1; j > 0; j--) {
        builder.addBand({
          series: [seriesIdxs[j], seriesIdxs[j - 1]],
        });
      }
    }
  }

  builder.scaleKeys = [xScaleKey, yScaleKey];

  // if hovered value is null, how far we may scan left/right to hover nearest non-null
  const hoverProximityPx = 15;

  let cursor: Partial<uPlot.Cursor> = {
    // this scans left and right from cursor position to find nearest data index with value != null
    // TODO: do we want to only scan past undefined values, but halt at explicit null values?
    dataIdx: (self, seriesIdx, hoveredIdx, cursorXVal) => {
      let seriesData = self.data[seriesIdx];

      if (seriesData[hoveredIdx] == null) {
        let nonNullLft = hoveredIdx,
          nonNullRgt = hoveredIdx,
          i;

        i = hoveredIdx;
        while (nonNullLft === hoveredIdx && i-- > 0) {
          if (seriesData[i] != null) {
            nonNullLft = i;
          }
        }

        i = hoveredIdx;
        while (nonNullRgt === hoveredIdx && i++ < seriesData.length) {
          if (seriesData[i] != null) {
            nonNullRgt = i;
          }
        }

        let xVals = self.data[0];

        let curPos = self.valToPos(cursorXVal, 'x');
        let rgtPos = self.valToPos(xVals[nonNullRgt], 'x');
        let lftPos = self.valToPos(xVals[nonNullLft], 'x');

        let lftDelta = curPos - lftPos;
        let rgtDelta = rgtPos - curPos;

        if (lftDelta <= rgtDelta) {
          if (lftDelta <= hoverProximityPx) {
            hoveredIdx = nonNullLft;
          }
        } else {
          if (rgtDelta <= hoverProximityPx) {
            hoveredIdx = nonNullRgt;
          }
        }
      }

      return hoveredIdx;
    },
  };

  if (sync !== DashboardCursorSync.Off) {
    const payload: DataHoverPayload = {
      point: {
        [xScaleKey]: null,
        [yScaleKey]: null,
      },
      data: frame,
    };
    const hoverEvent = new DataHoverEvent(payload);
    cursor.sync = {
      key: '__global_',
      filters: {
        pub: (type: string, src: uPlot, x: number, y: number, w: number, h: number, dataIdx: number) => {
          payload.rowIndex = dataIdx;
          if (x < 0 && y < 0) {
            payload.point[xScaleUnit] = null;
            payload.point[yScaleKey] = null;
            eventBus.publish(new DataHoverClearEvent(payload));
          } else {
            // convert the points
            payload.point[xScaleUnit] = src.posToVal(x, xScaleKey);
            payload.point[yScaleKey] = src.posToVal(y, yScaleKey);
            eventBus.publish(hoverEvent);
            hoverEvent.payload.down = undefined;
          }
          return true;
        },
      },
      // ??? setSeries: syncMode === DashboardCursorSync.Tooltip,
      scales: builder.scaleKeys,
      match: [() => true, () => true],
    };
  }

  builder.setSync();
  builder.setCursor(cursor);

  return builder;
};

export function getNamesToFieldIndex(frame: DataFrame, allFrames: DataFrame[]): Map<string, number> {
  const originNames = new Map<string, number>();
  for (let i = 0; i < frame.fields.length; i++) {
    const origin = frame.fields[i].state?.origin;
    if (origin) {
      originNames.set(
        getFieldDisplayName(
          allFrames[origin.frameIndex].fields[origin.fieldIndex],
          allFrames[origin.frameIndex],
          allFrames
        ),
        i
      );
    }
  }
  return originNames;
}
