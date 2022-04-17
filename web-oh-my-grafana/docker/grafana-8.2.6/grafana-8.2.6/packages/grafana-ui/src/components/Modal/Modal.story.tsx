import React, { useState } from 'react';
import { oneLineTrim } from 'common-tags';
import { Story, Meta } from '@storybook/react';
import { getAvailableIcons } from '../../types';
import { Button, Modal, ModalTabsHeader, TabContent } from '@grafana/ui';
import { css, cx } from '@emotion/css';

import { withCenteredStory } from '../../utils/storybook/withCenteredStory';
import { UseState } from '../../utils/storybook/UseState';
import mdx from './Modal.mdx';

export default {
  title: 'Overlays/Modal',
  component: Modal,
  decorators: [withCenteredStory],
  parameters: {
    docs: {
      page: mdx,
    },
    controls: {
      exclude: ['className', 'contentClassName', 'onDismiss', 'onClickBackdrop'],
    },
  },
  args: {
    body: oneLineTrim(`Id incididunt do pariatur qui labore. Sint culpa irure cillum et ullamco proident. Deserunt ipsum velit dolore est enim proident dolore consectetur. Et cillum tempor pariatur et. Est tempor cillum ad id nulla. Cillum ut proident
    magna do cillum consequat reprehenderit excepteur. Pariatur culpa id excepteur reprehenderit consequat qui qui sit
    consectetur esse enim mollit incididunt. Ea excepteur nisi mollit reprehenderit eiusmod tempor. Eiusmod incididunt
    occaecat velit consectetur dolor cillum anim commodo fugiat cupidatat ut tempor officia. Aliquip fugiat occaecat
    excepteur consectetur ullamco consectetur exercitation occaecat sint sint incididunt cillum minim. Sint aliquip ea
    pariatur anim. Veniam laboris mollit in voluptate exercitation sint deserunt dolor ullamco ex dolor. Enim
    reprehenderit ut Lorem aliquip est laborum in. Aliqua in ut aute elit nulla amet. Ex proident pariatur ex in
    aliquip. Labore eu Lorem sint aliqua reprehenderit ipsum veniam aliquip laborum dolor deserunt cupidatat velit
    amet.`),
  },
  argTypes: {
    icon: {
      control: {
        type: 'select',
        options: getAvailableIcons(),
      },
    },
    title: {
      control: {
        type: 'text',
      },
    },
  },
} as Meta;

export const Basic: Story = ({ body, title, ...args }) => {
  return (
    <Modal title={title} {...args}>
      {body}
      <Modal.ButtonRow>
        <Button>Button1</Button>
        <Button variant="secondary">Cancel</Button>
      </Modal.ButtonRow>
    </Modal>
  );
};
Basic.args = {
  title: 'My Modal',
  icon: 'exclamation-triangle',
  isOpen: true,
  closeOnEscape: false,
  iconTooltip: 'icon tooltip',
};

const tabs = [
  { label: '1st child', value: 'first', active: true },
  { label: '2nd child', value: 'second', active: false },
  { label: '3rd child', value: 'third', active: false },
];

export const WithTabs: Story = (args) => {
  const [activeTab, setActiveTab] = useState('first');
  const modalHeader = (
    <ModalTabsHeader
      title={args.title}
      icon={args.icon}
      tabs={tabs}
      activeTab={activeTab}
      onChangeTab={(t) => {
        setActiveTab(t.value);
      }}
    />
  );
  return (
    <UseState initialState={tabs}>
      {(state, updateState) => {
        return (
          <div>
            <Modal title={modalHeader} isOpen={true}>
              <TabContent>
                {activeTab === state[0].value && <div>{args.body}</div>}
                {activeTab === state[1].value && <div>Second tab content</div>}
                {activeTab === state[2].value && <div>Third tab content</div>}
              </TabContent>
            </Modal>
          </div>
        );
      }}
    </UseState>
  );
};
WithTabs.args = {
  title: 'My Modal',
  icon: 'cog',
};

export const UsingContentClassName: Story = ({ title, body, ...args }) => {
  const override = {
    modalContent: css`
      background-color: darkorange;
    `,
  };
  return (
    <Modal title={title} {...args} contentClassName={cx(override.modalContent)}>
      {body}
    </Modal>
  );
};
UsingContentClassName.args = {
  title: 'Using contentClassName to override background',
  icon: 'exclamation-triangle',
  isOpen: true,
  closeOnEscape: false,
  iconTooltip: 'icon tooltip',
};
