import { ref } from "vue";
import { defineStore } from "pinia";
import type { Venue } from "@/types";
import { venuesApi } from "@/api";
import { useToast } from "@/composables/useToast";
import { errMessage } from "@/utils/apiError";

export const useVenuesStore = defineStore("venues", () => {
  const venues = ref<Venue[]>([]);
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchVenues() {
    loading.value = true;
    error.value = null;
    try {
      venues.value = await venuesApi.list();
    } catch (err) {
      error.value = errMessage(err, "Failed to fetch venues");
    } finally {
      loading.value = false;
    }
  }

  // Destructive (losers deleted): call, then refetch. Callers refresh events themselves,
  // since repointed events now carry a different venue_id.
  async function merge(from: string[], into: string) {
    if (from.length === 0 || !into) return null;
    error.value = null;
    try {
      const res = await venuesApi.merge(from, into);
      await fetchVenues();
      useToast().success(
        `Merged ${res.merged} venue${res.merged === 1 ? "" : "s"} · ${res.repointed_events} event${res.repointed_events === 1 ? "" : "s"} repointed`,
      );
      return res;
    } catch (err) {
      const msg = errMessage(err, "Failed to merge venues");
      error.value = msg;
      useToast().error(msg);
      return null;
    }
  }

  return { venues, loading, error, fetchVenues, merge };
});
