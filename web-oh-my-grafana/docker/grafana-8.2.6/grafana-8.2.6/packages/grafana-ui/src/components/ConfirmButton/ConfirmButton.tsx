import React, { PureComponent, SyntheticEvent } from 'react';
import { cx, css } from '@emotion/css';
import { stylesFactory, withTheme } from '../../themes';
import { GrafanaTheme } from '@grafana/data';
import { Themeable } from '../../types';
import { ComponentSize } from '../../types/size';
import { Button, ButtonVariant } from '../Button';

export interface Props extends Themeable {
  /** Confirm action callback */
  onConfirm(): void;
  /** Custom button styles */
  className?: string;
  /** Button size */
  size?: ComponentSize;
  /** Text for the Confirm button */
  confirmText?: string;
  /** Disable button click action */
  disabled?: boolean;
  /** Variant of the Confirm button */
  confirmVariant?: ButtonVariant;
  /** Hide confirm actions when after of them is clicked */
  closeOnConfirm?: boolean;

  /** Optional on click handler for the original button */
  onClick?(): void;
  /** Callback for the cancel action */
  onCancel?(): void;
}

interface State {
  showConfirm: boolean;
}

class UnThemedConfirmButton extends PureComponent<Props, State> {
  state: State = {
    showConfirm: false,
  };

  onClickButton = (event: SyntheticEvent) => {
    if (event) {
      event.preventDefault();
    }

    this.setState({
      showConfirm: true,
    });

    if (this.props.onClick) {
      this.props.onClick();
    }
  };

  onClickCancel = (event: SyntheticEvent) => {
    if (event) {
      event.preventDefault();
    }
    this.setState({
      showConfirm: false,
    });
    if (this.props.onCancel) {
      this.props.onCancel();
    }
  };
  onConfirm = (event: SyntheticEvent) => {
    if (event) {
      event.preventDefault();
    }
    this.props.onConfirm();
    if (this.props.closeOnConfirm) {
      this.setState({
        showConfirm: false,
      });
    }
  };

  render() {
    const {
      theme,
      className,
      size,
      disabled,
      confirmText,
      confirmVariant: confirmButtonVariant,
      children,
    } = this.props;
    const styles = getStyles(theme);
    const buttonClass = cx(
      className,
      this.state.showConfirm ? styles.buttonHide : styles.buttonShow,
      disabled && styles.buttonDisabled
    );
    const confirmButtonClass = cx(
      styles.confirmButton,
      this.state.showConfirm ? styles.confirmButtonShow : styles.confirmButtonHide
    );

    const onClick = disabled ? () => {} : this.onClickButton;

    return (
      <span className={styles.buttonContainer}>
        {typeof children === 'string' ? (
          <span className={buttonClass}>
            <Button size={size} fill="text" onClick={onClick}>
              {children}
            </Button>
          </span>
        ) : (
          <span className={buttonClass} onClick={onClick}>
            {children}
          </span>
        )}
        <span className={confirmButtonClass}>
          <Button size={size} fill="text" onClick={this.onClickCancel}>
            Cancel
          </Button>
          <Button size={size} variant={confirmButtonVariant} onClick={this.onConfirm}>
            {confirmText}
          </Button>
        </span>
      </span>
    );
  }
}

export const ConfirmButton = withTheme(UnThemedConfirmButton);

const getStyles = stylesFactory((theme: GrafanaTheme) => {
  return {
    buttonContainer: css`
      direction: rtl;
      display: flex;
      align-items: center;
    `,
    buttonDisabled: css`
      text-decoration: none;
      color: ${theme.colors.text};
      opacity: 0.65;
      cursor: not-allowed;
      pointer-events: none;
    `,
    buttonShow: css`
      opacity: 1;
      transition: opacity 0.1s ease;
      z-index: 2;
    `,
    buttonHide: css`
      opacity: 0;
      transition: opacity 0.1s ease;
      z-index: 0;
    `,
    confirmButton: css`
      align-items: flex-start;
      background: ${theme.colors.bg1};
      display: flex;
      overflow: hidden;
      position: absolute;
      pointer-events: none;
    `,
    confirmButtonShow: css`
      z-index: 1;
      opacity: 1;
      transition: opacity 0.08s ease-out, transform 0.1s ease-out;
      transform: translateX(0);
      pointer-events: all;
    `,
    confirmButtonHide: css`
      opacity: 0;
      transition: opacity 0.12s ease-in, transform 0.14s ease-in;
      transform: translateX(100px);
    `,
  };
});

// Declare defaultProps directly on the themed component so they are displayed
// in the props table
ConfirmButton.defaultProps = {
  size: 'md',
  confirmText: 'Save',
  disabled: false,
  confirmVariant: 'primary',
};
ConfirmButton.displayName = 'ConfirmButton';
