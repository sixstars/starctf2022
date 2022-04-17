import React from 'react';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { applyFieldOverrides, createTheme, DataFrame, FieldType, toDataFrame } from '@grafana/data';
import { Props, Table } from './Table';

function getDefaultDataFrame(): DataFrame {
  const dataFrame = toDataFrame({
    name: 'A',
    fields: [
      {
        name: 'time',
        type: FieldType.time,
        values: [1609459200000, 1609470000000, 1609462800000, 1609466400000],
        config: {
          custom: {
            filterable: false,
          },
        },
      },
      {
        name: 'temperature',
        type: FieldType.number,
        values: [10, NaN, 11, 12],
        config: {
          custom: {
            filterable: false,
          },
          links: [
            {
              targetBlank: true,
              title: 'Value link',
              url: '${__value.text}',
            },
          ],
        },
      },
      {
        name: 'img',
        type: FieldType.string,
        values: ['data:image/png;base64,1', 'data:image/png;base64,2', 'data:image/png;base64,3'],
        config: {
          custom: {
            filterable: false,
            displayMode: 'image',
          },
          links: [
            {
              targetBlank: true,
              title: 'Image link',
              url: '${__value.text}',
            },
          ],
        },
      },
    ],
  });
  const dataFrames = applyFieldOverrides({
    data: [dataFrame],
    fieldConfig: {
      defaults: {},
      overrides: [],
    },
    replaceVariables: (value, vars, format) => {
      return vars && value === '${__value.text}' ? vars['__value'].value.text : value;
    },
    timeZone: 'utc',
    theme: createTheme(),
  });
  return dataFrames[0];
}

function getTestContext(propOverrides: Partial<Props> = {}) {
  const onSortByChange = jest.fn();
  const onCellFilterAdded = jest.fn();
  const onColumnResize = jest.fn();
  const props: Props = {
    ariaLabel: 'aria-label',
    data: getDefaultDataFrame(),
    height: 600,
    width: 800,
    onSortByChange,
    onCellFilterAdded,
    onColumnResize,
  };

  Object.assign(props, propOverrides);
  const { rerender } = render(<Table {...props} />);

  return { rerender, onSortByChange, onCellFilterAdded, onColumnResize };
}

function getTable(): HTMLElement {
  return screen.getAllByRole('table')[0];
}

function getFooter(): HTMLElement {
  return screen.getByTestId('table-footer');
}

function getColumnHeader(name: string | RegExp): HTMLElement {
  return within(getTable()).getByRole('columnheader', { name });
}

function getLinks(row: HTMLElement): HTMLElement[] {
  return within(row).getAllByRole('link');
}

function getRowsData(rows: HTMLElement[]): Object[] {
  let content = [];
  for (let i = 1; i < rows.length; i++) {
    const row = getLinks(rows[i])[0];
    content.push({
      time: within(rows[i]).getByText(/2021*/).textContent,
      temperature: row.textContent,
      link: row.getAttribute('href'),
    });
  }
  return content;
}

describe('Table', () => {
  describe('when mounted without data', () => {
    it('then no data to show should be displayed', () => {
      getTestContext({ data: toDataFrame([]) });
      expect(getTable()).toBeInTheDocument();
      expect(screen.queryByRole('row')).not.toBeInTheDocument();
      expect(screen.getByText(/No data/i)).toBeInTheDocument();
    });
  });

  describe('when mounted with data', () => {
    it('then correct rows should be rendered', () => {
      getTestContext();
      expect(getTable()).toBeInTheDocument();
      expect(screen.getAllByRole('columnheader')).toHaveLength(3);
      expect(getColumnHeader(/time/)).toBeInTheDocument();
      expect(getColumnHeader(/temperature/)).toBeInTheDocument();
      expect(getColumnHeader(/img/)).toBeInTheDocument();

      const rows = within(getTable()).getAllByRole('row');
      expect(rows).toHaveLength(5);
      expect(getRowsData(rows)).toEqual([
        { time: '2021-01-01 00:00:00', temperature: '10', link: '10' },
        { time: '2021-01-01 03:00:00', temperature: 'NaN', link: 'NaN' },
        { time: '2021-01-01 01:00:00', temperature: '11', link: '11' },
        { time: '2021-01-01 02:00:00', temperature: '12', link: '12' },
      ]);
    });
  });

  describe('when mounted with footer', () => {
    it('then footer should be displayed', () => {
      const footerValues = ['a', 'b', 'c'];
      getTestContext({ footerValues });
      expect(getTable()).toBeInTheDocument();
      expect(getFooter()).toBeInTheDocument();
    });
  });

  describe('when sorting with column header', () => {
    it('then correct rows should be rendered', () => {
      getTestContext();

      userEvent.click(within(getColumnHeader(/temperature/)).getByText(/temperature/i));
      userEvent.click(within(getColumnHeader(/temperature/)).getByText(/temperature/i));

      const rows = within(getTable()).getAllByRole('row');
      expect(rows).toHaveLength(5);
      expect(getRowsData(rows)).toEqual([
        { time: '2021-01-01 02:00:00', temperature: '12', link: '12' },
        { time: '2021-01-01 01:00:00', temperature: '11', link: '11' },
        { time: '2021-01-01 00:00:00', temperature: '10', link: '10' },
        { time: '2021-01-01 03:00:00', temperature: 'NaN', link: 'NaN' },
      ]);
    });
  });
});
