import { dateTimeParse } from './parser';
import { systemDateFormats } from './formats';

describe('dateTimeParse', () => {
  it('should be able to parse using default format', () => {
    const date = dateTimeParse('2020-03-02 15:00:22', { timeZone: 'utc' });
    expect(date.format()).toEqual('2020-03-02T15:00:22Z');
  });

  it('should be able to parse using default format', () => {
    systemDateFormats.update({
      fullDate: 'MMMM D, YYYY, h:mm:ss a',
      interval: {} as any,
      useBrowserLocale: false,
    });

    const date = dateTimeParse('Aug 20, 2020 10:30:20 am', { timeZone: 'utc' });
    expect(date.format()).toEqual('2020-08-20T10:30:20Z');
  });

  it('should be able to parse array formats used by calendar', () => {
    const date = dateTimeParse([2020, 5, 10, 10, 30, 20], { timeZone: 'utc' });
    expect(date.format()).toEqual('2020-06-10T10:30:20Z');
  });
});
