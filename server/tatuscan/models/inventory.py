"""Inventory model."""
from sqlalchemy import event

from tatuscan.extensions import db
from .base import BaseModel
from .mixins import TimestampMixin, ensure_created_at


class Inventory(BaseModel, TimestampMixin):
    """Inventory model for tracking machine information."""

    __tablename__ = "inventory"

    machine_id = db.Column(db.String(64), primary_key=True)
    hostname = db.Column(db.String(100), nullable=False)
    ip = db.Column(db.String(45), nullable=False)
    os = db.Column(db.String(100), nullable=False)
    os_version = db.Column(db.String(100), nullable=True)

    cpu_percent = db.Column(db.Float, nullable=False)
    memory_total_mb = db.Column(db.BigInteger, nullable=False)
    memory_used_mb = db.Column(db.BigInteger, nullable=True)

    computer_model = db.Column(db.String(100), nullable=True)
    computer_activation = db.Column(db.DateTime(timezone=True), nullable=True)
    activation_days = db.Column(db.Integer, nullable=True)

    def __repr__(self):
        return f"<Inventory {self.hostname}>"


# Register event listener for automatic created_at
event.listens_for(Inventory, "before_insert")(ensure_created_at)
