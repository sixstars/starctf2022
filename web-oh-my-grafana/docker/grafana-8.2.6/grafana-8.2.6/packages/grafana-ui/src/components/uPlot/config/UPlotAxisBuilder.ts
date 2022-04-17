import { dateTimeFormat, GrafanaTheme2, isBooleanUnit, systemDateFormats, TimeZone } from '@grafana/data';
import uPlot, { Axis } from 'uplot';
import { PlotConfigBuilder } from '../types';
import { measureText } from '../../../utils/measureText';
import { AxisPlacement } from '@grafana/schema';
import { optMinMax } from './UPlotScaleBuilder';

export interface AxisProps {
  scaleKey: string;
  theme: GrafanaTheme2;
  label?: string;
  show?: boolean;
  size?: number | null;
  gap?: number;
  placement?: AxisPlacement;
  grid?: Axis.Grid;
  ticks?: boolean;
  formatValue?: (v: any) => string;
  incrs?: Axis.Incrs;
  splits?: Axis.Splits;
  values?: Axis.Values;
  isTime?: boolean;
  timeZone?: TimeZone;
}

const fontSize = 12;
const labelPad = 8;

export class UPlotAxisBuilder extends PlotConfigBuilder<AxisProps, Axis> {
  merge(props: AxisProps) {
    this.props.size = optMinMax('max', this.props.size, props.size);
    if (!this.props.label) {
      this.props.label = props.label;
    }
    if (this.props.placement === AxisPlacement.Auto) {
      this.props.placement = props.placement;
    }
  }

  getConfig(): Axis {
    let {
      scaleKey,
      label,
      show = true,
      placement = AxisPlacement.Auto,
      grid = { show: true },
      ticks = true,
      gap = 5,
      formatValue,
      splits,
      values,
      isTime,
      timeZone,
      theme,
    } = this.props;

    const font = `${fontSize}px ${theme.typography.fontFamily}`;

    const gridColor = theme.isDark ? 'rgba(240, 250, 255, 0.09)' : 'rgba(0, 10, 23, 0.09)';

    if (isBooleanUnit(scaleKey)) {
      splits = [0, 1];
    }

    let config: Axis = {
      scale: scaleKey,
      show,
      stroke: theme.colors.text.primary,
      side: getUPlotSideFromAxis(placement),
      font,
      size: this.props.size ?? calculateAxisSize,
      gap,

      labelGap: 0,

      grid: {
        show: grid.show,
        stroke: gridColor,
        width: 1 / devicePixelRatio,
      },
      ticks: {
        show: ticks,
        stroke: gridColor,
        width: 1 / devicePixelRatio,
        size: 4,
      },
      splits,
      values: values,
      space: calculateSpace,
    };

    if (label != null && label.length > 0) {
      config.label = label;
      config.labelSize = fontSize + labelPad;
      config.labelFont = font;
      config.labelGap = labelPad;
    }

    if (values) {
      config.values = values;
    } else if (isTime) {
      config.values = formatTime;
    } else if (formatValue) {
      config.values = (u: uPlot, vals: any[]) => vals.map(formatValue!);
    }

    // store timezone
    (config as any).timeZone = timeZone;

    return config;
  }
}

/* Minimum grid & tick spacing in CSS pixels */
function calculateSpace(self: uPlot, axisIdx: number, scaleMin: number, scaleMax: number, plotDim: number): number {
  const axis = self.axes[axisIdx];
  const scale = self.scales[axis.scale!];

  // for axis left & right
  if (axis.side !== 2 || !scale) {
    return 30;
  }

  const defaultSpacing = 40;

  if (scale.time) {
    const maxTicks = plotDim / defaultSpacing;
    const increment = (scaleMax - scaleMin) / maxTicks;
    const sample = formatTime(self, [scaleMin], axisIdx, defaultSpacing, increment);
    const width = measureText(sample[0], fontSize).width + 18;
    return width;
  }

  return defaultSpacing;
}

/** height of x axis or width of y axis in CSS pixels alloted for values, gap & ticks, but excluding axis label */
function calculateAxisSize(self: uPlot, values: string[], axisIdx: number) {
  const axis = self.axes[axisIdx];

  let axisSize = axis.ticks!.size!;

  if (axis.side === 2) {
    axisSize += axis!.gap! + fontSize;
  } else if (values?.length) {
    let maxTextWidth = values.reduce((acc, value) => Math.max(acc, measureText(value, fontSize).width), 0);
    axisSize += axis!.gap! + axis!.labelGap! + maxTextWidth;
  }

  return Math.ceil(axisSize);
}

const timeUnitSize = {
  second: 1000,
  minute: 60 * 1000,
  hour: 60 * 60 * 1000,
  day: 24 * 60 * 60 * 1000,
  month: 28 * 24 * 60 * 60 * 1000,
  year: 365 * 24 * 60 * 60 * 1000,
};

/** Format time axis ticks */
function formatTime(self: uPlot, splits: number[], axisIdx: number, foundSpace: number, foundIncr: number): string[] {
  const timeZone = (self.axes[axisIdx] as any).timeZone;
  const scale = self.scales.x;
  const range = (scale?.max ?? 0) - (scale?.min ?? 0);
  const yearRoundedToDay = Math.round(timeUnitSize.year / timeUnitSize.day) * timeUnitSize.day;
  const incrementRoundedToDay = Math.round(foundIncr / timeUnitSize.day) * timeUnitSize.day;

  let format = systemDateFormats.interval.year;

  if (foundIncr < timeUnitSize.second) {
    format = systemDateFormats.interval.second.replace('ss', 'ss.SS');
  } else if (foundIncr <= timeUnitSize.minute) {
    format = systemDateFormats.interval.second;
  } else if (range <= timeUnitSize.day) {
    format = systemDateFormats.interval.minute;
  } else if (foundIncr <= timeUnitSize.day) {
    format = systemDateFormats.interval.hour;
  } else if (range < timeUnitSize.year) {
    format = systemDateFormats.interval.day;
  } else if (incrementRoundedToDay === yearRoundedToDay) {
    format = systemDateFormats.interval.year;
  } else if (foundIncr <= timeUnitSize.year) {
    format = systemDateFormats.interval.month;
  }

  return splits.map((v) => dateTimeFormat(v, { format, timeZone }));
}

export function getUPlotSideFromAxis(axis: AxisPlacement) {
  switch (axis) {
    case AxisPlacement.Top:
      return 0;
    case AxisPlacement.Right:
      return 1;
    case AxisPlacement.Bottom:
      return 2;
    case AxisPlacement.Left:
  }

  return 3; // default everythign to the left
}
