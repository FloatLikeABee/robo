# Setuptools pytest plugins (e.g. langsmith) can break under some pydantic/Python
# combinations; disabling autoload keeps runs deterministic for this repo.
.PHONY: test
test:
	PYTEST_DISABLE_PLUGIN_AUTOLOAD=1 python -m pytest -q -p pytest_asyncio
