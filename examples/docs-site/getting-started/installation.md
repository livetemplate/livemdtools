---
title: "Installation"
---

# Installation

Get Livemdtools up and running on your machine.

## Prerequisites

- Go 1.21 or later
- A terminal/command line

## Install from Source

```bash
git clone https://github.com/livetemplate/livemdtools
cd livemdtools
go install ./cmd/livemdtools
```

## Verify Installation

Check that Livemdtools is installed correctly:

```bash
livemdtools --version
```

## Create Your First Page

Create a new directory and a simple markdown file:

```bash
mkdir my-tutorial
cd my-tutorial
echo "# Hello Livemdtools" > index.md
```

Start the development server:

```bash
livemdtools serve .
```

Open your browser to `http://localhost:8080` and you should see your page!

## Next Steps

Now that you have Livemdtools installed, learn how to [create pages](/guides/creating-pages).
