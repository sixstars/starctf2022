import {
  applyFieldOverrides,
  applyRawFieldOverrides,
  FieldOverrideEnv,
  findNumericFieldMinMax,
  getLinksSupplier,
  setDynamicConfigValue,
  setFieldConfigDefaults,
} from './fieldOverrides';
import { ArrayDataFrame, MutableDataFrame, toDataFrame } from '../dataframe';
import {
  DataFrame,
  Field,
  FieldColorModeId,
  FieldConfig,
  FieldConfigPropertyItem,
  FieldConfigSource,
  FieldType,
  InterpolateFunction,
  ScopedVars,
  ThresholdsMode,
} from '../types';
import { locationUtil, Registry } from '../utils';
import { mockStandardProperties } from '../utils/tests/mockStandardProperties';
import { FieldMatcherID } from '../transformations';
import { FieldConfigOptionsRegistry } from './FieldConfigOptionsRegistry';
import { getFieldDisplayName } from './fieldState';
import { ArrayVector } from '../vector';
import { getDisplayProcessor } from './displayProcessor';
import { createTheme } from '../themes';

const property1: any = {
  id: 'custom.property1', // Match field properties
  path: 'property1', // Match field properties
  isCustom: true,
  process: (value: any) => value,
  shouldApply: () => true,
};

const property2 = {
  id: 'custom.property2', // Match field properties
  path: 'property2', // Match field properties
  isCustom: true,
  process: (value: any) => value,
  shouldApply: () => true,
};

const property3: any = {
  id: 'custom.property3.nested', // Match field properties
  path: 'property3.nested', // Match field properties
  isCustom: true,
  process: (value: any) => value,
  shouldApply: () => true,
};

const shouldApplyFalse: any = {
  id: 'custom.shouldApplyFalse', // Match field properties
  path: 'shouldApplyFalse', // Match field properties
  isCustom: true,
  process: (value: any) => value,
  shouldApply: () => false,
};

export const customFieldRegistry: FieldConfigOptionsRegistry = new Registry<FieldConfigPropertyItem>(() => {
  return [property1, property2, property3, shouldApplyFalse, ...mockStandardProperties()];
});

locationUtil.initialize({
  config: { appSubUrl: '/subUrl' } as any,
  getVariablesUrlParams: (() => {}) as any,
  getTimeRangeForUrl: (() => {}) as any,
});

describe('Global MinMax', () => {
  it('find global min max', () => {
    const f0 = new ArrayDataFrame<{ title: string; value: number; value2: number | null }>([
      { title: 'AAA', value: 100, value2: 1234 },
      { title: 'BBB', value: -20, value2: null },
      { title: 'CCC', value: 200, value2: 1000 },
    ]);

    const minmax = findNumericFieldMinMax([f0]);
    expect(minmax.min).toEqual(-20);
    expect(minmax.max).toEqual(1234);
  });

  it('find global min max when all values are zero', () => {
    const f0 = new ArrayDataFrame<{ title: string; value: number; value2: number | null }>([
      { title: 'AAA', value: 0, value2: 0 },
      { title: 'CCC', value: 0, value2: 0 },
    ]);

    const minmax = findNumericFieldMinMax([f0]);
    expect(minmax.min).toEqual(0);
    expect(minmax.max).toEqual(0);
  });

  describe('when value is null', () => {
    it('then global min max should be null', () => {
      const frame = toDataFrame({
        fields: [
          { name: 'Time', type: FieldType.time, values: [1] },
          { name: 'Value', type: FieldType.number, values: [null] },
        ],
      });
      const { min, max } = findNumericFieldMinMax([frame]);

      expect(min).toBe(null);
      expect(max).toBe(null);
    });
  });

  describe('when value values are zeo', () => {
    it('then global min max should be correct', () => {
      const frame = toDataFrame({
        fields: [
          { name: 'Time', type: FieldType.time, values: [1, 2] },
          { name: 'Value', type: FieldType.number, values: [1, 2] },
        ],
      });
      const frame2 = toDataFrame({
        fields: [
          { name: 'Time', type: FieldType.time, values: [1, 2] },
          { name: 'Value', type: FieldType.number, values: [0, 0] },
        ],
      });

      const { min, max } = findNumericFieldMinMax([frame, frame2]);

      expect(min).toBe(0);
      expect(max).toBe(2);
    });
  });
});

