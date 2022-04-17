import React from 'react';
import { fireEvent, render, RenderResult } from '@testing-library/react';
import { dateTimeParse, TimeRange } from '@grafana/data';
import { TimeRangeForm } from './TimeRangeForm';
import { selectors } from '@grafana/e2e-selectors';

type TimeRangeFormRenderResult = RenderResult & {
  getCalendarDayByLabelText(label: string): HTMLButtonElement;
};

const defaultTimeRange: TimeRange = {
  from: dateTimeParse('2021-06-17 00:00:00', { timeZone: 'utc' }),
  to: dateTimeParse('2021-06-19 23:59:00', { timeZone: 'utc' }),
  raw: {
    from: '2021-06-17 00:00:00',
    to: '2021-06-19 23:59:00',
  },
};

function setup(initial: TimeRange = defaultTimeRange, timeZone = 'utc'): TimeRangeFormRenderResult {
  const result = render(<TimeRangeForm isFullscreen={true} value={initial} onApply={() => {}} timeZone={timeZone} />);

  return {
    ...result,
    getCalendarDayByLabelText: (label: string) => {
      const item = result.getByLabelText(label);
      return item?.parentElement as HTMLButtonElement;
    },
  };
}

describe('TimeRangeForm', () => {
  it('should render form correcty', () => {
    const { getByLabelText, getByText } = setup();
    const { TimePicker } = selectors.components;

    expect(getByText('Apply time range')).toBeInTheDocument();
    expect(getByLabelText(TimePicker.fromField)).toBeInTheDocument();
    expect(getByLabelText(TimePicker.toField)).toBeInTheDocument();
  });

  it('should display calendar when clicking the from input field', () => {
    const { getByLabelText } = setup();
    const { TimePicker } = selectors.components;

    fireEvent.focus(getByLabelText(TimePicker.fromField));
    expect(getByLabelText(TimePicker.calendar)).toBeInTheDocument();
  });

  it('should have passed time range entered in form', () => {
    const { getByLabelText } = setup();
    const { TimePicker } = selectors.components;

    const fromValue = defaultTimeRange.raw.from as string;
    const toValue = defaultTimeRange.raw.to as string;

    expect(getByLabelText(TimePicker.fromField)).toHaveValue(fromValue);
    expect(getByLabelText(TimePicker.toField)).toHaveValue(toValue);
  });

  it('should display calendar when clicking the to input field', () => {
    const { getByLabelText } = setup();
    const { TimePicker } = selectors.components;

    fireEvent.focus(getByLabelText(TimePicker.toField));
    expect(getByLabelText(TimePicker.calendar)).toBeInTheDocument();
  });

  it('should not display calendar without clicking any input field', () => {
    const { queryByLabelText } = setup();
    const { TimePicker } = selectors.components;

    expect(queryByLabelText(TimePicker.calendar)).toBeNull();
  });

  it('should have passed time range selected in calendar', () => {
    const { getByLabelText, getCalendarDayByLabelText } = setup();
    const { TimePicker } = selectors.components;

    fireEvent.focus(getByLabelText(TimePicker.toField));
    const from = getCalendarDayByLabelText('June 17, 2021');
    const to = getCalendarDayByLabelText('June 19, 2021');

    expect(from).toHaveClass('react-calendar__tile--rangeStart');
    expect(to).toHaveClass('react-calendar__tile--rangeEnd');
  });

  it('should select correct time range in calendar when having a custom time zone', () => {
    const { getByLabelText, getCalendarDayByLabelText } = setup(defaultTimeRange, 'Asia/Tokyo');
    const { TimePicker } = selectors.components;

    fireEvent.focus(getByLabelText(TimePicker.toField));
    const from = getCalendarDayByLabelText('June 17, 2021');
    const to = getCalendarDayByLabelText('June 19, 2021');

    expect(from).toHaveClass('react-calendar__tile--rangeStart');
    expect(to).toHaveClass('react-calendar__tile--rangeEnd');
  });

  describe('dates error handling', () => {
    it('should show error on invalid dates', () => {
      const invalidTimeRange: TimeRange = {
        from: dateTimeParse('foo', { timeZone: 'utc' }),
        to: dateTimeParse('2021-06-19 23:59:00', { timeZone: 'utc' }),
        raw: {
          from: 'foo',
          to: '2021-06-19 23:59:00',
        },
      };
      const { getAllByRole } = setup(invalidTimeRange, 'Asia/Tokyo');
      const error = getAllByRole('alert');

      expect(error).toHaveLength(1);
      expect(error[0]).toBeVisible();
      expect(error[0]).toHaveTextContent('Please enter a past date or "now"');
    });

    it('should show error on invalid range', () => {
      const invalidTimeRange: TimeRange = {
        from: dateTimeParse('2021-06-19 00:00:00', { timeZone: 'utc' }),
        to: dateTimeParse('2021-06-17 23:59:00', { timeZone: 'utc' }),
        raw: {
          from: '2021-06-19 00:00:00',
          to: '2021-06-17 23:59:00',
        },
      };
      const { getAllByRole } = setup(invalidTimeRange, 'Asia/Tokyo');
      const error = getAllByRole('alert');

      expect(error[0]).toBeVisible();
      expect(error[0]).toHaveTextContent('"From" can\'t be after "To"');
    });

    it('should not show range error when "to" is invalid', () => {
      const invalidTimeRange: TimeRange = {
        from: dateTimeParse('2021-06-19 00:00:00', { timeZone: 'utc' }),
        to: dateTimeParse('foo', { timeZone: 'utc' }),
        raw: {
          from: '2021-06-19 00:00:00',
          to: 'foo',
        },
      };
      const { getAllByRole } = setup(invalidTimeRange, 'Asia/Tokyo');
      const error = getAllByRole('alert');

      expect(error).toHaveLength(1);
      expect(error[0]).toBeVisible();
      expect(error[0]).toHaveTextContent('Please enter a past date or "now"');
    });
  });
});
