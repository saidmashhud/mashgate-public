import type { MashgateClient } from "../client.js";

/** Состояние задания. Строки, а не числа: REST отдаёт enum именем. */
export type AIJobState =
  | "JOB_STATE_UNSPECIFIED"
  | "JOB_STATE_QUEUED"
  | "JOB_STATE_RUNNING"
  | "JOB_STATE_DONE"
  | "JOB_STATE_FAILED";

export interface AICompleteRequest {
  /**
   * Системная часть промпта.
   *
   * Держите её неизменной между вызовами одного сценария: на платформе включено
   * переиспользование кеша префилла, и совпадение системной части экономит около
   * трети времени. Подстановка сюда даты или имени пользователя тихо ломает кеш —
   * запрос отработает, но заметно дольше.
   */
  system?: string;
  user: string;
  /** Схема ответа в JSON Schema. Единственный способ получить строгий JSON. */
  jsonSchema?: string;
  maxTokens?: number;
  /** Ноль означает значение по умолчанию, а не нулевую температуру. */
  temperature?: number;
}

export interface AICompleteResponse {
  text: string;
  promptTokens: number;
  completionTokens: number;
  /** Сколько токенов префилла переиспользовано из кеша. */
  cachedTokens: number;
}

export interface AICompletionJob {
  jobId: string;
  state: AIJobState;
  result?: AICompleteResponse;
  error?: string;
}

export interface AISubmitRequest {
  request: AICompleteRequest;
  /** Куда сообщить о готовности. Пусто — опрашивайте `getJob` сами. */
  callbackUrl?: string;
}

export interface AIEmbedResponse {
  embeddings: Array<{ values: number[] }>;
}

export interface AIStatus {
  available: boolean;
  model: string;
  contextSize: number;
  /** Модель обслуживает один запрос за раз — очередь и есть время ожидания. */
  queueDepth: number;
}

export interface AIWaitOptions {
  /** Сколько ждать, миллисекунды. По умолчанию пять минут. */
  timeoutMs?: number;
  /** Пауза между опросами, миллисекунды. По умолчанию две секунды. */
  pollIntervalMs?: number;
  signal?: AbortSignal;
}

/**
 * Языковая модель как возможность платформы.
 *
 * Модель одна на установку и живёт на самом сервере; ключ к ней не покидает
 * платформу, поэтому вертикали обращаются сюда, а не к провайдеру напрямую.
 *
 * Определяющее ограничение — **префилл дороже генерации**. Обработка промпта
 * идёт быстрее генерации, но контекст в полторы тысячи токенов всё равно
 * занимает около минуты. Отсюда два пути вместо одного:
 *
 * - `complete` — синхронный, пока вход укладывается в сотни токенов:
 *   классификация, извлечение полей, короткая сводка;
 * - `submit` + `wait` — всё остальное. Синхронный вызов на таком входе держал бы
 *   соединение минуту и упёрся бы в таймаут шлюза.
 *
 * Для поиска по смыслу передавайте отрывки, а не документы целиком: платить надо
 * за длину входа, и двадцать фрагментов обойдутся дороже самого ответа.
 */
export class AIResource {
  constructor(private readonly client: MashgateClient) {}

  /** Короткий синхронный вызов. */
  async complete(data: AICompleteRequest): Promise<AICompleteResponse> {
    return this.client.request<AICompleteResponse>("POST", "/v1/ai/complete", { body: data });
  }

  /** Ставит задание в очередь и возвращает его сразу, не дожидаясь ответа. */
  async submit(data: AISubmitRequest): Promise<AICompletionJob> {
    return this.client.request<AICompletionJob>("POST", "/v1/ai/jobs", { body: data });
  }

  async getJob(jobId: string): Promise<AICompletionJob> {
    return this.client.request<AICompletionJob>("GET", `/v1/ai/jobs/${encodeURIComponent(jobId)}`);
  }

  /**
   * Ждёт завершения задания, опрашивая его.
   *
   * Отказ модели — это разрешившееся задание со `state: JOB_STATE_FAILED`, а не
   * исключение, поэтому проверяйте состояние. Исключение здесь означает, что
   * ответа не дождались или вызов отменили.
   */
  async wait(jobId: string, options: AIWaitOptions = {}): Promise<AICompletionJob> {
    const timeoutMs = options.timeoutMs ?? 5 * 60_000;
    const pollIntervalMs = options.pollIntervalMs ?? 2_000;
    const deadline = Date.now() + timeoutMs;

    for (;;) {
      const job = await this.getJob(jobId);
      if (job.state === "JOB_STATE_DONE" || job.state === "JOB_STATE_FAILED") return job;

      if (Date.now() + pollIntervalMs > deadline) {
        throw new Error(`mashgate: задание ${jobId} не завершилось за ${timeoutMs} мс (состояние ${job.state})`);
      }
      await new Promise<void>((resolve, reject) => {
        const timer = setTimeout(resolve, pollIntervalMs);
        options.signal?.addEventListener(
          "abort",
          () => {
            clearTimeout(timer);
            reject(new Error("mashgate: ожидание задания отменено"));
          },
          { once: true },
        );
      });
    }
  }

  /** Векторы для поиска по смыслу. Отдельный вызов: другая модель и другая цена. */
  async embed(texts: string[]): Promise<AIEmbedResponse> {
    return this.client.request<AIEmbedResponse>("POST", "/v1/ai/embed", { body: { texts } });
  }

  /**
   * Живость и текущая модель.
   *
   * Нужно, чтобы отличить «модель не ответила» от «модель ответила плохо»: без
   * этого разбор жалоб идёт вслепую.
   */
  async status(): Promise<AIStatus> {
    return this.client.request<AIStatus>("GET", "/v1/ai/status");
  }
}
