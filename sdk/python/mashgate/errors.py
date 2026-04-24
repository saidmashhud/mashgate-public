"""Custom exception for Mashgate API errors."""

from __future__ import annotations

from typing import Any


class MashgateError(Exception):
    """Raised when the Mashgate API returns a non-2xx response."""

    def __init__(
        self,
        message: str,
        *,
        status: int = 0,
        code: str = "unknown_error",
        retryable: bool = False,
        details: dict[str, Any] | None = None,
    ) -> None:
        super().__init__(message)
        self.status = status
        self.code = code
        self.retryable = retryable
        self.details = details or {}

    def __repr__(self) -> str:
        return f"MashgateError(status={self.status}, code={self.code!r}, message={str(self)!r})"
