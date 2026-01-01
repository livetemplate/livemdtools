# {{.Title}}

A Kubernetes pods dashboard built with [Tinkerdown](https://github.com/livetemplate/tinkerdown).

## Prerequisites

- `kubectl` configured with cluster access
- `jq` installed for JSON processing

## Running

```bash
cd {{.ProjectName}}
tinkerdown serve
```

Open http://localhost:8080 in your browser.

## Customizing

Edit `get-pods.sh` to change the kubectl command or jq filter.
