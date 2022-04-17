import React, { HTMLAttributes } from 'react';
import { Icon } from '../Icon/Icon';
import { useTheme } from '../../themes/ThemeContext';
import { stylesFactory } from '../../themes/stylesFactory';
import { IconName } from '../../types';
import { Tooltip } from '../Tooltip/Tooltip';
import { getColorForTheme, GrafanaTheme } from '@grafana/data';
import tinycolor from 'tinycolor2';
import { css, cx } from '@emotion/css';
import { HorizontalGroup } from '../Layout/Layout';

export type BadgeColor = 'blue' | 'red' | 'green' | 'orange' | 'purple';

export interface BadgeProps extends HTMLAttributes<HTMLDivElement> {
  text: React.ReactNode;
  color: BadgeColor;
  icon?: IconName;
  tooltip?: string;
}

export const Badge = React.memo<BadgeProps>(({ icon, color, text, tooltip, className, ...otherProps }) => {
  const theme = useTheme();
  const styles = getStyles(theme, color);
  const badge = (
    <div className={cx(styles.wrapper, className)} {...otherProps}>
      <HorizontalGroup align="center" spacing="xs">
        {icon && <Icon name={icon} size="sm" />}
        <span>{text}</span>
      </HorizontalGroup>
    </div>
  );

  return tooltip ? (
    <Tooltip content={tooltip} placement="auto">
      {badge}
    </Tooltip>
  ) : (
    badge
  );
});

Badge.displayName = 'Badge';

const getStyles = stylesFactory((theme: GrafanaTheme, color: BadgeColor) => {
  let sourceColor = getColorForTheme(color, theme);
  let borderColor = '';
  let bgColor = '';
  let textColor = '';

  if (theme.isDark) {
    bgColor = tinycolor(sourceColor).setAlpha(0.15).toString();
    borderColor = tinycolor(sourceColor).darken(30).toString();
    textColor = tinycolor(sourceColor).lighten(15).toString();
  } else {
    bgColor = tinycolor(sourceColor).setAlpha(0.15).toString();
    borderColor = tinycolor(sourceColor).lighten(20).toString();
    textColor = tinycolor(sourceColor).darken(15).toString();
  }

  return {
    wrapper: css`
      font-size: ${theme.typography.size.sm};
      display: inline-flex;
      padding: 1px 4px;
      border-radius: 3px;
      background: ${bgColor};
      border: 1px solid ${borderColor};
      color: ${textColor};
      font-weight: ${theme.typography.weight.regular};

      > span {
        position: relative;
        top: 1px;
        margin-left: 2px;
      }
    `,
  };
});
