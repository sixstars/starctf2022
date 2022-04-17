import { getDisplayProcessor, getRawDisplayProcessor } from './displayProcessor';
import { DisplayProcessor, DisplayValue } from '../types/displayValue';
import { MappingType, ValueMapping } from '../types/valueMapping';
import { FieldConfig, FieldType, ThresholdsMode } from '../types';
import { systemDateFormats } from '../datetime';
import { createTheme } from '../themes';

function getDisplayProcessorFromConfig(config: FieldConfig) {
  return getDisplayProcessor({
    field: {
      config,
      type: FieldType.number,
    },
    theme: createTheme(),
  });
}

function assertSame(input: any, processors: DisplayProcessor[], match: DisplayValue) {
  processors.forEach((processor) => {
    const value = processor(input);
    for (const key of Object.keys(match)) {
      expect((value as any)[key]).toEqual((match as any)[key]);
    }
  });
}

describe('Process simple display values', () => {
  // Don't test float values here since the decimal formatting changes
  const processors = [
    // Without options, this shortcuts to a much easier implementation
    getDisplayProcessor({ field: { config: {} }, theme: createTheme() }),

    // Add a simple option that is not used (uses a different base class)
    getDisplayProcessorFromConfig({ min: 0, max: 100 }),

    // Add a simple option that is not used (uses a different base class)
    getDisplayProcessorFromConfig({ unit: 'locale' }),
  ];

  it('support null', () => {
    assertSame(null, processors, { text: '', numeric: NaN });
  });

  it('support undefined', () => {
    assertSame(undefined, processors, { text: '', numeric: NaN });
  });

  it('support NaN', () => {
    assertSame(NaN, processors, { text: 'NaN', numeric: NaN });
  });

  it('Integer', () => {
    assertSame(3, processors, { text: '3', numeric: 3 });
  });

  it('Text to number', () => {
    assertSame('3', processors, { text: '3', numeric: 3 });
  });

  it('Empty string is NaN', () => {
    assertSame('', processors, { text: '', numeric: NaN });
  });

  it('Simple String', () => {
    assertSame('hello', processors, { text: 'hello', numeric: NaN });
  });

  it('empty array', () => {
    assertSame([], processors, { text: '', numeric: NaN });
  });

  it('array of text', () => {
    assertSame(['a', 'b', 'c'], processors, { text: 'a,b,c', numeric: NaN });
  });

  it('array of numbers', () => {
    assertSame([1, 2, 3], processors, { text: '1,2,3', numeric: NaN });
  });

  it('empty object', () => {
    assertSame({}, processors, { text: '[object Object]', numeric: NaN });
  });

  it('boolean true', () => {
    assertSame(true, processors, { text: 'true', numeric: 1 });
  });

  it('boolean false', () => {
    assertSame(false, processors, { text: 'false', numeric: 0 });
  });
});

describe('Process null values', () => {
  const processors = [
    getDisplayProcessorFromConfig({
      min: 0,
      max: 100,
      thresholds: {
        mode: ThresholdsMode.Absolute,
        steps: [
          { value: -Infinity, color: '#000' },
          { value: 0, color: '#100' },
          { value: 100, color: '#200' },
        ],
      },
    }),
  ];

  it('Null should get -Infinity (base) color', () => {
    assertSame(null, processors, { text: '', numeric: NaN, color: '#000' });
  });
});