describe('applyFieldOverrides', () => {
  const f0 = new ArrayDataFrame<{ title: string; value: number; value2: number | null }>([
    { title: 'AAA', value: 100, value2: 1234 },
    { title: 'BBB', value: -20, value2: null },
    { title: 'CCC', value: 200, value2: 1000 },
  ]);

  // Hardcode the max value
  f0.fields[1].config.max = 0;
  f0.fields[1].config.decimals = 6;
  f0.fields[1].config.custom = { value: 1 };

  const src: FieldConfigSource = {
    defaults: {
      unit: 'xyz',
      decimals: 2,
      links: [{ title: 'link', url: '${__value.text}' }],
    },
    overrides: [
      {
        matcher: { id: FieldMatcherID.numeric },
        properties: [
          { id: 'decimals', value: 1 }, // Numeric
          { id: 'displayName', value: 'Kittens' }, // Text
        ],
      },
    ],
  };

  describe('given multiple data frames', () => {
    const f0 = new MutableDataFrame({
      name: 'A',
      fields: [{ name: 'message', type: FieldType.string, values: [10, 20] }],
    });
    const f1 = new MutableDataFrame({
      name: 'B',
      fields: [{ name: 'info', type: FieldType.string, values: [10, 20] }],
    });

    it('should add scopedVars to fields', () => {
      const withOverrides = applyFieldOverrides({
        data: [f0, f1],
        fieldConfig: {
          defaults: {},
          overrides: [],
        },
        replaceVariables: (value: any) => value,
        theme: createTheme(),
        fieldConfigRegistry: new FieldConfigOptionsRegistry(),
      });

      expect(withOverrides[0].fields[0].state!.scopedVars).toMatchInlineSnapshot(`
                                                                                 Object {
                                                                                   "__field": Object {
                                                                                     "text": "Field",
                                                                                     "value": Object {},
                                                                                   },
                                                                                   "__series": Object {
                                                                                     "text": "Series",
                                                                                     "value": Object {
                                                                                       "name": "A",
                                                                                     },
                                                                                   },
                                                                                 }
                                                                                 `);

      expect(withOverrides[1].fields[0].state!.scopedVars).toMatchInlineSnapshot(`
        Object {
          "__field": Object {
            "text": "Field",
            "value": Object {},
          },
          "__series": Object {
            "text": "Series",
            "value": Object {
              "name": "B",
            },
          },
        }
      `);
    });
  });

  it('will merge FieldConfig with default values', () => {
    const field: FieldConfig = {
      min: 0,
      max: 100,
    };

    const f1 = {
      unit: 'ms',
      dateFormat: '', // should be ignored
      max: parseFloat('NOPE'), // should be ignored
      min: null, // should alo be ignored!
      displayName: 'newTitle',
    };

    const f: DataFrame = toDataFrame({
      fields: [{ type: FieldType.number, name: 'x', config: field, values: [] }],
    });

    const processed = applyFieldOverrides({
      data: [f],
      fieldConfig: {
        defaults: f1 as FieldConfig,
        overrides: [],
      },
      fieldConfigRegistry: customFieldRegistry,
      replaceVariables: (v) => v,
      theme: createTheme(),
    })[0];

    const outField = processed.fields[0];

    expect(outField.config.min).toEqual(0);
    expect(outField.config.max).toEqual(100);
    expect(outField.config.unit).toEqual('ms');
    expect(getFieldDisplayName(outField, f)).toEqual('newTitle');
  });

  it('will apply field overrides', () => {
    const data = applyFieldOverrides({
      data: [f0], // the frame
      fieldConfig: src as FieldConfigSource, // defaults + overrides
      replaceVariables: (undefined as any) as InterpolateFunction,
      theme: createTheme(),
      fieldConfigRegistry: customFieldRegistry,
    })[0];
    const valueColumn = data.fields[1];
    const config = valueColumn.config;

    // Keep max from the original setting
    expect(config.max).toEqual(0);

    // Don't Automatically pick the min value
    expect(config.min).toEqual(undefined);

    // The default value applied
    expect(config.unit).toEqual('xyz');

    // The default value applied
    expect(config.displayName).toEqual('Kittens');

    // The override applied
    expect(config.decimals).toEqual(1);
  });

  it('will apply set min/max when asked', () => {
    const data = applyFieldOverrides({
      data: [f0], // the frame
      fieldConfig: src as FieldConfigSource, // defaults + overrides
      replaceVariables: (undefined as any) as InterpolateFunction,
      theme: createTheme(),
    })[0];
    const valueColumn = data.fields[1];
    const range = valueColumn.state!.range!;

    // Keep max from the original setting
    expect(range.max).toEqual(0);

    // Don't Automatically pick the min value
    expect(range.min).toEqual(-20);
  });

  it('getLinks should use applied field config', () => {
    const replaceVariablesCalls: any[] = [];

    const data = applyFieldOverrides({
      data: [f0], // the frame
      fieldConfig: src as FieldConfigSource, // defaults + overrides
      replaceVariables: ((value: string, variables: ScopedVars) => {
        replaceVariablesCalls.push(variables);
        return value;
      }) as InterpolateFunction,
      theme: createTheme(),
      fieldConfigRegistry: customFieldRegistry,
    })[0];

    data.fields[1].getLinks!({ valueRowIndex: 0 });

    expect(data.fields[1].config.decimals).toEqual(1);
    expect(replaceVariablesCalls[3].__value.value.text).toEqual('100.0');
  });

  it('creates a deep clone of field config', () => {
    const data = applyFieldOverrides({
      data: [f0], // the frame
      fieldConfig: src as FieldConfigSource, // defaults + overrides
      replaceVariables: (undefined as any) as InterpolateFunction,
      theme: createTheme(),
    })[0];

    expect(data.fields[1].config).not.toBe(f0.fields[1].config);
    expect(data.fields[1].config.custom).not.toBe(f0.fields[1].config.custom);
  });

  it('Can modify displayName and have it affect later overrides', () => {
    const config: FieldConfigSource = {
      defaults: {},
      overrides: [
        {
          matcher: { id: FieldMatcherID.byName, options: 'value' },
          properties: [{ id: 'displayName', value: 'Kittens' }],
        },
        {
          matcher: { id: FieldMatcherID.byName, options: 'Kittens' },
          properties: [{ id: 'displayName', value: 'Kittens improved' }],
        },
      ],
    };

    const data = applyFieldOverrides({
      data: [f0], // the frame
      fieldConfig: config,
      replaceVariables: (str: string) => str,
      fieldConfigRegistry: customFieldRegistry,
      theme: createTheme(),
    })[0];

    expect(data.fields[1].config.displayName).toBe('Kittens improved');
    expect(getFieldDisplayName(data.fields[1], data)).toBe('Kittens improved');
  });
});

