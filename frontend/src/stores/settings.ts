import { ref } from "vue";
import { defineStore } from "pinia";
import { settingsApi } from "@/api";

// Canonical setting keys.
export const SETTING_KEYS = {
  regionCodes: "region.codes",
  tmApiKey: "tm.api_key",
} as const;

export const useSettingsStore = defineStore("settings", () => {
  const map = ref<Record<string, string>>({});
  const loaded = ref(false);
  const loading = ref(false);

  async function load() {
    loading.value = true;
    try {
      const list = await settingsApi.list();
      const next: Record<string, string> = {};
      for (const s of list) next[s.key] = s.value;
      map.value = next;
      loaded.value = true;
    } finally {
      loading.value = false;
    }
  }

  async function set(key: string, value: string) {
    const prev = map.value[key];
    map.value = { ...map.value, [key]: value };
    try {
      await settingsApi.update([{ key, value }]);
    } catch (err) {
      if (prev === undefined) delete map.value[key];
      else map.value = { ...map.value, [key]: prev };
      throw err;
    }
  }

  function getBool(key: string): boolean {
    return map.value[key] === "1";
  }
  async function setBool(key: string, v: boolean) {
    await set(key, v ? "1" : "0");
  }

  function getList(key: string): string[] {
    return (map.value[key] || "")
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
  }
  async function setList(key: string, arr: string[]) {
    await set(key, arr.join(","));
  }

  return { map, loaded, loading, load, set, getBool, setBool, getList, setList };
});
