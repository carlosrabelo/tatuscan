"""Base model class."""
from tatuscan.extensions import db


class BaseModel(db.Model):
    """Base model class with common functionality."""

    __abstract__ = True

    def to_dict(self):
        """Convert model to dictionary."""
        return {c.name: getattr(self, c.name) for c in self.__table__.columns}
