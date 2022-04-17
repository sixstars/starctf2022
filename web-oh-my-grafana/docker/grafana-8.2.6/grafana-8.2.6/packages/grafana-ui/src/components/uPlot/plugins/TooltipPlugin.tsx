import {
  CartesianCoords2D,
  DashboardCursorSync,
  DataFrame,
  FALLBACK_COLOR,
  FieldType,
  formattedValueToString,
  getDisplayProcessor,
  getFieldDisplayName,
  TimeZone,
} from '@grafana/data';
import { TooltipDisplayMode } from '@grafana/schema';
import React, { useEffect, useLayoutEffect, useState } from 'react';
import { useMountedState } from 'react-use';
import uPlot from 'uplot';
import { useTheme2 } from '../../../themes/ThemeContext';
import { Portal } from '../../Portal/Portal';
import { SeriesTable, SeriesTableRowProps, VizTooltipContainer } from '../../VizTooltip';
import { UPlotConfigBuilder } from '../config/UPlotConfigBuilder';
import { findMidPointYPosition, pluginLog } from '../utils';

interface TooltipPluginProps {
  timeZone: TimeZone;
  data: DataFrame;
  config: UPlotConfigBuilder;
  mode?: TooltipDisplayMode;
  sync?: DashboardCursorSync;
  // Allows custom tooltip content rendering. Exposes aligned data frame with relevant indexes for data inspection
  // Use field.state.origin indexes from alignedData frame field to get access to original data frame and field index.
  renderTooltip?: (alignedFrame: DataFrame, seriesIdx: number | null, datapointIdx: number | null) => React.ReactNode;
}

const TOOLTIP_OFFSET = 10;

/**
 * @alpha
 */
