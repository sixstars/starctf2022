import React from 'react';
import { Meta, Story } from '@storybook/react';
import { action } from '@storybook/addon-actions';
import { withCenteredStory } from '../../utils/storybook/withCenteredStory';
import mdx from './CodeEditor.mdx';
import { CodeEditor } from './CodeEditorLazy';

export default {
  title: 'CodeEditor',
  component: CodeEditor,
  decorators: [withCenteredStory],
  parameters: {
    docs: {
      page: mdx,
    },
    controls: {
      exclude: ['monacoOptions', 'onEditorDidMount', 'onBlur', 'onSave', 'getSuggestions', 'showLineNumbers'],
    },
  },
  argTypes: {
    width: { control: { type: 'range', min: 100, max: 500, step: 10 } },
    height: { control: { type: 'range', min: 100, max: 800, step: 10 } },
    language: { control: { type: 'select' }, options: ['sql', 'json'] },
  },
} as Meta;

export const Basic: Story = (args) => {
  return (
    <CodeEditor
      width={args.width}
      height={args.height}
      value={args.value}
      language={args.language}
      onBlur={(text: string) => {
        action('code blur')(text);
      }}
      onSave={(text: string) => {
        action('code saved')(text);
      }}
      showLineNumbers={args.showLineNumbers}
      showMiniMap={args.showMiniMap}
      readOnly={args.readOnly}
    />
  );
};
Basic.args = {
  width: 300,
  height: 400,
  value: `CREATE TABLE Persons (
  PersonID int,
  LastName varchar(255),
  FirstName varchar(255),
  Address varchar(255),
  City varchar(255)
);
SELECT * FROM Persons LIMIT 10 '
  `,
  language: 'sql',
  showLineNumbers: false,
  showMiniMap: false,
  readOnly: false,
};