describe('setFieldConfigDefaults', () => {
  it('applies field config defaults', () => {
    const dsFieldConfig: FieldConfig = {
      decimals: 2,
      min: 0,
      max: 100,
    };

    const panelFieldConfig: FieldConfig = {
      decimals: 1,
      min: 10,
      max: 50,
      unit: 'km',
    };

    const context: FieldOverrideEnv = {
      data: [] as any,
      field: { type: FieldType.number } as any,
      dataFrameIndex: 0,
      fieldConfigRegistry: customFieldRegistry,
    };

    // we mutate dsFieldConfig
    setFieldConfigDefaults(dsFieldConfig, panelFieldConfig, context);

    expect(dsFieldConfig).toMatchInlineSnapshot(`
      Object {
        "custom": Object {},
        "decimals": 2,
        "max": 100,
        "min": 0,
        "unit": "km",
      }
    `);
  });

  it('applies field config defaults for custom properties', () => {
    const dsFieldConfig: FieldConfig = {
      custom: {
        property1: 10,
      },
    };

    const panelFieldConfig: FieldConfig = {
      custom: {
        property1: 20,
        property2: 10,
      },
    };

    const context: FieldOverrideEnv = {
      data: [] as any,
      field: { type: FieldType.number } as any,
      dataFrameIndex: 0,
      fieldConfigRegistry: customFieldRegistry,
    };

    // we mutate dsFieldConfig
    setFieldConfigDefaults(dsFieldConfig, panelFieldConfig, context);

    expect(dsFieldConfig).toMatchInlineSnapshot(`
      Object {
        "custom": Object {
          "property1": 10,
          "property2": 10,
        },
      }
    `);
  });
});

