import type { MashgateClient } from "../client.js";
import type { MerchantSettings, UpdateSettingsRequest } from "../types.js";

export class SettingsResource {
  constructor(private readonly client: MashgateClient) {}

  async get(): Promise<MerchantSettings> {
    return this.client.request<MerchantSettings>("GET", "/v1/settings");
  }

  async update(data: UpdateSettingsRequest): Promise<MerchantSettings> {
    return this.client.request<MerchantSettings>("PUT", "/v1/settings", { body: data });
  }
}
