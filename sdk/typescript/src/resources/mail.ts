// Mail capability — Mashgate core primitive (ADR-0019).
//
// Mirrors `mail.v1.MailService` from
// mashgate/contracts/proto/v1/mail.proto, exposed over the gateway as REST
// via google.api.http transcoding.
//
// Auth: pass an end-user JWT for self-service mailbox operations (read/write
// scope), or an admin/service-account token for tenant-wide operations
// (mail:admin scope).
//
// Events (subscribe via webhooks): mail.received / mail.sent /
// mail.delivered / mail.bounced — see contracts/events/mail.*.json.

import type { MashgateClient } from "../client.js";

// ── Enum-like string types from mail.proto ──────────────────────────────────

export const MessageFolder = {
  Inbox: "MESSAGE_FOLDER_INBOX",
  Sent: "MESSAGE_FOLDER_SENT",
  Drafts: "MESSAGE_FOLDER_DRAFTS",
  Spam: "MESSAGE_FOLDER_SPAM",
  Trash: "MESSAGE_FOLDER_TRASH",
} as const;
export type MessageFolder = (typeof MessageFolder)[keyof typeof MessageFolder];

export const MailboxStatus = {
  Active: "MAILBOX_STATUS_ACTIVE",
  Frozen: "MAILBOX_STATUS_FROZEN",
  Closed: "MAILBOX_STATUS_CLOSED",
} as const;
export type MailboxStatus = (typeof MailboxStatus)[keyof typeof MailboxStatus];

export const DomainStatus = {
  Pending: "DOMAIN_STATUS_PENDING",
  Active: "DOMAIN_STATUS_ACTIVE",
  Suspended: "DOMAIN_STATUS_SUSPENDED",
} as const;
export type DomainStatus = (typeof DomainStatus)[keyof typeof DomainStatus];

export const SendStatus = {
  Queued: "SEND_STATUS_QUEUED",
  Sent: "SEND_STATUS_SENT",
  Delivered: "SEND_STATUS_DELIVERED",
  Failed: "SEND_STATUS_FAILED",
} as const;
export type SendStatus = (typeof SendStatus)[keyof typeof SendStatus];

// ── Domain types ────────────────────────────────────────────────────────────

export interface Mailbox {
  mailbox_id: string;
  tenant_id: string;
  subject_id: string;
  email: string;
  display_name?: string;
  status: MailboxStatus;
  quota_bytes: number;
  used_bytes: number;
  created_at: string;
  updated_at: string;
}

export interface MessagePreview {
  message_id: string;
  tenant_id: string;
  mailbox_id: string;
  from: string;
  to: string[];
  subject: string;
  preview?: string;
  received_at: string;
  read: boolean;
  folder: MessageFolder;
  labels?: string[];
  has_attachments?: boolean;
}

export interface MessageAttachment {
  attachment_id: string;
  filename: string;
  content_type: string;
  size_bytes: number;
  /** Signed URL (Mashgate storage-service). Не embedded в payload. */
  url: string;
}

export interface Message {
  preview: MessagePreview;
  body_text?: string;
  body_html?: string;
  headers?: Record<string, string>;
  attachments?: MessageAttachment[];
  cc?: string[];
  bcc?: string[];
  sent_at?: string;
  in_reply_to?: string;
}

export interface Domain {
  domain_id: string;
  tenant_id: string;
  name: string;
  status: DomainStatus;
  dkim_selector?: string;
  /** Public key для копи-паста в DNS TXT-record. */
  dkim_public_key?: string;
  mx_records: string[];
  spf_record?: string;
  dmarc_record?: string;
  /** Что не настроено в DNS — список записей которые VerifyDomain не нашёл. */
  verification_errors?: string[];
  created_at: string;
  last_verified_at?: string;
}

// ── Request shapes ──────────────────────────────────────────────────────────

export interface ListMessagesQuery {
  folder?: MessageFolder;
  limit?: number;
  cursor?: string;
  /** Admin only (mail:admin scope): чужой mailbox. Без этого — собственный. */
  mailbox_id_override?: string;
}

export interface ListMessagesResponse {
  items: MessagePreview[];
  next_cursor?: string;
  total?: number;
}

export interface SendMessageRequest {
  to: string[];
  cc?: string[];
  bcc?: string[];
  subject: string;
  body_text?: string;
  body_html?: string;
  headers?: Record<string, string>;
  /** ID файлов pre-uploaded через Mashgate storage-service. */
  attachment_ids?: string[];
  /** Message-ID письма-родителя для threading. */
  in_reply_to?: string;
  idempotency_key?: string;
}

export interface SendMessageResponse {
  message_id: string;
  status: SendStatus;
  note?: string;
  queued_at: string;
}

export interface UpdateMessageRequest {
  read?: boolean;
  folder?: MessageFolder;
  labels?: string[];
}

export interface ListMailboxesQuery {
  status?: MailboxStatus;
  limit?: number;
  cursor?: string;
}

export interface ListMailboxesResponse {
  items: Mailbox[];
  next_cursor?: string;
}

