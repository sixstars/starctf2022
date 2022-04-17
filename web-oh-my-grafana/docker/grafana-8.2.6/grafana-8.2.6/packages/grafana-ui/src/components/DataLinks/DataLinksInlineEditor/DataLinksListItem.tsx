import React, { FC } from 'react';
import { css, cx } from '@emotion/css';
import { DataFrame, DataLink, GrafanaTheme2 } from '@grafana/data';
import { stylesFactory, useTheme2 } from '../../../themes';
import { HorizontalGroup, VerticalGroup } from '../../Layout/Layout';
import { IconButton } from '../../IconButton/IconButton';

export interface DataLinksListItemProps {
  index: number;
  link: DataLink;
  data: DataFrame[];
  onChange: (index: number, link: DataLink) => void;
  onEdit: () => void;
  onRemove: () => void;
  isEditing?: boolean;
}

export const DataLinksListItem: FC<DataLinksListItemProps> = ({ link, onEdit, onRemove }) => {
  const theme = useTheme2();
  const styles = getDataLinkListItemStyles(theme);
  const { title = '', url = '' } = link;

  const hasTitle = title.trim() !== '';
  const hasUrl = url.trim() !== '';

  return (
    <div className={styles.wrapper}>
      <VerticalGroup spacing="xs">
        <HorizontalGroup justify="space-between" align="flex-start" width="100%">
          <div className={cx(styles.title, !hasTitle && styles.notConfigured)}>
            {hasTitle ? title : 'Data link title not provided'}
          </div>
          <HorizontalGroup>
            <IconButton name="pen" onClick={onEdit} />
            <IconButton name="times" onClick={onRemove} />
          </HorizontalGroup>
        </HorizontalGroup>
        <div className={cx(styles.url, !hasUrl && styles.notConfigured)} title={url}>
          {hasUrl ? url : 'Data link url not provided'}
        </div>
      </VerticalGroup>
    </div>
  );
};

const getDataLinkListItemStyles = stylesFactory((theme: GrafanaTheme2) => {
  return {
    wrapper: css`
      margin-bottom: ${theme.spacing(2)};
      width: 100%;
      &:last-child {
        margin-bottom: 0;
      }
    `,
    notConfigured: css`
      font-style: italic;
    `,
    title: css`
      color: ${theme.colors.text.primary};
      font-size: ${theme.typography.size.sm};
      font-weight: ${theme.typography.fontWeightMedium};
    `,
    url: css`
      color: ${theme.colors.text.secondary};
      font-size: ${theme.typography.size.sm};
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      max-width: 90%;
    `,
  };
});
