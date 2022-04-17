import { Observable } from 'rxjs';

/**
 * Used to initiate a remote call via the {@link BackendSrv}
 *
 * @public
 */
export type BackendSrvRequest = {
  /**
   * Request URL
   */
  url: string;

  /**
   * Number of times to retry the remote call if it fails.
   */
  retry?: number;

  /**
   * HTTP headers that should be passed along with the remote call.
   * Please have a look at {@link https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API | Fetch API}
   * for supported headers.
   */
  headers?: Record<string, any>;

  /**
   * HTTP verb to perform in the remote call GET, POST, PUT etc.
   */
  method?: string;

  /**
   * Set to false an success application alert box will not be shown for successful PUT, DELETE, POST requests
   */
  showSuccessAlert?: boolean;

  /**
   * Set to false to not show an application alert box for request errors
   */
  showErrorAlert?: boolean;

  /**
   * Provided by the initiator to identify a particular remote call. An example
   * of this is when a datasource plugin triggers a query. If the request id already
   * exist the backendSrv will try to cancel and replace the previous call with the
   * new one.
   */
  requestId?: string;

  /**
   * Set to to true to not include call in query inspector
   */
  hideFromInspector?: boolean;

  /**
   * The data to send
   */
  data?: any;

  /**
   * Query params
   */
  params?: Record<string, any>;

  /**
   * Define how the response object should be parsed.  See:
   *
   * https://developer.mozilla.org/en-US/docs/Web/API/XMLHttpRequest/Sending_and_Receiving_Binary_Data
   *
   * By default values are json parsed from text
   */
  responseType?: 'json' | 'text' | 'arraybuffer' | 'blob';

  /**
   * The credentials read-only property of the Request interface indicates whether the user agent should send cookies from the other domain in the case of cross-origin requests.
   */
  credentials?: RequestCredentials;

  /**
   * @deprecated withCredentials is deprecated in favor of credentials
   */
  withCredentials?: boolean;
};

/**
 * Response for fetch function in {@link BackendSrv}
 *
 * @public
 */
export interface FetchResponse<T = any> {
  data: T;
  readonly status: number;
  readonly statusText: string;
  readonly ok: boolean;
  readonly headers: Headers;
  readonly redirected: boolean;
  readonly type: ResponseType;
  readonly url: string;
  readonly config: BackendSrvRequest;
}

/**
 * Error type for fetch function in {@link BackendSrv}
 *
 * @public
 */
export interface FetchErrorDataProps {
  message?: string;
  status?: string;
  error?: string | any;
}

/**
 * Error type for fetch function in {@link BackendSrv}
 *
 * @public
 */
export interface FetchError<T extends FetchErrorDataProps = any> {
  status: number;
  statusText?: string;
  data: T;
  cancelled?: boolean;
  isHandled?: boolean;
  config: BackendSrvRequest;
}

/**
 * Used to communicate via http(s) to a remote backend such as the Grafana backend,
 * a datasource etc. The BackendSrv is using the {@link https://developer.mozilla.org/en-US/docs/Web/API/Fetch_API | Fetch API}
 * under the hood to handle all the communication.
 *
 * The request function can be used to perform a remote call by specifying a {@link BackendSrvRequest}.
 * To make the BackendSrv a bit easier to use we have added a couple of shorthand functions that will
 * use default values executing the request.
 *
 * @remarks
 * By default, Grafana displays an error message alert if the remote call fails. To prevent this from
 * happening `showErrorAlert = true` on the options object.
 *
 * @public
 */
export interface BackendSrv {
  get(url: string, params?: any, requestId?: string): Promise<any>;
  delete(url: string): Promise<any>;
  post(url: string, data?: any): Promise<any>;
  patch(url: string, data?: any): Promise<any>;
  put(url: string, data?: any): Promise<any>;

  /**
   * @deprecated Use the fetch function instead. If you prefer to work with a promise
   * wrap the Observable returned by fetch with the lastValueFrom function.
   */
  request(options: BackendSrvRequest): Promise<any>;

  /**
   * Special function used to communicate with datasources that will emit core
   * events that the Grafana QueryInspector and QueryEditor is listening for to be able
   * to display datasource query information. Can be skipped by adding `option.silent`
   * when initializing the request.
   *
   * @deprecated Use the fetch function instead
   */
  datasourceRequest<T = any>(options: BackendSrvRequest): Promise<FetchResponse<T>>;

  /**
   * Observable http request interface
   */
  fetch<T>(options: BackendSrvRequest): Observable<FetchResponse<T>>;
}

let singletonInstance: BackendSrv;

/**
 * Used during startup by Grafana to set the BackendSrv so it is available
 * via the {@link getBackendSrv} to the rest of the application.
 *
 * @internal
 */
export const setBackendSrv = (instance: BackendSrv) => {
  singletonInstance = instance;
};

/**
 * Used to retrieve the {@link BackendSrv} that can be used to communicate
 * via http(s) to a remote backend such as the Grafana backend, a datasource etc.
 *
 * @public
 */
export const getBackendSrv = (): BackendSrv => singletonInstance;
