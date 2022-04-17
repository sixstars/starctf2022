import { localTimeFormat } from './formats';

describe('Date Formats', () => {
  it('localTimeFormat', () => {
    const format = localTimeFormat(
      {
        year: '2-digit',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
      },
      ''
    );

    expect(format).toBe('MM/DD/YYYY, hh:mm:ss A');
  });
});

describe('Date Formats without hour12', () => {
  it('localTimeFormat', () => {
    const format = localTimeFormat(
      {
        year: '2-digit',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false,
      },
      ''
    );

    expect(format).toBe('MM/DD/YYYY, HH:mm:ss');
  });
});
