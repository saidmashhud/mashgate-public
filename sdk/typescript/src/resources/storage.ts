import type { MashgateClient } from "../client.js";
import type { FileEntry, UploadUrlResponse } from "../types.js";

export interface GenerateUploadUrlRequest {
  tenantId: string;
  filename: string;
  mimeType: string;
}

export class StorageResource {
  constructor(private readonly client: MashgateClient) {}

  async generateUploadUrl(data: GenerateUploadUrlRequest): Promise<UploadUrlResponse> {
    return this.client.request<UploadUrlResponse>("POST", "/v1/storage/upload-url", {
      body: data,
    });
  }

  async listFiles(tenantId: string): Promise<FileEntry[]> {
    return this.client.request<FileEntry[]>("GET", "/v1/storage/files", {
      query: { tenantId },
    });
  }

  async deleteFile(fileId: string, tenantId: string): Promise<void> {
    return this.client.request<void>("DELETE", `/v1/storage/files/${fileId}`, {
      query: { tenantId },
    });
  }

  async getDownloadUrl(fileId: string, tenantId: string): Promise<string> {
    const result = await this.client.request<{ url: string }>(
      "GET",
      `/v1/storage/files/${fileId}/url`,
      { query: { tenantId } },
    );
    return result.url;
  }
}
