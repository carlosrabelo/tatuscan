"""Inventory service - Business logic for inventory management."""
from datetime import datetime
from typing import Any

import pytz
from flask import current_app
from sqlalchemy import func

from tatuscan.extensions import db
from tatuscan.models.inventory import Inventory
from .exceptions import NotFoundError, ValidationError, DatabaseError


class InventoryService:
    """Service for managing inventory operations."""

    @staticmethod
    def get_timezone() -> pytz.tzinfo.BaseTzInfo:
        """Get configured timezone for the application."""
        tz_name = current_app.config.get("TIMEZONE", "America/Cuiaba")
        return pytz.timezone(tz_name)

    @staticmethod
    def list_all(order_by: str = "hostname", direction: str = "asc") -> list[Inventory]:
        """
        List all inventory items with optional sorting.

        Args:
            order_by: Column to sort by (hostname, os, created_at, etc.)
            direction: Sort direction (asc or desc)

        Returns:
            List of Inventory objects
        """
        sort_map = {
            "hostname": Inventory.hostname,
            "os": Inventory.os,
            "os_version": Inventory.os_version,
            "created_at": Inventory.created_at,
            "updated_at": Inventory.updated_at,
            "computer_activation": Inventory.computer_activation,
        }
        column = sort_map.get(order_by, Inventory.hostname)
        order_clause = column.desc() if direction == "desc" else column.asc()
        return Inventory.query.order_by(order_clause).all()

    @staticmethod
    def get_by_id(machine_id: str) -> Inventory:
        """
        Get inventory by machine ID.

        Args:
            machine_id: Machine identifier

        Returns:
            Inventory object

        Raises:
            NotFoundError: If inventory not found
        """
        inventory = db.session.get(Inventory, machine_id)
        if not inventory:
            raise NotFoundError("Inventory", machine_id)
        return inventory

    @staticmethod
    def create_or_update(data: dict[str, Any]) -> tuple[Inventory, bool]:
        """
        Create new inventory or update existing one.

        Args:
            data: Inventory data dictionary

        Returns:
            Tuple of (Inventory object, is_created boolean)

        Raises:
            ValidationError: If required fields are missing
            DatabaseError: If database operation fails
        """
        required = ["machine_id", "hostname", "ip", "os", "cpu_percent", "memory_total_mb"]
        missing = [k for k in required if k not in data]
        if missing:
            raise ValidationError(f"Missing required fields: {missing}", missing_fields=missing)

        try:
            existing = db.session.get(Inventory, data["machine_id"])
            tz = InventoryService.get_timezone()

            if existing:
                # Update existing
                existing.hostname = data["hostname"][:100]
                existing.ip = data["ip"][:45]
                existing.os = data["os"][:100]
                existing.os_version = (data.get("os_version") or "")[:100]
                existing.cpu_percent = float(data["cpu_percent"])
                existing.memory_total_mb = int(data["memory_total_mb"])
                existing.memory_used_mb = int(data.get("memory_used_mb") or 0)
                existing.computer_model = (data.get("computer_model") or "")[:100] if data.get("computer_model") else None

                # Handle optional datetime fields
                if "computer_activation" in data:
                    existing.computer_activation = InventoryService._parse_datetime(data["computer_activation"])
                if "activation_days" in data:
                    existing.activation_days = int(data["activation_days"]) if data.get("activation_days") is not None else None

                existing.updated_at = datetime.now(tz)
                db.session.commit()
                return existing, False

            # Create new
            now_tz = datetime.now(tz)
            inv = Inventory(
                machine_id=data["machine_id"],
                hostname=data["hostname"][:100],
                ip=data["ip"][:45],
                os=data["os"][:100],
                os_version=(data.get("os_version") or "")[:100],
                cpu_percent=float(data["cpu_percent"]),
                memory_total_mb=int(data["memory_total_mb"]),
                memory_used_mb=int(data.get("memory_used_mb") or 0),
                computer_model=(data.get("computer_model") or "")[:100] if data.get("computer_model") else None,
                created_at=now_tz,
            )

            # Handle optional datetime fields
            if "computer_activation" in data:
                inv.computer_activation = InventoryService._parse_datetime(data["computer_activation"])
            if "activation_days" in data:
                inv.activation_days = int(data["activation_days"]) if data.get("activation_days") is not None else None

            db.session.add(inv)
            db.session.commit()
            return inv, True

        except ValidationError:
            raise
        except Exception as e:
            db.session.rollback()
            raise DatabaseError(f"Failed to create/update inventory: {str(e)}", original_error=e)

    @staticmethod
    def partial_update(machine_id: str, data: dict[str, Any]) -> Inventory:
        """
        Partially update inventory fields.

        Args:
            machine_id: Machine identifier
            data: Fields to update

        Returns:
            Updated Inventory object

        Raises:
            NotFoundError: If inventory not found
            ValidationError: If no valid fields provided
            DatabaseError: If database operation fails
        """
        inventory = InventoryService.get_by_id(machine_id)
        tz = InventoryService.get_timezone()
        updated_fields = []

        try:
            if "computer_activation" in data:
                inventory.computer_activation = InventoryService._parse_datetime(data["computer_activation"])
                updated_fields.append("computer_activation")

            if "activation_days" in data:
                inventory.activation_days = (
                    int(data["activation_days"]) if data.get("activation_days") is not None else None
                )
                updated_fields.append("activation_days")

            if not updated_fields:
                raise ValidationError("No valid fields provided for update")

            if any(field != "computer_activation" for field in updated_fields):
                inventory.updated_at = datetime.now(tz)

            db.session.commit()
            return inventory

        except (ValidationError, NotFoundError):
            raise
        except Exception as e:
            db.session.rollback()
            raise DatabaseError(f"Failed to update inventory: {str(e)}", original_error=e)

    @staticmethod
    def delete(machine_id: str) -> None:
        """
        Delete inventory by machine ID.

        Args:
            machine_id: Machine identifier

        Raises:
            NotFoundError: If inventory not found
            DatabaseError: If database operation fails
        """
        inventory = InventoryService.get_by_id(machine_id)
        try:
            db.session.delete(inventory)
            db.session.commit()
        except Exception as e:
            db.session.rollback()
            raise DatabaseError(f"Failed to delete inventory: {str(e)}", original_error=e)

    @staticmethod
    def get_os_distribution() -> list[dict[str, Any]]:
        """
        Get distribution of inventories by operating system.

        Returns:
            List of dicts with 'os' and 'count' keys
        """
        results = (
            db.session.query(
                Inventory.os.label("os"),
                func.count(Inventory.os).label("count")
            )
            .group_by(Inventory.os)
            .order_by(func.count(Inventory.os).desc())
            .all()
        )
        return [{"os": r.os or "-", "count": int(r.count or 0)} for r in results]

    @staticmethod
    def get_version_distribution(top_n: int = 8) -> list[dict[str, Any]]:
        """
        Get distribution of inventories by OS version.

        Args:
            top_n: Number of top versions to show, rest grouped as "Others"

        Returns:
            List of dicts with 'version' and 'count' keys
        """
        results = (
            db.session.query(
                Inventory.os_version.label("os_version"),
                func.count(Inventory.os_version).label("count")
            )
            .group_by(Inventory.os_version)
            .order_by(func.count(Inventory.os_version).desc())
            .all()
        )

        if len(results) > top_n:
            top = results[:top_n]
            others_count = sum(int(r.count or 0) for r in results[top_n:])
            data = [{"version": r.os_version or "-", "count": int(r.count or 0)} for r in top]
            if others_count:
                data.append({"version": "Outros", "count": others_count})
            return data

        return [{"version": r.os_version or "-", "count": int(r.count or 0)} for r in results]

    @staticmethod
    def get_age_distribution() -> list[dict[str, Any]]:
        """
        Get distribution of computer ages in month ranges.

        Returns:
            List of dicts with 'range' and 'count' keys
        """
        tz = InventoryService.get_timezone()
        computers = Inventory.query.filter(Inventory.computer_activation.isnot(None)).all()
        ages_in_months = []

        for comp in computers:
            activation_date = comp.computer_activation
            if not activation_date:
                continue
            try:
                if isinstance(activation_date, str):
                    for fmt in ("%Y-%m-%d %H:%M:%S", "%Y-%m-%d"):
                        try:
                            activation_date = datetime.strptime(activation_date, fmt)
                            break
                        except ValueError:
                            continue
                    else:
                        continue

                if getattr(activation_date, "tzinfo", None) is None:
                    activation_date = tz.localize(activation_date)
                else:
                    activation_date = activation_date.astimezone(tz)

                days_diff = (datetime.now(tz) - activation_date).days
                months = days_diff / 30.42
                ages_in_months.append(months)
            except Exception:
                continue

        ranges = [
            {"min": 0,   "max": 12,    "label": "0–12",    "count": 0},
            {"min": 12,  "max": 36,    "label": "12–36",   "count": 0},
            {"min": 36,  "max": 60,    "label": "36–60",   "count": 0},
            {"min": 60,  "max": 120,   "label": "60–120",  "count": 0},
            {"min": 120, "max": 9999,  "label": ">120",    "count": 0},
        ]

        for m in ages_in_months:
            for r in ranges:
                if r["min"] <= m < r["max"]:
                    r["count"] += 1
                    break

        return [{"range": r["label"], "count": r["count"]} for r in ranges]

    @staticmethod
    def _parse_datetime(value: Any) -> datetime | None:
        """
        Parse datetime from various formats.

        Args:
            value: Value to parse (str, int, float, datetime, or None)

        Returns:
            Parsed datetime with timezone or None

        Raises:
            ValidationError: If format is invalid
        """
        tz = InventoryService.get_timezone()

        if value in (None, ""):
            return None
        if isinstance(value, datetime):
            return value.astimezone(tz) if value.tzinfo else tz.localize(value)

        if isinstance(value, (int, float)):
            return datetime.fromtimestamp(value, tz)

        s = str(value).strip()
        if not s:
            return None

        if s.endswith("Z"):
            s = s[:-1] + "+00:00"

        try:
            dt = datetime.fromisoformat(s)
            return dt.astimezone(tz) if dt.tzinfo else tz.localize(dt)
        except ValueError:
            pass

        for fmt in ("%Y-%m-%d", "%d/%m/%Y"):
            try:
                dt = datetime.strptime(s, fmt)
                return tz.localize(dt)
            except ValueError:
                continue

        raise ValidationError(f"Invalid datetime format: {value!r}")
