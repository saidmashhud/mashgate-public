import type { MashgateClient } from "../client.js";
import type { Channel, Message } from "../types.js";

export interface CreateChannelRequest {
  tenantId: string;
  channelId: string;
  name: string;
  channelType?: "public" | "private" | "direct";
  memberIds?: string[];
}

export interface SendMessageRequest {
  tenantId: string;
  senderId: string;
  content: string;
  contentType?: "text" | "image" | "file";
}

export interface ListMessagesOptions {
  tenantId: string;
  before?: string;
  limit?: number;
}

export class ChatResource {
  constructor(private readonly client: MashgateClient) {}

  async createChannel(data: CreateChannelRequest): Promise<Channel> {
    return this.client.request<Channel>("POST", "/v1/chat/channels", { body: data });
  }

  async listChannels(tenantId: string): Promise<Channel[]> {
    return this.client.request<Channel[]>("GET", "/v1/chat/channels", {
      query: { tenantId },
    });
  }

  async sendMessage(channelId: string, data: SendMessageRequest): Promise<Message> {
    return this.client.request<Message>("POST", `/v1/chat/channels/${channelId}/messages`, {
      body: data,
    });
  }

  async listMessages(channelId: string, options: ListMessagesOptions): Promise<Message[]> {
    return this.client.request<Message[]>("GET", `/v1/chat/channels/${channelId}/messages`, {
      query: {
        tenantId: options.tenantId,
        before: options.before,
        limit: options.limit,
      },
    });
  }

  async getMembers(channelId: string, tenantId: string): Promise<string[]> {
    return this.client.request<string[]>("GET", `/v1/chat/channels/${channelId}/members`, {
      query: { tenantId },
    });
  }

  async deleteMessage(channelId: string, messageId: string, tenantId: string): Promise<void> {
    return this.client.request<void>(
      "DELETE",
      `/v1/chat/channels/${channelId}/messages/${messageId}`,
      { query: { tenantId } },
    );
  }
}