describe('Format value', () => {
  it('should return if value isNaN', () => {
    const valueMappings: ValueMapping[] = [];
    const value = 'N/A';
    const instance = getDisplayProcessorFromConfig({ mappings: valueMappings });

    const result = instance(value);

    expect(result.text).toEqual('N/A');
  });

  it('should return formatted value if there are no value mappings', () => {
    const valueMappings: ValueMapping[] = [];
    const value = '6';

    const instance = getDisplayProcessorFromConfig({ decimals: 1, mappings: valueMappings });

    const result = instance(value);

    expect(result.text).toEqual('6.0');
  });

  it('should return formatted value if there are no matching value mappings', () => {
    const valueMappings: ValueMapping[] = [
      { type: MappingType.ValueToText, options: { '11': { text: 'elva' } } },
      { type: MappingType.RangeToText, options: { from: 1, to: 9, result: { text: '1-9' } } },
    ];

    const instance = getDisplayProcessorFromConfig({ decimals: 1, mappings: valueMappings });
    const result = instance('10');

    expect(result.text).toEqual('10.0');
  });

  it('should return mapped value if there are matching value mappings', () => {
    const valueMappings: ValueMapping[] = [
      { type: MappingType.ValueToText, options: { '11': { text: 'elva' } } },
      { type: MappingType.RangeToText, options: { from: 1, to: 9, result: { text: '1-9' } } },
    ];

    const instance = getDisplayProcessorFromConfig({ decimals: 1, mappings: valueMappings });
    const result = instance('11');

    expect(result.text).toEqual('elva');
  });

  it('should return mapped color but use value format if no value mapping text specified', () => {
    const valueMappings: ValueMapping[] = [
      { type: MappingType.RangeToText, options: { from: 1, to: 9, result: { color: '#FFF' } } },
    ];

    const instance = getDisplayProcessorFromConfig({ decimals: 2, mappings: valueMappings });
    const result = instance(5);

    expect(result.color).toEqual('#FFF');
    expect(result.text).toEqual('5.00');
  });

  it('should replace a matching regex', () => {
    const valueMappings: ValueMapping[] = [
      { type: MappingType.RegexToText, options: { pattern: '([^.]*).example.com', result: { text: '$1' } } },
    ];

    const instance = getDisplayProcessorFromConfig({ decimals: 1, mappings: valueMappings });
    const result = instance('hostname.example.com');

    expect(result.text).toEqual('hostname');
  });

  it('should not replace a non-matching regex', () => {
    const valueMappings: ValueMapping[] = [
      { type: MappingType.RegexToText, options: { pattern: '([^.]*).example.com', result: { text: '$1' } } },
    ];

    const instance = getDisplayProcessorFromConfig({ decimals: 1, mappings: valueMappings });
    const result = instance('hostname.acme.com');

    expect(result.text).toEqual('hostname.acme.com');
  });

  it('should empty a matching regex without replacement', () => {
    const valueMappings: ValueMapping[] = [
      { type: MappingType.RegexToText, options: { pattern: '([^.]*).example.com', result: { text: '' } } },
    ];

    const instance = getDisplayProcessorFromConfig({ decimals: 1, mappings: valueMappings });
    const result = instance('hostname.example.com');

    expect(result.text).toEqual('');
  });

  it('should not empty a non-matching regex', () => {
    const valueMappings: ValueMapping[] = [
      { type: MappingType.RegexToText, options: { pattern: '([^.]*).example.com', result: { text: '' } } },
    ];

    const instance = getDisplayProcessorFromConfig({ decimals: 1, mappings: valueMappings });
    const result = instance('hostname.acme.com');

    expect(result.text).toEqual('hostname.acme.com');
  });

  it('should return value with color if mapping has color', () => {
    const valueMappings: ValueMapping[] = [{ type: MappingType.ValueToText, options: { Low: { color: 'red' } } }];

    const instance = getDisplayProcessorFromConfig({ decimals: 1, mappings: valueMappings });
    const result = instance('Low');

    expect(result.text).toEqual('Low');
    expect(result.color).toEqual('#F2495C');
  });

  it('should return mapped value and leave numeric value in tact if value mapping maps to empty string', () => {
    const valueMappings: ValueMapping[] = [{ type: MappingType.ValueToText, options: { '1': { text: '' } } }];
    const value = '1';
    const instance = getDisplayProcessorFromConfig({ decimals: 1, mappings: valueMappings });

    expect(instance(value).text).toEqual('');
    expect(instance(value).numeric).toEqual(1);
  });

  it('should not map 1kW to the value for 1W', () => {
    const valueMappings: ValueMapping[] = [{ type: MappingType.ValueToText, options: { '1': { text: 'mapped' } } }];
    const value = '1000';
    const instance = getDisplayProcessorFromConfig({ decimals: 1, mappings: valueMappings, unit: 'watt' });

    const result = instance(value);

    expect(result.text).toEqual('1.0');
  });

  it('With null value and thresholds should use base color', () => {
    const instance = getDisplayProcessorFromConfig({
      thresholds: {
        mode: ThresholdsMode.Absolute,
        steps: [{ value: -Infinity, color: '#AAA' }],
      },
    });
    const disp = instance(null);
    expect(disp.text).toEqual('');
    expect(disp.color).toEqual('#AAA');
  });

  //
  // Below is current behavior but it's clearly not working great
  //

  it('with value 1000 and unit short', () => {
    const value = 1000;
    const instance = getDisplayProcessorFromConfig({ decimals: null, unit: 'short' });
    const disp = instance(value);
    expect(disp.text).toEqual('1');
    expect(disp.suffix).toEqual(' K');
  });

  it('with value 1200 and unit short', () => {
    const value = 1200;
    const instance = getDisplayProcessorFromConfig({ decimals: null, unit: 'short' });
    const disp = instance(value);
    expect(disp.text).toEqual('1.20');
    expect(disp.suffix).toEqual(' K');
  });

  it('with value 1250 and unit short', () => {
    const value = 1250;
    const instance = getDisplayProcessorFromConfig({ decimals: null, unit: 'short' });
    const disp = instance(value);
    expect(disp.text).toEqual('1.25');
    expect(disp.suffix).toEqual(' K');
  });

  it('with value 10000000 and unit short', () => {
    const value = 1000000;
    const instance = getDisplayProcessorFromConfig({ decimals: null, unit: 'short' });
    const disp = instance(value);
    expect(disp.text).toEqual('1');
    expect(disp.suffix).toEqual(' Mil');
  });

  it('with value 15000000 and unit short', () => {
    const value = 1500000;
    const instance = getDisplayProcessorFromConfig({ decimals: null, unit: 'short' });
    const disp = instance(value);
    expect(disp.text).toEqual('1.50');
    expect(disp.suffix).toEqual(' Mil');
  });

  it('with value 128000000 and unit bytes', () => {
    const value = 1280000125;
    const instance = getDisplayProcessorFromConfig({ decimals: null, unit: 'bytes' });
    const disp = instance(value);
    expect(disp.text).toEqual('1.19');
    expect(disp.suffix).toEqual(' GiB');
  });
});

