import { beforeEach, describe, expect, it, vi } from "vitest";
import { MashgateClient } from "../client.js";
import { MashgateError } from "../errors.js";
import { KycCheckType, KycSubjectType } from "../resources/kyc.js";
import { AlertCategory, AlertSeverity, AlertStatus } from "../resources/compliance.js";
import { MerchantType } from "../resources/merchant.js";

interface MockCall {
  url: string;
  init: RequestInit & { headers: Record<string, string> };
}

function mockFetchReturning(body: unknown) {
  return vi.fn().mockResolvedValue({
    ok: true,
    status: 200,
    statusText: "OK",
    headers: new Headers(),
    json: () => Promise.resolve(body),
  });
}

function lastCall(mock: ReturnType<typeof mockFetchReturning>): MockCall {
  const [url, init] = mock.mock.calls[mock.mock.calls.length - 1];
  return { url: String(url), init: init as MockCall["init"] };
}

describe("Fintech resources", () => {
  let mockFetch: ReturnType<typeof mockFetchReturning>;
  let client: MashgateClient;

  beforeEach(() => {
    mockFetch = mockFetchReturning({});
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      tenantId: "tenant-1",
      fetch: mockFetch,
    });
  });

  it("requests KYC with tenant context and explicit idempotency", async () => {
    mockFetch = mockFetchReturning({ check: { checkId: "check-1" } });
    client = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      tenantId: "tenant-1",
      fetch: mockFetch,
    });

    const result = await client.kyc.request(
      {
        subjectId: "user-1",
        subjectType: KycSubjectType.Individual,
        checkType: KycCheckType.Full,
      },
      "kyc-1",
    );

    expect(result.check.checkId).toBe("check-1");
    const { url, init } = lastCall(mockFetch);
    expect(url).toBe("https://api.mashgate.uz/v1/kyc/checks");
    expect(init.headers["X-Tenant-ID"]).toBe("tenant-1");
    expect(init.headers["Idempotency-Key"]).toBe("kyc-1");
    expect(JSON.parse(String(init.body))).toMatchObject({
      tenantId: "tenant-1",
      subjectType: "KYC_SUBJECT_INDIVIDUAL",
      checkType: "KYC_CHECK_FULL",
    });
  });

  it("lists compliance alerts with canonical query names", async () => {
    await client.compliance.list({
      subjectId: "user-1",
      status: AlertStatus.Open,
      severity: AlertSeverity.High,
      limit: 1,
    });
    const { url } = lastCall(mockFetch);
    expect(url).toContain("tenant_id=tenant-1");
    expect(url).toContain("subject_id=user-1");
    expect(url).toContain("status=ALERT_STATUS_OPEN");
    expect(url).toContain("severity=ALERT_SEVERITY_HIGH");
  });

  it("raises and resolves a compliance alert", async () => {
    await client.compliance.raise(
      {
        subjectId: "user-1",
        subjectType: "user",
        category: AlertCategory.AML,
        severity: AlertSeverity.High,
        source: "screening",
        sourceRef: "screen-1",
        description: "review",
      },
      "alert-1",
    );
    let { init } = lastCall(mockFetch);
    expect(init.headers["Idempotency-Key"]).toBe("alert-1");
    expect(JSON.parse(String(init.body)).tenantId).toBe("tenant-1");

    await client.compliance.resolve("alert-1", "cleared");
    ({ init } = lastCall(mockFetch));
    expect(JSON.parse(String(init.body))).toMatchObject({
      tenant_id: "tenant-1",
      alert_id: "alert-1",
      resolve_note: "cleared",
    });
  });

  it("onboards and suspends a merchant", async () => {
    await client.merchant.onboard(
      {
        subjectId: "business-1",
        merchantType: MerchantType.Business,
        displayName: "Merchant",
        legalName: "Merchant LLC",
        registrationNumber: "123",
        countryCode: "TJ",
        config: {
          acceptedCurrencies: ["TJS"],
          maxTransactionAmount: "1000",
          dailyVolumeLimit: "10000",
          monthlyVolumeLimit: "100000",
          primaryCurrency: "TJS",
          cryptoEnabled: false,
          fiatEnabled: true,
          allowedPaymentMethods: ["card"],
        },
      },
      "merchant-1",
    );
    let { init } = lastCall(mockFetch);
    expect(init.headers["Idempotency-Key"]).toBe("merchant-1");
    expect(JSON.parse(String(init.body)).tenantId).toBe("tenant-1");

    await client.merchant.suspend("merchant-1", "review");
    ({ init } = lastCall(mockFetch));
    expect(JSON.parse(String(init.body))).toMatchObject({
      tenant_id: "tenant-1",
      merchant_id: "merchant-1",
      suspend_reason: "review",
    });
  });

  it("fails before network access when tenant context is missing", () => {
    const tenantless = new MashgateClient({
      baseUrl: "https://api.mashgate.uz",
      apiKey: "mg_test_key",
      fetch: mockFetch,
    });
    expect(() => tenantless.kyc.list()).toThrowError(MashgateError);
    try {
      tenantless.kyc.list();
    } catch (error) {
      expect(error).toMatchObject({ code: "tenant_id_required", status: 400 });
    }
    expect(mockFetch).not.toHaveBeenCalled();
  });
});
