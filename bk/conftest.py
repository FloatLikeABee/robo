# Manual scripts at repo root are not automated tests (avoid accidental collection).
collect_ignore = [
    "test_system.py",
    "test_llm_factory.py",
    "testing_alone.py",
]
