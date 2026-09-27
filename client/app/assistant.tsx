"use client";

import { RuntimeProvider } from "./runtime";
import { Thread } from "@/components/assistant-ui/elements/thread.aui";

export const Assistant = () => {
  return (
    <RuntimeProvider>
      <div className="h-dvh">
        <Thread />
      </div>
    </RuntimeProvider>
  );
};