describe('setDynamicConfigValue', () => {
  it('applies dynamic config values', () => {
    const config = {
      displayName: 'test',
    };

    setDynamicConfigValue(
      config,
      {
        id: 'displayName',
        value: 'applied',
      },
      {
        fieldConfigRegistry: customFieldRegistry,
        data: [] as any,
        field: { type: FieldType.number } as any,
        dataFrameIndex: 0,
      }
    );

    expect(config.displayName).toEqual('applied');
  });

  it('applies custom dynamic config values', () => {
    const config = {
      custom: {
        property1: 1,
      },
    };
    setDynamicConfigValue(
      config,
      {
        id: 'custom.property1',
        value: 'applied',
      },
      {
        fieldConfigRegistry: customFieldRegistry,
        data: [] as any,
        field: { type: FieldType.number } as any,
        dataFrameIndex: 0,
      }
    );

    expect(config.custom.property1).toEqual('applied');
  });

  it('applies overrides even when shouldApply returns false', () => {
    const config: FieldConfig = {
      custom: {},
    };
    setDynamicConfigValue(
      config,
      {
        id: 'custom.shouldApplyFalse',
        value: 'applied',
      },
      {
        fieldConfigRegistry: customFieldRegistry,
        data: [] as any,
        field: { type: FieldType.number } as any,
        dataFrameIndex: 0,
      }
    );

    expect(config.custom.shouldApplyFalse).toEqual('applied');
  });

  it('applies nested custom dynamic config values', () => {
    const config = {
      custom: {
        property3: {
          nested: 1,
        },
      },
    };
    setDynamicConfigValue(
      config,
      {
        id: 'custom.property3.nested',
        value: 'applied',
      },
      {
        fieldConfigRegistry: customFieldRegistry,
        data: [] as any,
        field: { type: FieldType.number } as any,
        dataFrameIndex: 0,
      }
    );

    expect(config.custom.property3.nested).toEqual('applied');
  });

  it('removes properties', () => {
    const config = {
      displayName: 'title',
      custom: {
        property3: {
          nested: 1,
        },
      },
    };
    setDynamicConfigValue(
      config,
      {
        id: 'custom.property3.nested',
        value: undefined,
      },
      {
        fieldConfigRegistry: customFieldRegistry,
        data: [] as any,
        field: { type: FieldType.number } as any,
        dataFrameIndex: 0,
      }
    );

    setDynamicConfigValue(
      config,
      {
        id: 'displayName',
        value: undefined,
      },
      {
        fieldConfigRegistry: customFieldRegistry,
        data: [] as any,
        field: { type: FieldType.number } as any,
        dataFrameIndex: 0,
      }
    );

    expect(config.custom.property3).toEqual({});
    expect(config.displayName).toBeUndefined();
  });
});

