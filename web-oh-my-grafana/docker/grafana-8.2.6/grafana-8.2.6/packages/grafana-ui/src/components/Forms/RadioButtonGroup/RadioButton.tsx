import React from 'react';
import { useTheme2, stylesFactory } from '../../../themes';
import { GrafanaTheme2 } from '@grafana/data';
import { css } from '@emotion/css';
import { getPropertiesForButtonSize } from '../commonStyles';
import { getFocusStyles, getMouseFocusStyles } from '../../../themes/mixins';
import { StringSelector } from '@grafana/e2e-selectors';

export type RadioButtonSize = 'sm' | 'md';

export interface RadioButtonProps {
  size?: RadioButtonSize;
  disabled?: boolean;
  name?: string;
  description?: string;
  active: boolean;
  id: string;
  onChange: () => void;
  fullWidth?: boolean;
  'aria-label'?: StringSelector;
}

export const RadioButton: React.FC<RadioButtonProps> = ({
  children,
  active = false,
  disabled = false,
  size = 'md',
  onChange,
  id,
  name = undefined,
  description,
  fullWidth,
  'aria-label': ariaLabel,
}) => {
  const theme = useTheme2();
  const styles = getRadioButtonStyles(theme, size, fullWidth);

  return (
    <>
      <input
        type="radio"
        className={styles.radio}
        onChange={onChange}
        disabled={disabled}
        id={id}
        checked={active}
        name={name}
        aria-label={ariaLabel}
      />
      <label className={styles.radioLabel} htmlFor={id} title={description}>
        {children}
      </label>
    </>
  );
};

RadioButton.displayName = 'RadioButton';

const getRadioButtonStyles = stylesFactory((theme: GrafanaTheme2, size: RadioButtonSize, fullWidth?: boolean) => {
  const { fontSize, height, padding } = getPropertiesForButtonSize(size, theme);

  const textColor = theme.colors.text.secondary;
  const textColorHover = theme.colors.text.primary;
  // remove the group inner padding (set on RadioButtonGroup)
  const labelHeight = height * theme.spacing.gridSize - 4 - 2;

  return {
    radio: css`
      position: absolute;
      opacity: 0;
      z-index: -1000;

      &:checked + label {
        color: ${theme.colors.text.primary};
        font-weight: ${theme.typography.fontWeightMedium};
        background: ${theme.colors.action.selected};
        z-index: 3;
      }

      &:focus + label,
      &:focus-visible + label {
        ${getFocusStyles(theme)};
      }

      &:focus:not(:focus-visible) + label {
        ${getMouseFocusStyles(theme)}
      }

      &:disabled + label {
        color: ${theme.colors.text.disabled};
        cursor: not-allowed;
      }
    `,
    radioLabel: css`
      display: inline-block;
      position: relative;
      font-size: ${fontSize};
      height: ${labelHeight}px;
      // Deduct border from line-height for perfect vertical centering on windows and linux
      line-height: ${labelHeight}px;
      color: ${textColor};
      padding: ${theme.spacing(0, padding)};
      border-radius: ${theme.shape.borderRadius()};
      background: ${theme.colors.background.primary};
      cursor: pointer;
      z-index: 1;
      flex: ${fullWidth ? `1 0 0` : 'none'};
      text-align: center;
      user-select: none;
      white-space: nowrap;

      &:hover {
        color: ${textColorHover};
      }
    `,
  };
});
