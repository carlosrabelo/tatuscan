"""API routes - REST endpoints for inventory management."""
import logging

from flask import jsonify, request

from tatuscan.services import InventoryService
from tatuscan.services.exceptions import ServiceException
from tatuscan.utils import serialize_inventory
from tatuscan.extensions import db

from . import bp

logger = logging.getLogger(__name__)


@bp.route("/machines", methods=["GET"])
def list_inventory():
    """List all inventory items."""
    try:
        inventories = InventoryService.list_all(order_by="hostname", direction="asc")
        data = [serialize_inventory(inv) for inv in inventories]
        return jsonify({"items": data, "count": len(data)})
    except ServiceException as e:
        return jsonify({"error": e.message}), e.status_code


@bp.route("/inventory", methods=["GET"])
def list_inventory_alias():
    """Alias for /machines endpoint (backwards compatibility)."""
    return list_inventory()

@bp.route("/machines", methods=["POST"])
def add_inventory():
    """Create or update inventory."""
    try:
        data = request.get_json(silent=True) or {}
        logger.debug(f"Received data: {data}")

        inventory, created = InventoryService.create_or_update(data)

        status_code = 201 if created else 200
        message = "Inventário adicionado com sucesso" if created else "Inventário atualizado com sucesso"

        logger.info(f"Inventory {'created' if created else 'updated'}: {inventory.hostname}")
        return jsonify({"message": message, "item": serialize_inventory(inventory)}), status_code

    except ServiceException as e:
        logger.error(f"Service error: {e.message}")
        return jsonify({"error": e.message}), e.status_code


@bp.route("/machines/<machine_id>", methods=["PATCH"])
def update_inventory(machine_id: str):
    """Partially update inventory fields."""
    try:
        data = request.get_json(silent=True) or {}
        inventory = InventoryService.partial_update(machine_id, data)

        logger.info(f"Inventory {machine_id} partially updated")
        return jsonify({"message": "Inventário atualizado", "item": serialize_inventory(inventory)})

    except ServiceException as e:
        logger.error(f"Service error: {e.message}")
        return jsonify({"error": e.message}), e.status_code


@bp.route("/machines/<machine_id>", methods=["DELETE"])
def delete_inventory(machine_id: str):
    """Delete inventory by machine ID."""
    try:
        InventoryService.delete(machine_id)
        logger.info(f"Inventory {machine_id} deleted")
        return jsonify({"message": "Inventário removido"}), 200

    except ServiceException as e:
        logger.error(f"Service error: {e.message}")
        return jsonify({"error": e.message}), e.status_code


@bp.route("/health", methods=["GET"])
def health_check():
    """Health check endpoint."""
    try:
        db.session.execute(db.text("SELECT 1"))
        return jsonify({"status": "healthy"}), 200
    except Exception as e:
        logger.error(f"Health check failed: {str(e)}")
        return jsonify({"status": "unhealthy"}), 500
