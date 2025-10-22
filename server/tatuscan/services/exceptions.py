"""Custom service exceptions."""


class ServiceException(Exception):
    """Base exception for service layer errors."""

    def __init__(self, message: str, status_code: int = 500):
        super().__init__(message)
        self.message = message
        self.status_code = status_code


class ValidationError(ServiceException):
    """Raised when input validation fails."""

    def __init__(self, message: str, missing_fields: list[str] | None = None):
        super().__init__(message, status_code=400)
        self.missing_fields = missing_fields or []


class NotFoundError(ServiceException):
    """Raised when a resource is not found."""

    def __init__(self, resource: str, identifier: str):
        message = f"{resource} not found: {identifier}"
        super().__init__(message, status_code=404)
        self.resource = resource
        self.identifier = identifier


class DatabaseError(ServiceException):
    """Raised when a database operation fails."""

    def __init__(self, message: str, original_error: Exception | None = None):
        super().__init__(message, status_code=500)
        self.original_error = original_error
