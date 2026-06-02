#!/bin/bash
source "$(dirname "$0")/lib/common.sh"
load_env

title "18 — dbt + Great Expectations + Feast"

if [ ! -d "$HOME/dbt-env" ]; then
  python3 -m venv ~/dbt-env
fi

~/dbt-env/bin/pip install dbt-core dbt-postgres great-expectations feast \
  --quiet 2>&1 | tail -3

grep -q "dbt-env" ~/.bashrc || \
  echo 'export PATH="$HOME/dbt-env/bin:$PATH"' >> ~/.bashrc

~/dbt-env/bin/dbt --version | head -2
~/dbt-env/bin/python -c "import feast; print('Feast:', feast.__version__)"
~/dbt-env/bin/python -c "import great_expectations; print('GE:', great_expectations.__version__)"

ok "dbt + Great Expectations + Feast installés"
