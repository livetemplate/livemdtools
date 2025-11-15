/**
 * MessageRouter - Multiplexes WebSocket messages by blockID
 */

import { MessageEnvelope } from "../types";

export type MessageHandler = (action: string, data: any) => void;

export class MessageRouter {
  private handlers: Map<string, MessageHandler> = new Map();
  private debug: boolean;

  constructor(debug = false) {
    this.debug = debug;
  }

  /**
   * Register a handler for a specific block ID
   */
  register(blockID: string, handler: MessageHandler): void {
    if (this.handlers.has(blockID)) {
      console.warn(`[MessageRouter] Overwriting handler for block: ${blockID}`);
    }
    this.handlers.set(blockID, handler);
    if (this.debug) {
      console.log(`[MessageRouter] Registered handler for block: ${blockID}`);
    }
  }

  /**
   * Unregister a handler for a specific block ID
   */
  unregister(blockID: string): void {
    this.handlers.delete(blockID);
    if (this.debug) {
      console.log(`[MessageRouter] Unregistered handler for block: ${blockID}`);
    }
  }

  /**
   * Route an incoming message to the appropriate handler
   */
  route(message: string | MessageEnvelope): void {
    try {
      const envelope: MessageEnvelope =
        typeof message === "string" ? JSON.parse(message) : message;

      const { blockID, action, data } = envelope;

      if (!blockID) {
        console.error("[MessageRouter] Message missing blockID:", envelope);
        return;
      }

      const handler = this.handlers.get(blockID);
      if (!handler) {
        console.warn(`[MessageRouter] No handler for block: ${blockID}`);
        return;
      }

      if (this.debug) {
        console.log(`[MessageRouter] Routing to ${blockID}:`, { action, data });
      }

      handler(action, data);
    } catch (error) {
      console.error("[MessageRouter] Error routing message:", error);
    }
  }

  /**
   * Send a message to the server (formatted as envelope)
   */
  createEnvelope(blockID: string, action: string, data: any = {}): MessageEnvelope {
    return { blockID, action, data };
  }

  /**
   * Get all registered block IDs
   */
  getRegisteredBlocks(): string[] {
    return Array.from(this.handlers.keys());
  }

  /**
   * Clear all handlers
   */
  clear(): void {
    this.handlers.clear();
    if (this.debug) {
      console.log("[MessageRouter] Cleared all handlers");
    }
  }
}
