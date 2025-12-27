# Livemdtools Client

Client runtime for interactive Livemdtools documentation.

## Features

- **Interactive Blocks**: Live reactive UI powered by LiveTemplate
- **WASM Execution**: Run Go code in the browser with TinyGo
- **Code Editor**: Monaco editor with syntax highlighting
- **Persistence**: Save code edits to localStorage
- **Message Multiplexing**: Single WebSocket for multiple blocks

## Installation

```bash
npm install @livetemplate/tinkerdown-client
```

## Usage

### Auto-initialization

The client auto-initializes when the script loads. Just include it in your HTML:

```html
<script src="/assets/tinkerdown-client.browser.js"></script>
```

Configure via meta tags:

```html
<meta name="tinkerdown-ws-url" content="ws://localhost:8080/ws">
<meta name="tinkerdown-debug" content="true">
```

### Manual initialization

```typescript
import { LivemdtoolsClient } from '@livetemplate/tinkerdown-client';

const client = new LivemdtoolsClient({
  wsUrl: 'ws://localhost:8080/ws',
  debug: true,
  persistence: true,
  onConnect: () => console.log('Connected'),
  onDisconnect: () => console.log('Disconnected'),
});

// Discover blocks on the page
client.discoverBlocks();

// Connect to server
client.connect();
```

## Block Types

### Server Blocks

Read-only code display (server-side code):

```html
<div data-tinkerdown-block
     data-block-id="counter-state"
     data-block-type="server"
     data-language="go"
     data-readonly="true">
  <pre><code>// Go code here</code></pre>
</div>
```

### Interactive Blocks

Live reactive UI:

```html
<div data-tinkerdown-block
     data-block-id="counter-demo"
     data-block-type="interactive"
     data-state-ref="counter-state">
  <!-- LiveTemplate markup -->
</div>
```

### WASM Blocks

Editable code with execution:

```html
<div data-tinkerdown-block
     data-block-id="playground"
     data-block-type="wasm"
     data-language="go"
     data-editable="true">
  <pre><code>package main

import "fmt"

func main() {
    fmt.Println("Hello, World!")
}
</code></pre>
</div>
```

## Development

```bash
# Install dependencies
npm install

# Build
npm run build

# Watch mode
npm run dev

# Run tests
npm test
```

## License

MIT