describe('Date display options', () => {
  it('should format UTC dates', () => {
    const processor = getDisplayProcessor({
      timeZone: 'utc',
      field: {
        type: FieldType.time,
        config: {
          unit: 'xyz', // ignore non-date formats
        },
      },
      theme: createTheme(),
    });
    expect(processor(0).text).toEqual('1970-01-01 00:00:00');
  });

  it('should pick configured time format', () => {
    const processor = getDisplayProcessor({
      timeZone: 'utc',
      field: {
        type: FieldType.time,
        config: {
          unit: 'dateTimeAsUS', // ignore non-date formats
        },
      },
      theme: createTheme(),
    });
    expect(processor(0).text).toEqual('01/01/1970 12:00:00 am');
  });

  it('respect the configured date format', () => {
    const processor = getDisplayProcessor({
      timeZone: 'utc',
      field: {
        type: FieldType.time,
        config: {
          unit: 'time:YYYY', // ignore non-date formats
        },
      },
      theme: createTheme(),
    });
    expect(processor(0).text).toEqual('1970');
  });

  it('Should use system date format by default', () => {
    const currentFormat = systemDateFormats.fullDate;
    systemDateFormats.fullDate = 'YYYY-MM';

    const processor = getDisplayProcessor({
      timeZone: 'utc',
      field: {
        type: FieldType.time,
        config: {},
      },
      theme: createTheme(),
    });

    expect(processor(0).text).toEqual('1970-01');

    systemDateFormats.fullDate = currentFormat;
  });

  it('should handle ISO string dates', () => {
    const processor = getDisplayProcessor({
      timeZone: 'utc',
      field: {
        type: FieldType.time,
        config: {},
      },
      theme: createTheme(),
    });

    expect(processor('2020-08-01T08:48:43.783337Z').text).toEqual('2020-08-01 08:48:43');
  });

  describe('number formatting for string values', () => {
    it('should preserve string unchanged if unit is strings', () => {
      const processor = getDisplayProcessor({
        field: {
          type: FieldType.string,
          config: { unit: 'string' },
        },
        theme: createTheme(),
      });
      expect(processor('22.1122334455').text).toEqual('22.1122334455');
    });

    it('should format string as number if no unit', () => {
      const processor = getDisplayProcessor({
        field: {
          type: FieldType.string,
          config: { decimals: 2 },
        },
        theme: createTheme(),
      });
      expect(processor('22.1122334455').text).toEqual('22.11');

      // Support empty/missing strings
      expect(processor(undefined).text).toEqual('');
      expect(processor(null).text).toEqual('');
      expect(processor('').text).toEqual('');
    });
  });
});

describe('getRawDisplayProcessor', () => {
  const processor = getRawDisplayProcessor();
  const date = new Date('2020-01-01T00:00:00.000Z');
  const timestamp = date.valueOf();

  it.each`
    value                             | expected
    ${0}                              | ${'0'}
    ${13.37}                          | ${'13.37'}
    ${true}                           | ${'true'}
    ${false}                          | ${'false'}
    ${date}                           | ${`${date}`}
    ${timestamp}                      | ${'1577836800000'}
    ${'a string'}                     | ${'a string'}
    ${null}                           | ${'null'}
    ${undefined}                      | ${'undefined'}
    ${{ value: 0, label: 'a label' }} | ${'[object Object]'}
  `('when called with value:{$value}', ({ value, expected }) => {
    const result = processor(value);

    expect(result).toEqual({ text: expected, numeric: null });
  });
});
