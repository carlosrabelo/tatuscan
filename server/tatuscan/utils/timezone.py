"""Timezone utilities."""
from datetime import datetime

import pytz
from flask import current_app


def get_timezone() -> pytz.tzinfo.BaseTzInfo:
    """Get configured timezone for the application."""
    tz_name = current_app.config.get("TIMEZONE", "America/Cuiaba")
    return pytz.timezone(tz_name)


def to_timezone(dt: datetime | None, tz: pytz.tzinfo.BaseTzInfo | None = None) -> datetime | None:
    """
    Convert datetime to timezone.

    Args:
        dt: Datetime to convert
        tz: Target timezone (uses app config if None)

    Returns:
        Converted datetime or None
    """
    if not dt:
        return None

    if tz is None:
        tz = get_timezone()

    if getattr(dt, "tzinfo", None) is None:
        return tz.localize(dt)
    return dt.astimezone(tz)
