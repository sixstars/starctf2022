import React from 'react';
import { css } from '@emotion/css';
import { IconButton, IconButtonVariant } from './IconButton';
import { withCenteredStory } from '../../utils/storybook/withCenteredStory';
import { useTheme2 } from '../../themes';
import { IconSize, IconName } from '../../types';
import mdx from './IconButton.mdx';
import { VerticalGroup } from '../Layout/Layout';

export default {
  title: 'Buttons/IconButton',
  component: IconButton,
  decorators: [withCenteredStory],
  parameters: {
    docs: {
      page: mdx,
    },
  },
};

export const Simple = () => {
  return (
    <div>
      <RenderScenario background="canvas" />
      <RenderScenario background="primary" />
      <RenderScenario background="secondary" />
    </div>
  );
};

interface ScenarioProps {
  background: 'canvas' | 'primary' | 'secondary';
}

const RenderScenario = ({ background }: ScenarioProps) => {
  const theme = useTheme2();
  const sizes: IconSize[] = ['sm', 'md', 'lg', 'xl', 'xxl'];
  const icons: IconName[] = ['search', 'trash-alt', 'arrow-left', 'times'];
  const variants: IconButtonVariant[] = ['secondary', 'primary', 'destructive'];

  return (
    <div
      className={css`
        padding: 30px;
        background: ${theme.colors.background[background]};
        button {
          margin-right: 8px;
          margin-left: 8px;
          margin-bottom: 8px;
        }
      `}
    >
      <VerticalGroup spacing="md">
        <div>{background}</div>
        {variants.map((variant) => {
          return (
            <div key={variant}>
              {icons.map((icon) => {
                return sizes.map((size) => (
                  <span key={icon + size}>
                    <IconButton name={icon} size={size} variant={variant} />
                  </span>
                ));
              })}
            </div>
          );
        })}
      </VerticalGroup>
    </div>
  );
};
