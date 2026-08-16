"""Lightweight checks that do not require network or local services."""

import pytest


def test_python_syntax_project_compiles():
    import compileall
    import pathlib

    root = pathlib.Path(__file__).resolve().parents[1]
    ok = compileall.compile_dir(root / "src", quiet=1)
    assert ok, "src/ failed to compile"


def test_config_module_importable():
    pytest.importorskip("pydantic_settings")
    from src.config import settings

    assert settings.api_port == 8000
