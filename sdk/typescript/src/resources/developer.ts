import type { MashgateClient } from "../client.js";
import type { CreateApplicationRequest, Application } from "../types.js";

export class DeveloperResource {
  constructor(private readonly client: MashgateClient) {}

  async createApplication(data: CreateApplicationRequest): Promise<Application> {
    return this.client.request<Application>("POST", "/v1/developer/applications", { body: data });
  }

  async listApplications(): Promise<{ applications: Application[] }> {
    return this.client.request("GET", "/v1/developer/applications");
  }

  async getApplication(appId: string): Promise<Application> {
    return this.client.request<Application>("GET", `/v1/developer/applications/${appId}`);
  }

  async deleteApplication(appId: string): Promise<void> {
    return this.client.request<void>("DELETE", `/v1/developer/applications/${appId}`);
  }
}
