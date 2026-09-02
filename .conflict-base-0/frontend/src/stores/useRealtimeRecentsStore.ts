import { create } from "zustand";
import { type RecentEndpoint, upsertRecents, recentsForKind } from "#lib/realtimeRecents";

interface RecentsState {
  recents: RecentEndpoint[];
  addRecent: (url: string, kind: "ws" | "sse") => void;
  forKind: (kind: "ws" | "sse") => RecentEndpoint[];
}

export const useRealtimeRecentsStore = create<RecentsState>((set, get) => ({
  recents: [],
  addRecent: (url, kind) => set((s) => ({ recents: upsertRecents(s.recents, url, kind) })),
  forKind: (kind) => recentsForKind(get().recents, kind),
}));
