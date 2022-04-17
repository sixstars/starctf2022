import { DataTransformerID } from './ids';
import { toDataFrame } from '../../dataframe/processDataFrame';
import { FieldType } from '../../types/dataFrame';
import { mockTransformationsRegistry } from '../../utils/tests/mockTransformationsRegistry';
import { filterFieldsByNameTransformer } from './filterByName';
import { filterFieldsTransformer } from './filter';
import { transformDataFrame } from '../transformDataFrame';

export const seriesWithNamesToMatch = toDataFrame({
  fields: [
    { name: 'startsWithA', type: FieldType.time, values: [1000, 2000] },
    { name: 'B', type: FieldType.boolean, values: [true, false] },
    { name: 'startsWithC', type: FieldType.string, values: ['a', 'b'] },
    { name: 'D', type: FieldType.number, values: [1, 2] },
  ],
});

describe('filterByName transformer', () => {
  beforeAll(() => {
    mockTransformationsRegistry([filterFieldsByNameTransformer, filterFieldsTransformer]);
  });

  it('returns original series if no options provided', async () => {
    const cfg = {
      id: DataTransformerID.filterFields,
      options: {},
    };

    await expect(transformDataFrame([cfg], [seriesWithNamesToMatch])).toEmitValuesWith((received) => {
      const data = received[0];
      const filtered = data[0];
      expect(filtered.fields.length).toBe(4);
    });
  });

  describe('respects', () => {
    it('inclusion by pattern', async () => {
      const cfg = {
        id: DataTransformerID.filterFieldsByName,
        options: {
          include: {
            pattern: '/^(startsWith)/',
          },
        },
      };

      await expect(transformDataFrame([cfg], [seriesWithNamesToMatch])).toEmitValuesWith((received) => {
        const data = received[0];
        const filtered = data[0];
        expect(filtered.fields.length).toBe(2);
        expect(filtered.fields[0].name).toBe('startsWithA');
      });
    });

    it('exclusion by pattern', async () => {
      const cfg = {
        id: DataTransformerID.filterFieldsByName,
        options: {
          exclude: {
            pattern: '/^(startsWith)/',
          },
        },
      };

      await expect(transformDataFrame([cfg], [seriesWithNamesToMatch])).toEmitValuesWith((received) => {
        const data = received[0];
        const filtered = data[0];
        expect(filtered.fields.length).toBe(2);
        expect(filtered.fields[0].name).toBe('B');
      });
    });

    it('inclusion and exclusion by pattern', async () => {
      const cfg = {
        id: DataTransformerID.filterFieldsByName,
        options: {
          exclude: { pattern: '/^(startsWith)/' },
          include: { pattern: '/^(B)$/' },
        },
      };

      await expect(transformDataFrame([cfg], [seriesWithNamesToMatch])).toEmitValuesWith((received) => {
        const data = received[0];
        const filtered = data[0];
        expect(filtered.fields.length).toBe(1);
        expect(filtered.fields[0].name).toBe('B');
      });
    });

    it('inclusion by names', async () => {
      const cfg = {
        id: DataTransformerID.filterFieldsByName,
        options: {
          include: {
            names: ['startsWithA', 'startsWithC'],
          },
        },
      };

      await expect(transformDataFrame([cfg], [seriesWithNamesToMatch])).toEmitValuesWith((received) => {
        const data = received[0];
        const filtered = data[0];
        expect(filtered.fields.length).toBe(2);
        expect(filtered.fields[0].name).toBe('startsWithA');
      });
    });

    it('exclusion by names', async () => {
      const cfg = {
        id: DataTransformerID.filterFieldsByName,
        options: {
          exclude: {
            names: ['startsWithA', 'startsWithC'],
          },
        },
      };

      await expect(transformDataFrame([cfg], [seriesWithNamesToMatch])).toEmitValuesWith((received) => {
        const data = received[0];
        const filtered = data[0];
        expect(filtered.fields.length).toBe(2);
        expect(filtered.fields[0].name).toBe('B');
      });
    });

    it('inclusion and exclusion by names', async () => {
      const cfg = {
        id: DataTransformerID.filterFieldsByName,
        options: {
          exclude: { names: ['startsWithA', 'startsWithC'] },
          include: { names: ['B'] },
        },
      };

      await expect(transformDataFrame([cfg], [seriesWithNamesToMatch])).toEmitValuesWith((received) => {
        const data = received[0];
        const filtered = data[0];
        expect(filtered.fields.length).toBe(1);
        expect(filtered.fields[0].name).toBe('B');
      });
    });

    it('inclusion by both', async () => {
      const cfg = {
        id: DataTransformerID.filterFieldsByName,
        options: {
          include: {
            pattern: '/^(startsWith)/',
            names: ['startsWithA'],
          },
        },
      };

      await expect(transformDataFrame([cfg], [seriesWithNamesToMatch])).toEmitValuesWith((received) => {
        const data = received[0];
        const filtered = data[0];
        expect(filtered.fields.length).toBe(2);
        expect(filtered.fields[0].name).toBe('startsWithA');
      });
    });

    it('exclusion by both', async () => {
      const cfg = {
        id: DataTransformerID.filterFieldsByName,
        options: {
          exclude: {
            pattern: '/^(startsWith)/',
            names: ['startsWithA'],
          },
        },
      };

      await expect(transformDataFrame([cfg], [seriesWithNamesToMatch])).toEmitValuesWith((received) => {
        const data = received[0];
        const filtered = data[0];
        expect(filtered.fields.length).toBe(2);
        expect(filtered.fields[0].name).toBe('B');
      });
    });

    it('inclusion and exclusion by both', async () => {
      const cfg = {
        id: DataTransformerID.filterFieldsByName,
        options: {
          exclude: { names: ['startsWithA', 'startsWithC'] },
          include: { pattern: '/^(B)$/' },
        },
      };

      await expect(transformDataFrame([cfg], [seriesWithNamesToMatch])).toEmitValuesWith((received) => {
        const data = received[0];
        const filtered = data[0];
        expect(filtered.fields.length).toBe(1);
        expect(filtered.fields[0].name).toBe('B');
      });
    });
  });
});
