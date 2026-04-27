"""Tests for the WalletAdminResource — admin/merchant-side WalletService.

Mocks the gateway via ``respx`` (httpx route mocker), so we exercise the
real ``MashgateClient.request`` path including header / query / body
serialisation, but without a network round-trip.
"""

from __future__ import annotations

import json

import httpx
import pytest
import respx

from mashgate import (
    Currency,
    MashgateClient,
    Mint,
    Network,
    TransactionReason,
    WalletStatus,
    WalletType,
)


BASE = "https://api.mashgate.uz"


@pytest.fixture
def client() -> MashgateClient:
    return MashgateClient(base_url=BASE, api_key="mg_test_key")


@respx.mock
def test_create_chain_posts_chain_endpoint_and_returns_mnemonic(client):
    route = respx.post(f"{BASE}/v1/wallets/chain").mock(
        return_value=httpx.Response(
            200,
            json={
                "wallet": {"wallet_id": "w-1", "currency": "USDC"},
                "mnemonic": "abandon ability ...",
            },
        )
    )

    out = client.wallet_admin.create_chain(
        subject_id="u-1",
        subject_type="user",
        currency=Currency.USDC,
        network=Network.SOLANA,
        idempotency_key="idem-1",
    )
    assert out["mnemonic"] == "abandon ability ..."
    assert out["wallet"]["wallet_id"] == "w-1"

    sent = json.loads(route.calls.last.request.content)
    # Enum members serialise to their plain string values.
    assert sent["currency"] == "USDC"
    assert sent["network"] == "SOLANA"
    assert sent["idempotency_key"] == "idem-1"


@respx.mock
def test_deposit_address_passes_mint(client):
    route = respx.get(f"{BASE}/v1/wallets/w-1/deposit-address").mock(
        return_value=httpx.Response(200, json={"address": "AtaAddrBase58Here"})
    )

    out = client.wallet_admin.deposit_address(
        "w-1", network=Network.SOLANA, mint=Mint.USDC_SOLANA_MAINNET
    )
    assert out["address"] == "AtaAddrBase58Here"

    qs = dict(route.calls.last.request.url.params)
    assert qs["network"] == "SOLANA"
    assert qs["mint"] == "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"


@respx.mock
def test_deposit_address_omits_mint_when_none(client):
    """Native asset path — empty / None mint must not appear in query string."""
    route = respx.get(f"{BASE}/v1/wallets/w-1/deposit-address").mock(
        return_value=httpx.Response(200, json={"address": "OwnerPubkey"})
    )
    client.wallet_admin.deposit_address("w-1", network=Network.SOLANA, mint=None)

    qs = dict(route.calls.last.request.url.params)
    assert qs.get("network") == "SOLANA"
    assert "mint" not in qs


@respx.mock
def test_withdraw_includes_mint_in_body_not_description(client):
    """L2 — explicit ``mint`` field; description must stay free of the legacy hack."""
    route = respx.post(f"{BASE}/v1/wallets/w-1/withdraw").mock(
        return_value=httpx.Response(
            200,
            json={"transaction_id": "tx-1", "status": "TRANSACTION_STATUS_PENDING"},
        )
    )

    client.wallet_admin.withdraw(
        "w-1",
        amount="10.50",
        destination_type="crypto_address",
        destination_id="DestSolanaAddr",
        network=Network.SOLANA,
        mint=Mint.USDC_SOLANA_MAINNET,
        idempotency_key="idem-w-1",
    )

    sent = json.loads(route.calls.last.request.content)
    assert sent["mint"] == "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
    assert sent["network"] == "SOLANA"
    assert sent.get("description") in (None, "")


@respx.mock
def test_freeze_and_unfreeze_endpoints(client):
    respx.post(f"{BASE}/v1/wallets/w-1/freeze").mock(
        return_value=httpx.Response(
            200, json={"wallet_id": "w-1", "status": "WALLET_STATUS_FROZEN"}
        )
    )
    respx.post(f"{BASE}/v1/wallets/w-1/unfreeze").mock(
        return_value=httpx.Response(
            200, json={"wallet_id": "w-1", "status": "WALLET_STATUS_ACTIVE"}
        )
    )

    f = client.wallet_admin.freeze("w-1", reason="fraud-investigation")
    assert f["status"] == WalletStatus.FROZEN.value

    u = client.wallet_admin.unfreeze("w-1", note="case-resolved")
    assert u["status"] == WalletStatus.ACTIVE.value


@respx.mock
def test_get_transaction(client):
    respx.get(f"{BASE}/v1/wallets/w-1/transactions/tx-99").mock(
        return_value=httpx.Response(200, json={"transaction_id": "tx-99"})
    )
    tx = client.wallet_admin.get_transaction("w-1", "tx-99")
    assert tx["transaction_id"] == "tx-99"


@respx.mock
def test_list_passes_cursor_and_limit(client):
    route = respx.get(f"{BASE}/v1/wallets").mock(
        return_value=httpx.Response(
            200, json={"wallets": [{"wallet_id": "w-1"}], "next_cursor": "opaque-token"}
        )
    )

    resp = client.wallet_admin.list(
        subject_id="user-1", limit=25, cursor="prev-token", status=WalletStatus.ACTIVE
    )
    assert resp["next_cursor"] == "opaque-token"

    qs = dict(route.calls.last.request.url.params)
    assert qs["subject_id"] == "user-1"
    assert qs["limit"] == "25"
    assert qs["cursor"] == "prev-token"
    assert qs["status"] == "WALLET_STATUS_ACTIVE"


@respx.mock
def test_credit_serialises_reason_enum(client):
    route = respx.post(f"{BASE}/v1/wallets/w-1/credit").mock(
        return_value=httpx.Response(200, json={"transaction_id": "tx-1"})
    )
    client.wallet_admin.credit(
        "w-1", amount="100.00", reason=TransactionReason.DEPOSIT, idempotency_key="i-1"
    )
    sent = json.loads(route.calls.last.request.content)
    assert sent["reason"] == "TRANSACTION_REASON_DEPOSIT"
    assert sent["amount"] == "100.00"


def test_typed_constants_have_expected_wire_values():
    """Sanity — the server-side parsers expect plain enum strings."""
    assert Currency.USDC.value == "USDC"
    assert Currency.UZS.value == "UZS"
    assert Network.SOLANA.value == "SOLANA"
    assert Network.ETHEREUM.value == "ETHEREUM"
    assert Mint.USDC_SOLANA_MAINNET.value == "EPjFWdd5AufqSSqeM2qN1xzybapC8G4wEGGkZwyTDt1v"
    assert WalletStatus.FROZEN.value == "WALLET_STATUS_FROZEN"
    assert WalletType.CRYPTO.value == "WALLET_TYPE_CRYPTO"


def test_string_aliases_accepted_alongside_enum_members(client):
    """Plain strings should still work — the resource ``str()``s them."""
    # No HTTP round-trip — just exercise _opt_str via list query params.
    # Deferred to a respx route to keep the call valid.
    with respx.mock:
        route = respx.get(f"{BASE}/v1/wallets").mock(
            return_value=httpx.Response(200, json={"wallets": []})
        )
        client.wallet_admin.list(status="WALLET_STATUS_ACTIVE")
        qs = dict(route.calls.last.request.url.params)
        assert qs["status"] == "WALLET_STATUS_ACTIVE"