describe('getLinksSupplier', () => {
  it('will replace variables in url and title of the data link', () => {
    locationUtil.initialize({
      config: {} as any,
      getVariablesUrlParams: (() => {}) as any,
      getTimeRangeForUrl: (() => {}) as any,
    });

    const f0 = new MutableDataFrame({
      name: 'A',
      fields: [
        {
          name: 'message',
          type: FieldType.string,
          values: [10, 20],
          config: {
            links: [
              {
                url: 'url to be interpolated',
                title: 'title to be interpolated',
              },
            ],
          },
        },
      ],
    });

    const replaceSpy = jest.fn();
    const supplier = getLinksSupplier(f0, f0.fields[0], {}, replaceSpy);
    supplier({});

    expect(replaceSpy).toBeCalledTimes(2);
    expect(replaceSpy.mock.calls[0][0]).toEqual('url to be interpolated');
    expect(replaceSpy.mock.calls[1][0]).toEqual('title to be interpolated');
  });

  it('handles internal links', () => {
    locationUtil.initialize({
      config: { appSubUrl: '' } as any,
      getVariablesUrlParams: (() => {}) as any,
      getTimeRangeForUrl: (() => {}) as any,
    });

    const f0 = new MutableDataFrame({
      name: 'A',
      fields: [
        {
          name: 'message',
          type: FieldType.string,
          values: [10, 20],
          config: {
            links: [
              {
                url: '',
                title: '',
                internal: {
                  datasourceUid: '0',
                  datasourceName: 'testDS',
                  query: '12345',
                },
              },
            ],
          },
          display: (v) => ({ numeric: v, text: String(v) }),
        },
      ],
    });

    const supplier = getLinksSupplier(
      f0,
      f0.fields[0],
      {},
      // We do not need to interpolate anything for this test
      (value, vars, format) => value
    );

    const links = supplier({ valueRowIndex: 0 });

    expect(links.length).toBe(1);
    expect(links[0]).toEqual(
      expect.objectContaining({
        title: 'testDS',
        href: `/explore?left=${encodeURIComponent('{"datasource":"testDS","queries":["12345"]}')}`,
        onClick: undefined,
      })
    );
  });
});

