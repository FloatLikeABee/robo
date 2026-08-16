import logging
import os
from datetime import datetime
from typing import Dict, List, Optional

from tinydb import TinyDB, Query

from .config import settings
from .models import MCPHostProfile, MCPHostCreateRequest, MCPHostUpdateRequest, MCPHostConfig


class MCPHostManager:
    """Manage MCP host configurations stored in TinyDB.

    This module provides a simple, standard CRUD interface for defining MCP servers
    that can be hosted or connected to by this application. It does not manage
    process lifecycles yet; it focuses purely on configuration storage.
    """

    def __init__(self) -> None:
        self.logger = logging.getLogger(__name__)
        self.hosts: Dict[str, MCPHostProfile] = {}

        os.makedirs(settings.data_directory, exist_ok=True)
        self.db_path = os.path.join(settings.data_directory, "mcp_hosts.json")
        self.db = TinyDB(self.db_path)
        self.query = Query()

        self._load_hosts()

    def _load_hosts(self) -> None:
        """Load MCP host profiles from TinyDB."""
        try:
            docs = self.db.all()
            for doc in docs:
                try:
                    host_id = doc.get("id")
                    data = doc.get("profile", {})
                    if "id" in data:
                        del data["id"]
                    profile = MCPHostProfile(id=host_id, **data)
                    self.hosts[host_id] = profile
                except Exception as e:
                    self.logger.error(f"Failed to load MCP host {doc.get('id')}: {e}")
            self.logger.info(f"Loaded {len(self.hosts)} MCP host profiles")
        except Exception as e:
            self.logger.error(f"Error loading MCP hosts: {e}")

    def _save_hosts(self) -> None:
        """Persist all MCP host profiles to TinyDB."""
        try:
            self.db.truncate()
            for host_id, profile in self.hosts.items():
                self.db.insert(
                    {
                        "id": host_id,
                        "profile": profile.model_dump(exclude={"id"}),
                    }
                )
            self.logger.info(f"Saved {len(self.hosts)} MCP host profiles")
        except Exception as e:
            self.logger.error(f"Error saving MCP hosts: {e}")

    def _generate_id(self, name: str) -> str:
        """Generate a stable, URL-friendly id from the host name."""
        base_id = name.strip().lower().replace(" ", "_")
        base_id = "".join(c for c in base_id if c.isalnum() or c in ("_", "-"))
        if not base_id:
            base_id = "mcp_host"

        candidate = base_id
        counter = 1
        while candidate in self.hosts:
            candidate = f"{base_id}_{counter}"
            counter += 1
        return candidate

    def list_hosts(self) -> List[MCPHostProfile]:
        """List all MCP host profiles."""
        return list(self.hosts.values())

    def get_host(self, host_id: str) -> Optional[MCPHostProfile]:
        """Get a single MCP host profile by id."""
        return self.hosts.get(host_id)

    def create_host(self, req: MCPHostCreateRequest) -> str:
        """Create a new MCP host profile."""
        host_id = self._generate_id(req.name)
        now = datetime.utcnow().isoformat()
        config = MCPHostConfig(
            name=req.name,
            description=req.description,
            transport=req.transport,
            command=req.command,
            args=req.args or [],
            env=req.env or {},
            working_dir=req.working_dir,
            host=req.host,
            port=req.port,
            url=req.url,
            is_active=req.is_active,
            metadata=req.metadata or {},
        )
        profile = MCPHostProfile(
            id=host_id,
            name=req.name,
            description=req.description,
            config=config,
            created_at=now,
            updated_at=now,
        )
        self.hosts[host_id] = profile
        self._save_hosts()
        self.logger.info(f"Created MCP host profile: {host_id}")
        return host_id

    def update_host(self, host_id: str, req: MCPHostUpdateRequest) -> bool:
        """Update an existing MCP host profile."""
        if host_id not in self.hosts:
            return False
        try:
            existing = self.hosts[host_id]
            now = datetime.utcnow().isoformat()
            config = MCPHostConfig(
                name=req.name,
                description=req.description,
                transport=req.transport,
                command=req.command,
                args=req.args or [],
                env=req.env or {},
                working_dir=req.working_dir,
                host=req.host,
                port=req.port,
                url=req.url,
                is_active=req.is_active,
                metadata=req.metadata or {},
            )
            profile = MCPHostProfile(
                id=host_id,
                name=req.name,
                description=req.description,
                config=config,
                created_at=existing.created_at,
                updated_at=now,
            )
            self.hosts[host_id] = profile
            self._save_hosts()
            self.logger.info(f"Updated MCP host profile: {host_id}")
            return True
        except Exception as e:
            self.logger.error(f"Error updating MCP host {host_id}: {e}")
            return False

    def delete_host(self, host_id: str) -> bool:
        """Delete an MCP host profile."""
        if host_id not in self.hosts:
            return False
        try:
            del self.hosts[host_id]
            self.db.remove(self.query.id == host_id)
            self.logger.info(f"Deleted MCP host profile: {host_id}")
            return True
        except Exception as e:
            self.logger.error(f"Error deleting MCP host {host_id}: {e}")
            return False

    def list_host_tools(self, host_id: str):
        """Return tools exposed by an external MCP host (placeholder for host-specific discovery)."""
        host = self.get_host(host_id)
        if not host:
            return []
        # Placeholder: real implementation would query the host via its transport
        metadata_tools = host.config.metadata.get("tools", [])
        return [
            {
                "id": f"{host_id}/{tool.get('name', f'tool_{i}')}",
                "name": tool.get("name", f"Tool {i}"),
                "tool_type": "external",
                "description": tool.get("description", ""),
                "is_active": host.config.is_active,
                "config": tool.get("config", {}),
                "external_host_id": host_id,
            }
            for i, tool in enumerate(metadata_tools)
        ]

    def call_host_tool(self, host_id: str, tool_id: str, arguments: dict):
        """Call a tool on an external MCP host (placeholder for host-specific transport)."""
        host = self.get_host(host_id)
        if not host:
            raise ValueError(f"Host {host_id} not found")
        # Placeholder: real implementation would forward via stdio/tcp/websocket
        return f"External tool {tool_id} on host {host_id} called with {arguments}"