export interface CreateMailboxRequest {
  subject_id: string;
  email: string;
  display_name?: string;
  /** 0 = дефолтная квота тенанта. */
  quota_bytes?: number;
  idempotency_key?: string;
}

export interface ListDomainsQuery {
  status?: DomainStatus;
}

export interface ListDomainsResponse {
  items: Domain[];
}

export interface CreateDomainRequest {
  /** FQDN, e.g. "mail.entry-i.com". */
  name: string;
}

export interface RotateDKIMRequest {
  /** 1024 / 2048. По умолчанию 2048. */
  key_bits?: number;
}

// ── Resource ────────────────────────────────────────────────────────────────

export class MailResource {
  constructor(private readonly client: MashgateClient) {}

  // ── User-facing (mail:read / mail:write) ─────────────────────────────

  /** Mailbox текущего пользователя (по subject из JWT). */
  getMyMailbox(): Promise<Mailbox> {
    return this.client.request<Mailbox>("GET", "/v1/mail/mailboxes/me");
  }

  /** Список писем в папке (cursor pagination). */
  listMessages(query?: ListMessagesQuery): Promise<ListMessagesResponse> {
    return this.client.request<ListMessagesResponse>("GET", "/v1/mail/messages", {
      query: {
        folder: query?.folder,
        limit: query?.limit,
        cursor: query?.cursor,
        mailbox_id_override: query?.mailbox_id_override,
      },
    });
  }

  /** Полное письмо (body + headers + attachment links). */
  getMessage(messageId: string): Promise<Message> {
    return this.client.request<Message>("GET", `/v1/mail/messages/${messageId}`);
  }

  /**
   * Отправить письмо. Идемпотентно через `idempotency_key` — повторный
   * вызов с тем же ключом возвращает тот же `message_id`. Реальная SMTP-доставка
   * происходит асинхронно; статус потом приходит через события
   * `mail.sent` / `mail.delivered` / `mail.bounced`.
   */
  sendMessage(req: SendMessageRequest): Promise<SendMessageResponse> {
    return this.client.request<SendMessageResponse>("POST", "/v1/mail/messages", {
      body: req,
    });
  }

  /** Обновить флаги (read/unread, folder, labels). Только переданные поля меняются. */
  updateMessage(messageId: string, req: UpdateMessageRequest): Promise<Message> {
    return this.client.request<Message>("PATCH", `/v1/mail/messages/${messageId}`, {
      body: req,
    });
  }

  /**
   * Soft-delete — moves to TRASH. Если письмо уже в TRASH и `hardDelete=true` —
   * удаляется навсегда.
   */
  deleteMessage(messageId: string, hardDelete = false): Promise<void> {
    return this.client.request<void>("DELETE", `/v1/mail/messages/${messageId}`, {
      query: { hard_delete: hardDelete },
    });
  }

  // ── Admin-facing (mail:admin) ────────────────────────────────────────

  /** Все ящики тенанта (admin). */
  listMailboxes(query?: ListMailboxesQuery): Promise<ListMailboxesResponse> {
    return this.client.request<ListMailboxesResponse>("GET", "/v1/mail/mailboxes", {
      query: {
        status: query?.status,
        limit: query?.limit,
        cursor: query?.cursor,
      },
    });
  }

  /** Создать ящик (admin). Email должен быть в одном из активных доменов тенанта. */
  createMailbox(req: CreateMailboxRequest): Promise<Mailbox> {
    return this.client.request<Mailbox>("POST", "/v1/mail/mailboxes", { body: req });
  }

  /** Домены тенанта. */
  listDomains(query?: ListDomainsQuery): Promise<ListDomainsResponse> {
    return this.client.request<ListDomainsResponse>("GET", "/v1/mail/domains", {
      query: { status: query?.status },
    });
  }

  /**
   * Зарегистрировать домен (admin). Возвращает Domain в статусе `pending` с
   * сгенерированным `dkim_public_key` — пользователь должен прописать TXT-записи
   * в свой DNS (DKIM/SPF/DMARC), затем вызвать `verifyDomain`.
   */
  createDomain(req: CreateDomainRequest): Promise<Domain> {
    return this.client.request<Domain>("POST", "/v1/mail/domains", { body: req });
  }

  /**
   * Проверить DNS-записи домена. Если все на месте — статус становится
   * `active` и домен может отправлять/принимать почту. Иначе возвращает
   * `verification_errors` со списком проблем.
   */
  verifyDomain(domainId: string): Promise<Domain> {
    return this.client.request<Domain>(
      "POST",
      `/v1/mail/domains/${domainId}/verify`,
      { body: {} },
    );
  }

  /**
   * Сгенерировать новый DKIM-ключ. Старый селектор остаётся активным до
   * следующей ротации (через ~30 дней) — это grace period для успешной
   * пропагации DNS у получателей.
   */
  rotateDKIM(domainId: string, req?: RotateDKIMRequest): Promise<Domain> {
    return this.client.request<Domain>(
      "POST",
      `/v1/mail/domains/${domainId}/dkim/rotate`,
      { body: { key_bits: req?.key_bits ?? 2048 } },
    );
  }
}
