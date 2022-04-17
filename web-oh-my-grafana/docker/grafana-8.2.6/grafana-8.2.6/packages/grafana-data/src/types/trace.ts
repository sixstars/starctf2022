/**
 * Type representing a tag in a trace span or fields of a log.
 */
export type TraceKeyValuePair<T = any> = {
  key: string;
  value: T;
};

/**
 * Type representing a log in a span.
 */
export type TraceLog = {
  // Millisecond epoch time
  timestamp: number;
  fields: TraceKeyValuePair[];
};

/**
 * This describes the structure of the dataframe that should be returned from a tracing data source to show trace
 * in a TraceView component.
 */
export interface TraceSpanRow {
  traceID: string;
  spanID: string;
  parentSpanID: string | undefined;
  operationName: string;
  serviceName: string;
  serviceTags: TraceKeyValuePair[];
  // Millisecond epoch time
  startTime: number;
  // Milliseconds
  duration: number;
  logs?: TraceLog[];

  // Note: To mark spen as having error add tag error: true
  tags?: TraceKeyValuePair[];
  warnings?: string[];
  stackTraces?: string[];

  // Specify custom color of the error icon
  errorIconColor?: string;
}
