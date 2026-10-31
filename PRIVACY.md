# Privacy

The Go skill contains local instructions and reference text. It has no install-time
code, telemetry, account, or network service. Your coding agent may run your project's
tools, tests, and build scripts under that agent's permissions. Your agent provider's
policies govern the code and conversations you share with it.

Repository evaluation scripts are maintainer tools. `evaluate.py run` sends the public
task prompts and supplied fixture code to the configured agent provider through Codex
CLI and consumes that account's model usage. It records generated code, timing, token
usage when available, and tool output locally. Local validation and packaging do not
make model calls. No evaluation results are uploaded automatically.