describe('applyRawFieldOverrides', () => {
  const getNumberFieldConfig = () => ({
    custom: {},
    thresholds: {
      mode: ThresholdsMode.Absolute,
      steps: [
        {
          color: 'green',
          value: (null as unknown) as number,
        },
        {
          color: 'red',
          value: 80,
        },
      ],
    },
    mappings: [],
    color: {
      mode: FieldColorModeId.Thresholds,
    },
    min: 0,
    max: 1599124316808,
  });

  const getEmptyConfig = () => ({
    custom: {},
    mappings: [],
  });

  const getDisplayValue = (frames: DataFrame[], frameIndex: number, fieldIndex: number) => {
    const field = frames[frameIndex].fields[fieldIndex];
    const value = field.values.get(0);
    return field.display!(value);
  };

  const expectRawDataDisplayValue = (frames: DataFrame[], frameIndex: number) => {
    expect(getDisplayValue(frames, frameIndex, 0)).toEqual({ text: '1599045551050', numeric: null });
    expect(getDisplayValue(frames, frameIndex, 1)).toEqual({ text: '3.14159265359', numeric: null });
    expect(getDisplayValue(frames, frameIndex, 2)).toEqual({ text: '0', numeric: null });
    expect(getDisplayValue(frames, frameIndex, 3)).toEqual({ text: '0', numeric: null });
    expect(getDisplayValue(frames, frameIndex, 4)).toEqual({ text: 'A - string', numeric: null });
    expect(getDisplayValue(frames, frameIndex, 5)).toEqual({ text: '1599045551050', numeric: null });
  };

  const expectFormattedDataDisplayValue = (frames: DataFrame[], frameIndex: number) => {
    expect(getDisplayValue(frames, frameIndex, 0)).toEqual({
      color: '#F2495C',
      numeric: 1599045551050,
      prefix: undefined,
      suffix: undefined,
      text: '1599045551050',
      percent: expect.any(Number),
    });

    expect(getDisplayValue(frames, frameIndex, 1)).toEqual({
      color: '#73BF69',
      numeric: 3.14159265359,
      percent: expect.any(Number),
      prefix: undefined,
      suffix: undefined,
      text: '3.142',
    });

    expect(getDisplayValue(frames, frameIndex, 2)).toEqual({
      color: '#73BF69',
      numeric: 0,
      percent: expect.any(Number),
      prefix: undefined,
      suffix: undefined,
      text: '0',
    });

    expect(getDisplayValue(frames, frameIndex, 3)).toEqual({
      color: '#F2495C', // red
      numeric: 0,
      percent: expect.any(Number),
      prefix: undefined,
      suffix: undefined,
      text: 'False',
    });

    expect(getDisplayValue(frames, frameIndex, 4)).toEqual({
      color: '#73BF69', // value from classic pallet
      numeric: NaN,
      percent: 1,
      prefix: undefined,
      suffix: undefined,
      text: 'A - string',
    });

    expect(getDisplayValue(frames, frameIndex, 5)).toEqual({
      color: '#808080',
      numeric: 1599045551050,
      percent: expect.any(Number),
      prefix: undefined,
      suffix: undefined,
      text: '2020-09-02 11:19:11',
    });
  };

  describe('when called', () => {
    it('then all fields should have their display processor replaced with the raw display processor', () => {
      const numberAsEpoc: Field = {
        name: 'numberAsEpoc',
        type: FieldType.number,
        values: new ArrayVector([1599045551050]),
        config: getNumberFieldConfig(),
      };

      const numberWithDecimals: Field = {
        name: 'numberWithDecimals',
        type: FieldType.number,
        values: new ArrayVector([3.14159265359]),
        config: {
          ...getNumberFieldConfig(),
          decimals: 3,
        },
      };

      const numberAsBoolean: Field = {
        name: 'numberAsBoolean',
        type: FieldType.number,
        values: new ArrayVector([0]),
        config: getNumberFieldConfig(),
      };

      const boolean: Field = {
        name: 'boolean',
        type: FieldType.boolean,
        values: new ArrayVector([0]),
        config: getEmptyConfig(),
      };

      const string: Field = {
        name: 'string',
        type: FieldType.boolean,
        values: new ArrayVector(['A - string']),
        config: getEmptyConfig(),
      };

      const datetime: Field = {
        name: 'datetime',
        type: FieldType.time,
        values: new ArrayVector([1599045551050]),
        config: {
          unit: 'dateTimeAsIso',
        },
      };

      const dataFrameA: DataFrame = toDataFrame({
        fields: [numberAsEpoc, numberWithDecimals, numberAsBoolean, boolean, string, datetime],
      });

      const theme = createTheme();

      dataFrameA.fields[0].display = getDisplayProcessor({ field: dataFrameA.fields[0], theme });
      dataFrameA.fields[1].display = getDisplayProcessor({ field: dataFrameA.fields[1], theme });
      dataFrameA.fields[2].display = getDisplayProcessor({ field: dataFrameA.fields[2], theme });
      dataFrameA.fields[3].display = getDisplayProcessor({ field: dataFrameA.fields[3], theme });
      dataFrameA.fields[4].display = getDisplayProcessor({ field: dataFrameA.fields[4], theme });
      dataFrameA.fields[5].display = getDisplayProcessor({ field: dataFrameA.fields[5], theme, timeZone: 'utc' });

      const dataFrameB: DataFrame = toDataFrame({
        fields: [numberAsEpoc, numberWithDecimals, numberAsBoolean, boolean, string, datetime],
      });

      dataFrameB.fields[0].display = getDisplayProcessor({ field: dataFrameB.fields[0], theme });
      dataFrameB.fields[1].display = getDisplayProcessor({ field: dataFrameB.fields[1], theme });
      dataFrameB.fields[2].display = getDisplayProcessor({ field: dataFrameB.fields[2], theme });
      dataFrameB.fields[3].display = getDisplayProcessor({ field: dataFrameB.fields[3], theme });
      dataFrameB.fields[4].display = getDisplayProcessor({ field: dataFrameB.fields[4], theme });
      dataFrameB.fields[5].display = getDisplayProcessor({ field: dataFrameB.fields[5], theme, timeZone: 'utc' });

      const data = [dataFrameA, dataFrameB];
      const rawData = applyRawFieldOverrides(data);

      // expect raw data is correct
      expectRawDataDisplayValue(rawData, 0);
      expectRawDataDisplayValue(rawData, 1);

      // expect the original data is still the same
      expectFormattedDataDisplayValue(data, 0);
      expectFormattedDataDisplayValue(data, 1);
    });
  });
});
