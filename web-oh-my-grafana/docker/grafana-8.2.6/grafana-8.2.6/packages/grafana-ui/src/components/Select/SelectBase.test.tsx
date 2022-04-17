import React, { useState } from 'react';
import { render, screen, fireEvent } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { selectOptionInTest } from './test-utils';
import { SelectableValue } from '@grafana/data';
import { SelectBase } from './SelectBase';

describe('SelectBase', () => {
  const onChangeHandler = () => jest.fn();
  const options: Array<SelectableValue<number>> = [
    {
      label: 'Option 1',
      value: 1,
    },
    {
      label: 'Option 2',
      value: 2,
    },
  ];

  it('renders without error', () => {
    render(<SelectBase menuShouldPortal onChange={onChangeHandler} />);
  });

  it('renders empty options information', () => {
    render(<SelectBase menuShouldPortal onChange={onChangeHandler} />);
    userEvent.click(screen.getByText(/choose/i));
    expect(screen.queryByText(/no options found/i)).toBeVisible();
  });

  it('is selectable via its label text', async () => {
    render(
      <>
        <label htmlFor="my-select">My select</label>
        <SelectBase menuShouldPortal onChange={onChangeHandler} options={options} inputId="my-select" />
      </>
    );

    expect(screen.getByLabelText('My select')).toBeInTheDocument();
  });

  it('allows the value to be unset', async () => {
    const Test = () => {
      const option = { value: 'test-value', label: 'Test label' };
      const [value, setValue] = useState<SelectableValue<string> | null>(option);

      return (
        <>
          <button onClick={() => setValue(null)}>clear value</button>
          <SelectBase menuShouldPortal value={value} onChange={setValue} options={[option]} />
        </>
      );
    };

    render(<Test />);
    expect(screen.queryByText('Test label')).toBeInTheDocument();
    userEvent.click(screen.getByText('clear value'));
    expect(screen.queryByText('Test label')).not.toBeInTheDocument();
  });

  describe('when openMenuOnFocus prop', () => {
    describe('is provided', () => {
      it('opens on focus', () => {
        render(<SelectBase menuShouldPortal onChange={onChangeHandler} openMenuOnFocus />);
        fireEvent.focus(screen.getByRole('textbox'));
        expect(screen.queryByText(/no options found/i)).toBeVisible();
      });
    });
    describe('is not provided', () => {
      it.each`
        key
        ${'ArrowDown'}
        ${'ArrowUp'}
        ${' '}
      `('opens on arrow down/up or space', ({ key }) => {
        render(<SelectBase menuShouldPortal onChange={onChangeHandler} />);
        fireEvent.focus(screen.getByRole('textbox'));
        fireEvent.keyDown(screen.getByRole('textbox'), { key });
        expect(screen.queryByText(/no options found/i)).toBeVisible();
      });
    });
  });

  describe('when maxVisibleValues prop', () => {
    let excessiveOptions: Array<SelectableValue<number>> = [];
    beforeAll(() => {
      excessiveOptions = [
        {
          label: 'Option 1',
          value: 1,
        },
        {
          label: 'Option 2',
          value: 2,
        },
        {
          label: 'Option 3',
          value: 3,
        },
        {
          label: 'Option 4',
          value: 4,
        },
        {
          label: 'Option 5',
          value: 5,
        },
      ];
    });

    describe('is provided', () => {
      it('should only display maxVisibleValues options, and additional number of values should be displayed as indicator', () => {
        render(
          <SelectBase
            menuShouldPortal
            onChange={onChangeHandler}
            isMulti={true}
            maxVisibleValues={3}
            options={excessiveOptions}
            value={excessiveOptions}
            isOpen={false}
          />
        );
        expect(screen.queryAllByText(/option/i).length).toBe(3);
        expect(screen.queryByText(/\(\+2\)/i)).toBeVisible();
      });

      describe('and showAllSelectedWhenOpen prop is true', () => {
        it('should show all selected options when menu is open', () => {
          render(
            <SelectBase
              menuShouldPortal
              onChange={onChangeHandler}
              isMulti={true}
              maxVisibleValues={3}
              options={excessiveOptions}
              value={excessiveOptions}
              showAllSelectedWhenOpen={true}
              isOpen={true}
            />
          );

          expect(screen.queryAllByText(/option/i).length).toBe(5);
          expect(screen.queryByText(/\(\+2\)/i)).not.toBeInTheDocument();
        });
      });

      describe('and showAllSelectedWhenOpen prop is false', () => {
        it('should not show all selected options when menu is open', () => {
          render(
            <SelectBase
              menuShouldPortal
              onChange={onChangeHandler}
              isMulti={true}
              maxVisibleValues={3}
              value={excessiveOptions}
              options={excessiveOptions}
              showAllSelectedWhenOpen={false}
              isOpen={true}
            />
          );

          expect(screen.queryAllByText(/option/i).length).toBe(3);
          expect(screen.queryByText(/\(\+2\)/i)).toBeVisible();
        });
      });
    });

    describe('is not provided', () => {
      it('should always show all selected options', () => {
        render(
          <SelectBase
            menuShouldPortal
            onChange={onChangeHandler}
            isMulti={true}
            options={excessiveOptions}
            value={excessiveOptions}
            isOpen={false}
          />
        );

        expect(screen.queryAllByText(/option/i).length).toBe(5);
        expect(screen.queryByText(/\(\+2\)/i)).not.toBeInTheDocument();
      });
    });
  });

  describe('options', () => {
    it('renders menu with provided options', () => {
      render(<SelectBase menuShouldPortal options={options} onChange={onChangeHandler} />);
      userEvent.click(screen.getByText(/choose/i));
      const menuOptions = screen.getAllByLabelText('Select option');
      expect(menuOptions).toHaveLength(2);
    });

    it('call onChange handler when option is selected', async () => {
      const spy = jest.fn();

      render(<SelectBase menuShouldPortal onChange={spy} options={options} aria-label="My select" />);

      const selectEl = screen.getByLabelText('My select');
      expect(selectEl).toBeInTheDocument();

      await selectOptionInTest(selectEl, 'Option 2');
      expect(spy).toHaveBeenCalledWith({
        label: 'Option 2',
        value: 2,
      });
    });
  });
});
