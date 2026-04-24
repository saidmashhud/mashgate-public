export class MashgateError extends Error {
  readonly status: number;
  readonly code: string;
  readonly retryable: boolean;
  readonly details?: Record<string, unknown>;
  /** Unique request identifier — include this when contacting support. */
  readonly requestId?: string;
  /** Link to the documentation page for this error code. */
  readonly docUrl: string;
  /** The request parameter that caused a validation error, if applicable. */
  readonly param?: string;

  constructor(options: {
    message: string;
    status: number;
    code?: string;
    retryable?: boolean;
    details?: Record<string, unknown>;
    requestId?: string;
    param?: string;
  }) {
    super(options.message);
    this.name = "MashgateError";
    this.status = options.status;
    this.code = options.code || "unknown_error";
    this.retryable = options.retryable ?? (options.status === 429 || options.status >= 500);
    this.details = options.details;
    this.requestId = options.requestId;
    this.docUrl = `https://docs.mashgate.io/errors#${this.code}`;
    this.param = options.param;
  }
}
