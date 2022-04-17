import React, { FC } from 'react';
import { GrafanaTheme2 } from '@grafana/data';
import { css } from '@emotion/css';
import { useStyles2 } from '../../themes';

export interface Props {
  children: JSX.Element | string;
}

const EmptySearchResult: FC<Props> = ({ children }) => {
  const styles = useStyles2(getStyles);
  return <div className={styles.container}>{children}</div>;
};

const getStyles = (theme: GrafanaTheme2) => {
  return {
    container: css`
      border-left: 3px solid ${theme.colors.info.main};
      background-color: ${theme.colors.background.secondary};
      padding: ${theme.spacing(2)};
      min-width: 350px;
      border-radius: ${theme.shape.borderRadius(2)};
      margin-bottom: ${theme.spacing(4)};
    `,
  };
};
export { EmptySearchResult };
