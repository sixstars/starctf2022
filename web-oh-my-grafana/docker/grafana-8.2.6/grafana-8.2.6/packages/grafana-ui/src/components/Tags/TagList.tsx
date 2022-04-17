import React, { FC, memo } from 'react';
import { css, cx } from '@emotion/css';
import { OnTagClick, Tag } from './Tag';

export interface Props {
  tags: string[];
  onClick?: OnTagClick;
  /** Custom styles for the wrapper component */
  className?: string;
}

export const TagList: FC<Props> = memo(({ tags, onClick, className }) => {
  const styles = getStyles();

  return (
    <span className={cx(styles.wrapper, className)}>
      {tags.map((tag) => (
        <Tag key={tag} name={tag} onClick={onClick} className={styles.tag} />
      ))}
    </span>
  );
});

TagList.displayName = 'TagList';

const getStyles = () => {
  return {
    wrapper: css`
      display: flex;
      flex: 1 1 auto;
      flex-wrap: wrap;
      margin-bottom: -6px;
      justify-content: flex-end;
    `,
    tag: css`
      margin: 0 0 6px 6px;
    `,
  };
};
