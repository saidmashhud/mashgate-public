import type { MashgateClient } from "../client.js";
import type {
  RegisterRequest,
  RegisterResponse,
  LoginRequest,
  LoginResponse,
  RefreshResponse,
  UserProfile,
  UserCapabilities,
  ValidateTokenResponse,
  UpdateProfileRequest,
} from "../types.js";

export class AuthResource {
  constructor(private readonly client: MashgateClient) {}

  /** Register a new user in a tenant. Returns the created user (no tokens). */
  async register(data: RegisterRequest): Promise<RegisterResponse> {
    return this.client.request<RegisterResponse>("POST", "/v1/auth/register", { body: data });
  }

  /** Login and receive JWT access + refresh tokens. Sets accessToken on the client. */
  async login(data: LoginRequest): Promise<LoginResponse> {
    const result = await this.client.request<LoginResponse>("POST", "/v1/auth/login", { body: data });
    this.client.setAccessToken(result.accessToken);
    return result;
  }

  async refreshToken(refreshToken: string): Promise<RefreshResponse> {
    const result = await this.client.request<RefreshResponse>("POST", "/v1/auth/refresh", {
      body: { refreshToken },
    });
    this.client.setAccessToken(result.accessToken);
    return result;
  }

  async logout(): Promise<void> {
    await this.client.request<void>("POST", "/v1/auth/logout");
    this.client.setAccessToken(undefined);
  }

  /** Get profile for a specific user (admin use). Omit userId to get own profile. */
  async getProfile(userId?: string): Promise<UserProfile> {
    const path = userId ? `/v1/auth/profile/${userId}` : "/v1/auth/profile";
    return this.client.request<UserProfile>("GET", path);
  }

  /** Update a user's profile fields. */
  async updateProfile(userId: string, data: UpdateProfileRequest): Promise<UserProfile> {
    return this.client.request<UserProfile>("PUT", `/v1/auth/profile/${userId}`, { body: data });
  }

  async getCapabilities(): Promise<UserCapabilities> {
    return this.client.request<UserCapabilities>("GET", "/v1/auth/capabilities");
  }

  /**
   * Validate a JWT token and return the embedded claims.
   * Used by backend middleware to verify tokens issued by Mashgate.
   * In production with ext-authz, tokens are pre-validated at the network layer.
   */
  async validateToken(token: string): Promise<ValidateTokenResponse> {
    return this.client.request<ValidateTokenResponse>("POST", "/v1/auth/validate", {
      body: { token },
    });
  }
}