export const TooltipPlugin: React.FC<TooltipPluginProps> = ({
  mode = TooltipDisplayMode.Single,
  sync,
  timeZone,
  config,
  renderTooltip,
  ...otherProps
}) => {
  const theme = useTheme2();
  const [focusedSeriesIdx, setFocusedSeriesIdx] = useState<number | null>(null);
  const [focusedPointIdx, setFocusedPointIdx] = useState<number | null>(null);
  const [focusedPointIdxs, setFocusedPointIdxs] = useState<Array<number | null>>([]);
  const [coords, setCoords] = useState<CartesianCoords2D | null>(null);
  const [isActive, setIsActive] = useState<boolean>(false);
  const isMounted = useMountedState();

  const pluginId = `TooltipPlugin`;

  // Debug logs
  useEffect(() => {
    pluginLog(pluginId, true, `Focused series: ${focusedSeriesIdx}, focused point: ${focusedPointIdx}`);
  }, [focusedPointIdx, focusedSeriesIdx]);

  // Add uPlot hooks to the config, or re-add when the config changed
  useLayoutEffect(() => {
    let plotInstance: uPlot | undefined = undefined;
    let bbox: DOMRect | undefined = undefined;

    const plotMouseLeave = () => {
      if (!isMounted()) {
        return;
      }
      setCoords(null);
      setIsActive(false);
      plotInstance?.root.classList.remove('plot-active');
    };

    const plotMouseEnter = () => {
      if (!isMounted()) {
        return;
      }
      setIsActive(true);
      plotInstance?.root.classList.add('plot-active');
    };

    // cache uPlot plotting area bounding box
    config.addHook('syncRect', (u, rect) => {
      bbox = rect;
    });

    config.addHook('init', (u) => {
      plotInstance = u;

      u.over.addEventListener('mouseleave', plotMouseLeave);
      u.over.addEventListener('mouseenter', plotMouseEnter);

      if (sync === DashboardCursorSync.Crosshair) {
        u.root.classList.add('shared-crosshair');
      }
    });

    const tooltipInterpolator = config.getTooltipInterpolator();

    if (tooltipInterpolator) {
      // Custom toolitp positioning
      config.addHook('setCursor', (u) => {
        tooltipInterpolator(
          setFocusedSeriesIdx,
          setFocusedPointIdx,
          (clear) => {
            if (clear) {
              setCoords(null);
              return;
            }

            if (!bbox) {
              return;
            }

            const { x, y } = positionTooltip(u, bbox);
            if (x !== undefined && y !== undefined) {
              setCoords({ x, y });
            }
          },
          u
        );
      });
    } else {
      config.addHook('setLegend', (u) => {
        if (!isMounted()) {
          return;
        }
        setFocusedPointIdx(u.legend.idx!);
        setFocusedPointIdxs(u.legend.idxs!.slice());
      });

      // default series/datapoint idx retireval
      config.addHook('setCursor', (u) => {
        if (!bbox || !isMounted()) {
          return;
        }

        const { x, y } = positionTooltip(u, bbox);
        if (x !== undefined && y !== undefined) {
          setCoords({ x, y });
        } else {
          setCoords(null);
        }
      });

      config.addHook('setSeries', (_, idx) => {
        if (!isMounted()) {
          return;
        }
        setFocusedSeriesIdx(idx);
      });
    }

    return () => {
      setCoords(null);
      if (plotInstance) {
        plotInstance.over.removeEventListener('mouseleave', plotMouseLeave);
        plotInstance.over.removeEventListener('mouseenter', plotMouseEnter);
      }
    };
  }, [config, setCoords, setIsActive, setFocusedPointIdx, setFocusedPointIdxs]);

  if (focusedPointIdx === null || (!isActive && sync === DashboardCursorSync.Crosshair)) {
    return null;
  }

  // GraphNG expects aligned data, let's take field 0 as x field. FTW
  let xField = otherProps.data.fields[0];
  if (!xField) {
    return null;
  }
  const xFieldFmt = xField.display || getDisplayProcessor({ field: xField, timeZone, theme });
  let tooltip: React.ReactNode = null;

  const xVal = xFieldFmt(xField!.values.get(focusedPointIdx)).text;

  if (!renderTooltip) {
    // when interacting with a point in single mode
    if (mode === TooltipDisplayMode.Single && focusedSeriesIdx !== null) {
      const field = otherProps.data.fields[focusedSeriesIdx];

      if (!field) {
        return null;
      }

      const fieldFmt = field.display || getDisplayProcessor({ field, timeZone, theme });
      const display = fieldFmt(field.values.get(focusedPointIdx));

      tooltip = (
        <SeriesTable
          series={[
            {
              color: display.color || FALLBACK_COLOR,
              label: getFieldDisplayName(field, otherProps.data),
              value: display ? formattedValueToString(display) : null,
            },
          ]}
          timestamp={xVal}
        />
      );
    }

    if (mode === TooltipDisplayMode.Multi) {
      let series: SeriesTableRowProps[] = [];
      const frame = otherProps.data;
      const fields = frame.fields;

      for (let i = 0; i < fields.length; i++) {
        const field = frame.fields[i];
        if (
          !field ||
          field === xField ||
          field.type === FieldType.time ||
          field.type !== FieldType.number ||
          field.config.custom?.hideFrom?.tooltip ||
          field.config.custom?.hideFrom?.viz
        ) {
          continue;
        }

        const display = field.display!(otherProps.data.fields[i].values.get(focusedPointIdxs[i]!));

        series.push({
          color: display.color || FALLBACK_COLOR,
          label: getFieldDisplayName(field, frame),
          value: display ? formattedValueToString(display) : null,
          isActive: focusedSeriesIdx === i,
        });
      }

      tooltip = <SeriesTable series={series} timestamp={xVal} />;
    }
  } else {
    tooltip = renderTooltip(otherProps.data, focusedSeriesIdx, focusedPointIdx);
  }

  return (
    <Portal>
      {tooltip && coords && (
        <VizTooltipContainer position={{ x: coords.x, y: coords.y }} offset={{ x: TOOLTIP_OFFSET, y: TOOLTIP_OFFSET }}>
          {tooltip}
        </VizTooltipContainer>
      )}
    </Portal>
  );
};

function isCursourOutsideCanvas({ left, top }: uPlot.Cursor, canvas: DOMRect) {
  if (left === undefined || top === undefined) {
    return false;
  }
  return left < 0 || left > canvas.width || top < 0 || top > canvas.height;
}

/**
 * Given uPlot cursor position, figure out position of the tooltip withing the canvas bbox
 * Tooltip is positioned relatively to a viewport
 * @internal
 **/
export function positionTooltip(u: uPlot, bbox: DOMRect) {
  let x, y;
  const cL = u.cursor.left || 0;
  const cT = u.cursor.top || 0;

  if (isCursourOutsideCanvas(u.cursor, bbox)) {
    const idx = u.posToIdx(cL);
    // when cursor outside of uPlot's canvas
    if (cT < 0 || cT > bbox.height) {
      let pos = findMidPointYPosition(u, idx);

      if (pos) {
        y = bbox.top + pos;
        if (cL >= 0 && cL <= bbox.width) {
          // find x-scale position for a current cursor left position
          x = bbox.left + u.valToPos(u.data[0][u.posToIdx(cL)], u.series[0].scale!);
        }
      }
    }
  } else {
    x = bbox.left + cL;
    y = bbox.top + cT;
  }

  return { x, y };
}
