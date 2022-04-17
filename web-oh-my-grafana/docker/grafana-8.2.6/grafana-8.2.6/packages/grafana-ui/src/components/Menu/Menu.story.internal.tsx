import React from 'react';
import { Story } from '@storybook/react';
import { Menu, MenuProps } from './Menu';
import { MenuItem } from './MenuItem';
import { MenuGroup } from './MenuGroup';
import { GraphContextMenuHeader } from '..';
import { StoryExample } from '../../utils/storybook/StoryExample';
import { VerticalGroup } from '../Layout/Layout';

export default {
  title: 'General/Menu',
  component: Menu,
  argTypes: {
    items: { control: { disable: true } },
    icon: { control: { type: 'select' } },
  },
  parameters: {
    knobs: {
      disabled: true,
    },
    controls: {
      disabled: true,
    },
    actions: {
      disabled: true,
    },
  },
};

export const Simple: Story<MenuProps> = (args) => {
  return (
    <VerticalGroup>
      <StoryExample name="Simple">
        <Menu>
          <MenuItem label="Google" icon="search-plus" />
          <MenuItem label="Filter" icon="filter" />
          <MenuItem label="History" icon="history" />
          <MenuItem label="Active" icon="history" active />
          <MenuItem label="Apps" icon="apps" />
        </Menu>
      </StoryExample>
      <StoryExample name="With header & groups">
        <Menu header={args.header} ariaLabel="Menu header">
          <MenuGroup label="Group 1">
            <MenuItem label="item1" icon="history" />
            <MenuItem label="item2" icon="filter" />
          </MenuGroup>
          <MenuGroup label="Group 2">
            <MenuItem label="item1" icon="history" />
          </MenuGroup>
        </Menu>
      </StoryExample>
    </VerticalGroup>
  );
};

Simple.args = {
  header: (
    <GraphContextMenuHeader
      timestamp="2020-11-25 19:04:25"
      seriesColor="#00ff00"
      displayName="A-series"
      displayValue={{
        text: '128',
        suffix: 'km/h',
      }}
    />
  ),
};
