import { api } from "./client";
import type { SettingEntry } from "@/types";

export const settingsApi = {
  list(): Promise<SettingEntry[]> {
    return api.get<SettingEntry[]>("/settings");
  },

  update(settings: { key: string; value: string }[]): Promise<SettingEntry[]> {
    return api.put<SettingEntry[]>("/settings", { settings });
  },
};
