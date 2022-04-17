import Plain from 'slate-plain-serializer';
import { Editor as SlateEditor } from 'slate';
import LanguageProvider from './language_provider';
import { PrometheusDatasource } from './datasource';
import { HistoryItem } from '@grafana/data';
import { PromQuery } from './types';
import Mock = jest.Mock;
import { SearchFunctionType } from '@grafana/ui';

describe('Language completion provider', () => {
  const datasource: PrometheusDatasource = ({
    metadataRequest: () => ({ data: { data: [] as any[] } }),
    getTimeRangeParams: () => ({ start: '0', end: '1' }),
  } as any) as PrometheusDatasource;

  describe('cleanText', () => {
    const cleanText = new LanguageProvider(datasource).cleanText;
    it('does not remove metric or label keys', () => {
      expect(cleanText('foo')).toBe('foo');
      expect(cleanText('foo_bar')).toBe('foo_bar');
    });

    it('keeps trailing space but removes leading', () => {
      expect(cleanText('foo ')).toBe('foo ');
      expect(cleanText(' foo')).toBe('foo');
    });

    it('removes label syntax', () => {
      expect(cleanText('foo="bar')).toBe('bar');
      expect(cleanText('foo!="bar')).toBe('bar');
      expect(cleanText('foo=~"bar')).toBe('bar');
      expect(cleanText('foo!~"bar')).toBe('bar');
      expect(cleanText('{bar')).toBe('bar');
    });

    it('removes previous operators', () => {
      expect(cleanText('foo + bar')).toBe('bar');
      expect(cleanText('foo+bar')).toBe('bar');
      expect(cleanText('foo - bar')).toBe('bar');
      expect(cleanText('foo * bar')).toBe('bar');
      expect(cleanText('foo / bar')).toBe('bar');
      expect(cleanText('foo % bar')).toBe('bar');
      expect(cleanText('foo ^ bar')).toBe('bar');
      expect(cleanText('foo and bar')).toBe('bar');
      expect(cleanText('foo or bar')).toBe('bar');
      expect(cleanText('foo unless bar')).toBe('bar');
      expect(cleanText('foo == bar')).toBe('bar');
      expect(cleanText('foo != bar')).toBe('bar');
      expect(cleanText('foo > bar')).toBe('bar');
      expect(cleanText('foo < bar')).toBe('bar');
      expect(cleanText('foo >= bar')).toBe('bar');
      expect(cleanText('foo <= bar')).toBe('bar');
      expect(cleanText('memory')).toBe('memory');
    });

    it('removes aggregation syntax', () => {
      expect(cleanText('(bar')).toBe('bar');
      expect(cleanText('(foo,bar')).toBe('bar');
      expect(cleanText('(foo, bar')).toBe('bar');
    });

    it('removes range syntax', () => {
      expect(cleanText('[1m')).toBe('1m');
    });
  });

  describe('fetchSeries', () => {
    it('should use match[] parameter', () => {
      const languageProvider = new LanguageProvider(datasource);
      const fetchSeries = languageProvider.fetchSeries;
      const requestSpy = jest.spyOn(languageProvider, 'request');
      fetchSeries('{job="grafana"}');
      expect(requestSpy).toHaveBeenCalled();
      expect(requestSpy).toHaveBeenCalledWith(
        '/api/v1/series',
        {},
        { end: '1', 'match[]': '{job="grafana"}', start: '0' }
      );
    });
  });

  describe('empty query suggestions', () => {
    it('returns no suggestions on empty context', async () => {
      const instance = new LanguageProvider(datasource);
      const value = Plain.deserialize('');
      const result = await instance.provideCompletionItems({ text: '', prefix: '', value, wrapperClasses: [] });
      expect(result.context).toBeUndefined();
      expect(result.suggestions).toMatchObject([]);
    });

    it('returns no suggestions with metrics on empty context even when metrics were provided', async () => {
      const instance = new LanguageProvider(datasource);
      instance.metrics = ['foo', 'bar'];
      const value = Plain.deserialize('');
      const result = await instance.provideCompletionItems({ text: '', prefix: '', value, wrapperClasses: [] });
      expect(result.context).toBeUndefined();
      expect(result.suggestions).toMatchObject([]);
    });

    it('returns history on empty context when history was provided', async () => {
      const instance = new LanguageProvider(datasource);
      const value = Plain.deserialize('');
      const history: Array<HistoryItem<PromQuery>> = [
        {
          ts: 0,
          query: { refId: '1', expr: 'metric' },
        },
      ];
      const result = await instance.provideCompletionItems(
        { text: '', prefix: '', value, wrapperClasses: [] },
        { history }
      );
      expect(result.context).toBeUndefined();

      expect(result.suggestions).toMatchObject([
        {
          label: 'History',
          items: [
            {
              label: 'metric',
            },
          ],
        },
      ]);
    });
  });

  describe('range suggestions', () => {
    it('returns range suggestions in range context', async () => {
      const instance = new LanguageProvider(datasource);
      const value = Plain.deserialize('1');
      const result = await instance.provideCompletionItems({
        text: '1',
        prefix: '1',
        value,
        wrapperClasses: ['context-range'],
      });
      expect(result.context).toBe('context-range');
      expect(result.suggestions).toMatchObject([
        {
          items: [
            { label: '$__interval', sortValue: '$__interval' },
            { label: '$__rate_interval', sortValue: '$__rate_interval' },
            { label: '$__range', sortValue: '$__range' },
            { label: '1m', sortValue: '00:01:00' },
            { label: '5m', sortValue: '00:05:00' },
            { label: '10m', sortValue: '00:10:00' },
            { label: '30m', sortValue: '00:30:00' },
            { label: '1h', sortValue: '01:00:00' },
            { label: '1d', sortValue: '24:00:00' },
          ],
          label: 'Range vector',
        },
      ]);
    });
  });

  describe('metric suggestions', () => {
    it('returns history, metrics and function suggestions in an uknown context ', async () => {
      const instance = new LanguageProvider(datasource);
      instance.metrics = ['foo', 'bar'];
      const history: Array<HistoryItem<PromQuery>> = [
        {
          ts: 0,
          query: { refId: '1', expr: 'metric' },
        },
      ];
      let value = Plain.deserialize('m');
      value = value.setSelection({ anchor: { offset: 1 }, focus: { offset: 1 } });
      // Even though no metric with `m` is present, we still get metric completion items, filtering is done by the consumer
      const result = await instance.provideCompletionItems(
        { text: 'm', prefix: 'm', value, wrapperClasses: [] },
        { history }
      );
      expect(result.context).toBeUndefined();
      expect(result.suggestions).toMatchObject([
        {
          label: 'History',
          items: [
            {
              label: 'metric',
            },
          ],
        },
        {
          label: 'Functions',
        },
        {
          label: 'Metrics',
        },
      ]);
    });

    it('returns no suggestions directly after a binary operator', async () => {
      const instance = new LanguageProvider(datasource);
      instance.metrics = ['foo', 'bar'];
      const value = Plain.deserialize('*');
      const result = await instance.provideCompletionItems({ text: '*', prefix: '', value, wrapperClasses: [] });
      expect(result.context).toBeUndefined();
      expect(result.suggestions).toMatchObject([]);
    });

    it('returns metric suggestions with prefix after a binary operator', async () => {
      const instance = new LanguageProvider(datasource);
      instance.metrics = ['foo', 'bar'];
      const value = Plain.deserialize('foo + b');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(7).value;
      const result = await instance.provideCompletionItems({
        text: 'foo + b',
        prefix: 'b',
        value: valueWithSelection,
        wrapperClasses: [],
      });
      expect(result.context).toBeUndefined();
      expect(result.suggestions).toMatchObject([
        {
          label: 'Functions',
        },
        {
          label: 'Metrics',
        },
      ]);
    });

    it('returns no suggestions at the beginning of a non-empty function', async () => {
      const instance = new LanguageProvider(datasource);
      const value = Plain.deserialize('sum(up)');
      const ed = new SlateEditor({ value });

      const valueWithSelection = ed.moveForward(4).value;
      const result = await instance.provideCompletionItems({
        text: '',
        prefix: '',
        value: valueWithSelection,
        wrapperClasses: [],
      });
      expect(result.context).toBeUndefined();
      expect(result.suggestions.length).toEqual(0);
    });
  });

  describe('label suggestions', () => {
    it('returns default label suggestions on label context and no metric', async () => {
      const instance = new LanguageProvider(datasource);
      const value = Plain.deserialize('{}');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(1).value;
      const result = await instance.provideCompletionItems({
        text: '',
        prefix: '',
        wrapperClasses: ['context-labels'],
        value: valueWithSelection,
      });
      expect(result.context).toBe('context-labels');
      expect(result.suggestions).toEqual([
        {
          items: [{ label: 'job' }, { label: 'instance' }],
          label: 'Labels',
          searchFunctionType: SearchFunctionType.Fuzzy,
        },
      ]);
    });

    it('returns label suggestions on label context and metric', async () => {
      const datasources: PrometheusDatasource = ({
        metadataRequest: () => ({ data: { data: [{ __name__: 'metric', bar: 'bazinga' }] as any[] } }),
        getTimeRangeParams: () => ({ start: '0', end: '1' }),
      } as any) as PrometheusDatasource;
      const instance = new LanguageProvider(datasources);
      const value = Plain.deserialize('metric{}');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(7).value;
      const result = await instance.provideCompletionItems({
        text: '',
        prefix: '',
        wrapperClasses: ['context-labels'],
        value: valueWithSelection,
      });
      expect(result.context).toBe('context-labels');
      expect(result.suggestions).toEqual([
        { items: [{ label: 'bar' }], label: 'Labels', searchFunctionType: SearchFunctionType.Fuzzy },
      ]);
    });

    it('returns label suggestions on label context but leaves out labels that already exist', async () => {
      const datasource: PrometheusDatasource = ({
        metadataRequest: () => ({
          data: {
            data: [
              {
                __name__: 'metric',
                bar: 'asdasd',
                job1: 'dsadsads',
                job2: 'fsfsdfds',
                job3: 'dsadsad',
              },
            ],
          },
        }),
        getTimeRangeParams: () => ({ start: '0', end: '1' }),
      } as any) as PrometheusDatasource;
      const instance = new LanguageProvider(datasource);
      const value = Plain.deserialize('{job1="foo",job2!="foo",job3=~"foo",__name__="metric",}');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(54).value;
      const result = await instance.provideCompletionItems({
        text: '',
        prefix: '',
        wrapperClasses: ['context-labels'],
        value: valueWithSelection,
      });
      expect(result.context).toBe('context-labels');
      expect(result.suggestions).toEqual([
        { items: [{ label: 'bar' }], label: 'Labels', searchFunctionType: SearchFunctionType.Fuzzy },
      ]);
    });

    it('returns label value suggestions inside a label value context after a negated matching operator', async () => {
      const instance = new LanguageProvider(({
        ...datasource,
        metadataRequest: () => {
          return { data: { data: ['value1', 'value2'] } };
        },
      } as any) as PrometheusDatasource);
      const value = Plain.deserialize('{job!=}');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(6).value;
      const result = await instance.provideCompletionItems({
        text: '!=',
        prefix: '',
        wrapperClasses: ['context-labels'],
        labelKey: 'job',
        value: valueWithSelection,
      });
      expect(result.context).toBe('context-label-values');
      expect(result.suggestions).toEqual([
        {
          items: [{ label: 'value1' }, { label: 'value2' }],
          label: 'Label values for "job"',
          searchFunctionType: SearchFunctionType.Fuzzy,
        },
      ]);
    });

    it('returns a refresher on label context and unavailable metric', async () => {
      const instance = new LanguageProvider(datasource);
      const value = Plain.deserialize('metric{}');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(7).value;
      const result = await instance.provideCompletionItems({
        text: '',
        prefix: '',
        wrapperClasses: ['context-labels'],
        value: valueWithSelection,
      });
      expect(result.context).toBeUndefined();
      expect(result.suggestions).toEqual([]);
    });

    it('returns label values on label context when given a metric and a label key', async () => {
      const instance = new LanguageProvider(({
        ...datasource,
        metadataRequest: () => simpleMetricLabelsResponse,
      } as any) as PrometheusDatasource);
      const value = Plain.deserialize('metric{bar=ba}');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(13).value;
      const result = await instance.provideCompletionItems({
        text: '=ba',
        prefix: 'ba',
        wrapperClasses: ['context-labels'],
        labelKey: 'bar',
        value: valueWithSelection,
      });
      expect(result.context).toBe('context-label-values');
      expect(result.suggestions).toEqual([
        { items: [{ label: 'baz' }], label: 'Label values for "bar"', searchFunctionType: SearchFunctionType.Fuzzy },
      ]);
    });

    it('returns label suggestions on aggregation context and metric w/ selector', async () => {
      const instance = new LanguageProvider(({
        ...datasource,
        metadataRequest: () => simpleMetricLabelsResponse,
      } as any) as PrometheusDatasource);
      const value = Plain.deserialize('sum(metric{foo="xx"}) by ()');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(26).value;
      const result = await instance.provideCompletionItems({
        text: '',
        prefix: '',
        wrapperClasses: ['context-aggregation'],
        value: valueWithSelection,
      });
      expect(result.context).toBe('context-aggregation');
      expect(result.suggestions).toEqual([
        { items: [{ label: 'bar' }], label: 'Labels', searchFunctionType: SearchFunctionType.Fuzzy },
      ]);
    });

    it('returns label suggestions on aggregation context and metric w/o selector', async () => {
      const instance = new LanguageProvider(({
        ...datasource,
        metadataRequest: () => simpleMetricLabelsResponse,
      } as any) as PrometheusDatasource);
      const value = Plain.deserialize('sum(metric) by ()');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(16).value;
      const result = await instance.provideCompletionItems({
        text: '',
        prefix: '',
        wrapperClasses: ['context-aggregation'],
        value: valueWithSelection,
      });
      expect(result.context).toBe('context-aggregation');
      expect(result.suggestions).toEqual([
        { items: [{ label: 'bar' }], label: 'Labels', searchFunctionType: SearchFunctionType.Fuzzy },
      ]);
    });

    it('returns label suggestions inside a multi-line aggregation context', async () => {
      const instance = new LanguageProvider(({
        ...datasource,
        metadataRequest: () => simpleMetricLabelsResponse,
      } as any) as PrometheusDatasource);
      const value = Plain.deserialize('sum(\nmetric\n)\nby ()');
      const aggregationTextBlock = value.document.getBlocks().get(3);
      const ed = new SlateEditor({ value });
      ed.moveToStartOfNode(aggregationTextBlock);
      const valueWithSelection = ed.moveForward(4).value;
      const result = await instance.provideCompletionItems({
        text: '',
        prefix: '',
        wrapperClasses: ['context-aggregation'],
        value: valueWithSelection,
      });
      expect(result.context).toBe('context-aggregation');
      expect(result.suggestions).toEqual([
        {
          items: [{ label: 'bar' }],
          label: 'Labels',
          searchFunctionType: SearchFunctionType.Fuzzy,
        },
      ]);
    });

    it('returns label suggestions inside an aggregation context with a range vector', async () => {
      const instance = new LanguageProvider(({
        ...datasource,
        metadataRequest: () => simpleMetricLabelsResponse,
      } as any) as PrometheusDatasource);
      const value = Plain.deserialize('sum(rate(metric[1h])) by ()');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(26).value;
      const result = await instance.provideCompletionItems({
        text: '',
        prefix: '',
        wrapperClasses: ['context-aggregation'],
        value: valueWithSelection,
      });
      expect(result.context).toBe('context-aggregation');
      expect(result.suggestions).toEqual([
        {
          items: [{ label: 'bar' }],
          label: 'Labels',
          searchFunctionType: SearchFunctionType.Fuzzy,
        },
      ]);
    });

    it('returns label suggestions inside an aggregation context with a range vector and label', async () => {
      const instance = new LanguageProvider(({
        ...datasource,
        metadataRequest: () => simpleMetricLabelsResponse,
      } as any) as PrometheusDatasource);
      const value = Plain.deserialize('sum(rate(metric{label1="value"}[1h])) by ()');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(42).value;
      const result = await instance.provideCompletionItems({
        text: '',
        prefix: '',
        wrapperClasses: ['context-aggregation'],
        value: valueWithSelection,
      });
      expect(result.context).toBe('context-aggregation');
      expect(result.suggestions).toEqual([
        {
          items: [{ label: 'bar' }],
          label: 'Labels',
          searchFunctionType: SearchFunctionType.Fuzzy,
        },
      ]);
    });

    it('returns no suggestions inside an unclear aggregation context using alternate syntax', async () => {
      const instance = new LanguageProvider(datasource);
      const value = Plain.deserialize('sum by ()');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(8).value;
      const result = await instance.provideCompletionItems({
        text: '',
        prefix: '',
        wrapperClasses: ['context-aggregation'],
        value: valueWithSelection,
      });
      expect(result.context).toBe('context-aggregation');
      expect(result.suggestions).toEqual([]);
    });

    it('returns label suggestions inside an aggregation context using alternate syntax', async () => {
      const instance = new LanguageProvider(({
        ...datasource,
        metadataRequest: () => simpleMetricLabelsResponse,
      } as any) as PrometheusDatasource);
      const value = Plain.deserialize('sum by () (metric)');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(8).value;
      const result = await instance.provideCompletionItems({
        text: '',
        prefix: '',
        wrapperClasses: ['context-aggregation'],
        value: valueWithSelection,
      });
      expect(result.context).toBe('context-aggregation');
      expect(result.suggestions).toEqual([
        {
          items: [{ label: 'bar' }],
          label: 'Labels',
          searchFunctionType: SearchFunctionType.Fuzzy,
        },
      ]);
    });

    it('does not re-fetch default labels', async () => {
      const datasource: PrometheusDatasource = ({
        metadataRequest: jest.fn(() => ({ data: { data: [] as any[] } })),
        getTimeRangeParams: jest.fn(() => ({ start: '0', end: '1' })),
      } as any) as PrometheusDatasource;

      const instance = new LanguageProvider(datasource);
      const value = Plain.deserialize('{}');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(1).value;
      const args = {
        text: '',
        prefix: '',
        wrapperClasses: ['context-labels'],
        value: valueWithSelection,
      };
      const promise1 = instance.provideCompletionItems(args);
      // one call for 2 default labels job, instance
      expect((datasource.metadataRequest as Mock).mock.calls.length).toBe(2);
      const promise2 = instance.provideCompletionItems(args);
      expect((datasource.metadataRequest as Mock).mock.calls.length).toBe(2);
      await Promise.all([promise1, promise2]);
      expect((datasource.metadataRequest as Mock).mock.calls.length).toBe(2);
    });
  });
  describe('disabled metrics lookup', () => {
    it('does not issue any metadata requests when lookup is disabled', async () => {
      const datasource: PrometheusDatasource = ({
        metadataRequest: jest.fn(() => ({ data: { data: ['foo', 'bar'] as string[] } })),
        getTimeRangeParams: jest.fn(() => ({ start: '0', end: '1' })),
        lookupsDisabled: true,
      } as any) as PrometheusDatasource;
      const instance = new LanguageProvider(datasource);
      const value = Plain.deserialize('{}');
      const ed = new SlateEditor({ value });
      const valueWithSelection = ed.moveForward(1).value;
      const args = {
        text: '',
        prefix: '',
        wrapperClasses: ['context-labels'],
        value: valueWithSelection,
      };

      expect((datasource.metadataRequest as Mock).mock.calls.length).toBe(0);
      await instance.start();
      expect((datasource.metadataRequest as Mock).mock.calls.length).toBe(0);
      await instance.provideCompletionItems(args);
      expect((datasource.metadataRequest as Mock).mock.calls.length).toBe(0);
    });
    it('issues metadata requests when lookup is not disabled', async () => {
      const datasource: PrometheusDatasource = ({
        metadataRequest: jest.fn(() => ({ data: { data: ['foo', 'bar'] as string[] } })),
        getTimeRangeParams: jest.fn(() => ({ start: '0', end: '1' })),
        lookupsDisabled: false,
      } as any) as PrometheusDatasource;
      const instance = new LanguageProvider(datasource);

      expect((datasource.metadataRequest as Mock).mock.calls.length).toBe(0);
      await instance.start();
      expect((datasource.metadataRequest as Mock).mock.calls.length).toBeGreaterThan(0);
    });
  });
});

const simpleMetricLabelsResponse = {
  data: {
    data: [
      {
        __name__: 'metric',
        bar: 'baz',
      },
    ],
  },
};
