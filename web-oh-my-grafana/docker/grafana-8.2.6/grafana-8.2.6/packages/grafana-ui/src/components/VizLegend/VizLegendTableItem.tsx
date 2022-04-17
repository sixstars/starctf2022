import React, { useCallback } from 'react';
import { css, cx } from '@emotion/css';
import { VizLegendSeriesIcon } from './VizLegendSeriesIcon';
import { VizLegendItem } from './types';
import { useStyles2 } from '../../themes/ThemeContext';
import { styleMixins } from '../../themes';
import { formattedValueToString, GrafanaTheme2 } from '@grafana/data';

export interface Props {
  key?: React.Key;
  item: VizLegendItem;
  className?: string;
  onLabelClick?: (item: VizLegendItem, event: React.MouseEvent<HTMLDivElement>) => void;
  onLabelMouseEnter?: (item: VizLegendItem, event: React.MouseEvent<HTMLDivElement>) => void;
  onLabelMouseOut?: (item: VizLegendItem, event: React.MouseEvent<HTMLDivElement>) => void;
  readonly?: boolean;
}

/**
 * @internal
 */
export const LegendTableItem: React.FunctionComponent<Props> = ({
  item,
  onLabelClick,
  onLabelMouseEnter,
  onLabelMouseOut,
  className,
  readonly,
}) => {
  const styles = useStyles2(getStyles);

  const onMouseEnter = useCallback(
    (event: React.MouseEvent<HTMLDivElement, MouseEvent>) => {
      if (onLabelMouseEnter) {
        onLabelMouseEnter(item, event);
      }
    },
    [item, onLabelMouseEnter]
  );

  const onMouseOut = useCallback(
    (event: React.MouseEvent<HTMLDivElement, MouseEvent>) => {
      if (onLabelMouseOut) {
        onLabelMouseOut(item, event);
      }
    },
    [item, onLabelMouseOut]
  );

  const onClick = useCallback(
    (event: React.MouseEvent<HTMLDivElement, MouseEvent>) => {
      if (onLabelClick) {
        onLabelClick(item, event);
      }
    },
    [item, onLabelClick]
  );

  return (
    <tr className={cx(styles.row, className)}>
      <td>
        <span className={styles.itemWrapper}>
          <VizLegendSeriesIcon color={item.color} seriesName={item.label} readonly={readonly} />
          <div
            onMouseEnter={onMouseEnter}
            onMouseOut={onMouseOut}
            onClick={!readonly ? onClick : undefined}
            className={cx(styles.label, item.disabled && styles.labelDisabled, !readonly && styles.clickable)}
          >
            {item.label} {item.yAxis === 2 && <span className={styles.yAxisLabel}>(right y-axis)</span>}
          </div>
        </span>
      </td>
      {item.getDisplayValues &&
        item.getDisplayValues().map((stat, index) => {
          return (
            <td className={styles.value} key={`${stat.title}-${index}`}>
              {formattedValueToString(stat)}
            </td>
          );
        })}
    </tr>
  );
};

LegendTableItem.displayName = 'LegendTableItem';

const getStyles = (theme: GrafanaTheme2) => {
  const rowHoverBg = styleMixins.hoverColor(theme.colors.background.primary, theme);

  return {
    row: css`
      label: LegendRow;
      font-size: ${theme.v1.typography.size.sm};
      border-bottom: 1px solid ${theme.colors.border.weak};
      td {
        padding: ${theme.spacing(0.25, 1)};
        white-space: nowrap;
      }

      &:hover {
        background: ${rowHoverBg};
      }
    `,
    label: css`
      label: LegendLabel;
      white-space: nowrap;
    `,
    labelDisabled: css`
      label: LegendLabelDisabled;
      color: ${theme.colors.text.disabled};
    `,
    clickable: css`
      label: LegendClickable;
      cursor: pointer;
    `,
    itemWrapper: css`
      display: flex;
      white-space: nowrap;
      align-items: center;
    `,
    value: css`
      text-align: right;
    `,
    yAxisLabel: css`
      color: ${theme.colors.text.secondary};
    `,
  };
};
