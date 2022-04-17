import React from 'react';
import { css, cx } from '@emotion/css';
import { RenderFunction } from '../../types';
import { StoryContext } from '@storybook/react';

const StoryContainer: React.FC<{ width?: number; height?: number; showBoundaries: boolean }> = ({
  children,
  width,
  height,
  showBoundaries,
}) => {
  const checkColor = '#f0f0f0';
  const finalWidth = width ? `${width}px` : '100%';
  const finalHeight = height !== 0 ? `${height}px` : 'auto';
  const bgStyles =
    showBoundaries &&
    css`
      background-color: white;
      background-size: 30px 30px;
      background-position: 0 0, 15px 15px;
      background-image: linear-gradient(
          45deg,
          ${checkColor} 25%,
          transparent 25%,
          transparent 75%,
          ${checkColor} 75%,
          ${checkColor}
        ),
        linear-gradient(45deg, ${checkColor} 25%, transparent 25%, transparent 75%, ${checkColor} 75%, ${checkColor});
    `;
  return (
    <div
      className={cx(
        css`
          width: ${finalWidth};
          height: ${finalHeight};
        `,
        bgStyles
      )}
    >
      {children}
    </div>
  );
};

export const withStoryContainer = (story: RenderFunction, context: StoryContext) => {
  return (
    <StoryContainer
      width={context.args.containerWidth}
      height={context.args.containerHeight}
      showBoundaries={context.args.showBoundaries}
    >
      {story()}
    </StoryContainer>
  );
};
