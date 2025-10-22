"""Model mixins for common fields and behaviors."""
from datetime import datetime
import os

import pytz
from flask import current_app
from sqlalchemy import event

from tatuscan.extensions import db


class TimestampMixin:
    """Mixin for automatic timestamp management."""

    created_at = db.Column(db.DateTime(timezone=True), nullable=False)
    updated_at = db.Column(db.DateTime(timezone=True), nullable=True)


def ensure_created_at(mapper, connection, target):
    """Ensure created_at is set before insert."""
    if target.created_at is None:
        try:
            tz_str = current_app.config.get("TIMEZONE")
        except Exception:
            tz_str = None
        tz_str = tz_str or os.getenv("TIMEZONE", "America/Cuiaba")
        tz = pytz.timezone(tz_str)
        target.created_at = datetime.now(tz)
