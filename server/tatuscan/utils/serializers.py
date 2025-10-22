"""Serialization utilities."""
from datetime import datetime
from typing import Any

from tatuscan.models.inventory import Inventory
from .timezone import get_timezone


def serialize_inventory(inv: Inventory) -> dict[str, Any]:
    """
    Serialize Inventory model to dictionary.

    Args:
        inv: Inventory object

    Returns:
        Dictionary with serialized data
    """
    tz = get_timezone()

    def to_iso(dt: datetime | None) -> str | None:
        return dt.astimezone(tz).isoformat() if dt else None

    return {
        "machine_id": inv.machine_id,
        "hostname": inv.hostname,
        "ip": inv.ip,
        "os": inv.os,
        "os_version": inv.os_version,
        "cpu_percent": inv.cpu_percent,
        "memory_total_mb": inv.memory_total_mb,
        "memory_used_mb": inv.memory_used_mb,
        "computer_model": inv.computer_model,
        "computer_activation": to_iso(inv.computer_activation),
        "activation_days": inv.activation_days,
        "created_at": to_iso(inv.created_at),
        "updated_at": to_iso(inv.updated_at),
    }
