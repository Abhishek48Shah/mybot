"use client";

import {
  AssistantRuntimeProvider,
  useLocalRuntime,
  type ChatModelAdapter,
} from "@assistant-ui/react";
import { useEffect, type ReactNode } from "react";

const API_URL = "http://localhost:4000/api/chat/stream";
const THREAD_HISTORY_KEY = "bharatpur-chat-thread-v1";

const modelAdapter: ChatModelAdapter = {
  async *run({ messages, abortSignal }) {
    const lastMessage = messages[messages.length - 1];

    if (!lastMessage || lastMessage.role !== "user") {
      throw new Error("No user message found");
    }

    const chatMessages = messages
      .map((message) => ({
        role: message.role,
        content: message.content
          .filter((part) => part.type === "text")
          .map((part) => part.text)
          .join(""),
      }))
      .filter((message) => message.content.trim() != "");
    const response = await fetch(API_URL, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Accept: "text/event-stream",
      },
      body: JSON.stringify({
        messages: chatMessages,
      }),
      signal: abortSignal,
    });

    if (!response.ok) {
      throw new Error(`Chat request failed: ${response.status}`);
    }

    if (!response.body) {
      throw new Error("Response body is empty");
    }

    const reader = response.body.getReader();
    const decoder = new TextDecoder();

    let buffer = "";
    let fullText = "";

    try {
      while (true) {
        const { value, done } = await reader.read();

        if (done) {
          break;
        }

        buffer += decoder.decode(value, { stream: true });

        const events = buffer.split(/\r?\n\r?\n/);
        buffer = events.pop() ?? "";

        for (const event of events) {
          let eventType = "";
          let data = "";

          for (const line of event.split(/\r?\n/)) {
            if (line.startsWith("event:")) {
              eventType = line.slice(6).trim();
            }

            if (line.startsWith("data:")) {
              data += line.slice(5).trim();
            }
          }

          if (eventType === "token" && data) {
            const parsed = JSON.parse(data);

            fullText += parsed.content ?? "";

            yield {
              content: [
                {
                  type: "text",
                  text: fullText,
                },
              ],
            };
          }

          if (eventType === "done") {
            return;
          }
        }
      }
    } finally {
      reader.releaseLock();
    }
  },
};

export function RuntimeProvider({ children }: { children: ReactNode }) {
  useEffect(() => {
    window.localStorage.removeItem(THREAD_HISTORY_KEY);
  }, []);

  const runtime = useLocalRuntime(modelAdapter);

  return <AssistantRuntimeProvider runtime={runtime}>{children}</AssistantRuntimeProvider>;
}
