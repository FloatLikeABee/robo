"""BK must not ship the incomplete length-prefixed TCP MCP server."""

import importlib
import pathlib

import pytest


def test_tcp_mcp_modules_removed():
    """Importing the deleted TCP server or host manager is an error.

    Fails if either module is importable again.
    """
    with pytest.raises(ModuleNotFoundError):
        importlib.import_module("src.mcp_service")
    with pytest.raises(ModuleNotFoundError):
        importlib.import_module("src.mcp_host_manager")


def test_mcp_host_models_removed():
    """Host-profile models exist only to serve the removed /mcp/hosts API."""
    import src.models as models

    for name in (
        "MCPTransportType",
        "MCPHostConfig",
        "MCPHostProfile",
        "MCPHostCreateRequest",
        "MCPHostUpdateRequest",
    ):
        assert not hasattr(models, name), name


def test_api_source_has_no_mcp_routes():
    """The API module must not register or construct the TCP MCP server."""
    api_path = pathlib.Path(__file__).resolve().parents[1] / "src" / "api.py"
    text = api_path.read_text(encoding="utf-8")
    assert "/mcp/start" not in text
    assert "/mcp/hosts" not in text
    assert "MCPService" not in text
    assert "mcp_host_manager" not in text
