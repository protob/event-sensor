<script setup lang="ts">
import { computed } from "vue";
import { flagUrl } from "@/utils";

// Small rectangular country flag, consistent sizing everywhere.
const props = withDefaults(defineProps<{ code?: string; w?: number }>(), {
  code: "",
  w: 18,
});

const h = computed(() => Math.round(props.w * 0.72));
</script>

<template>
  <img
    :src="flagUrl(code)"
    :alt="code"
    loading="lazy"
    class="inline-block shrink-0 rounded-[1px] border border-line-2 object-cover"
    :style="{ width: w + 'px', height: h + 'px' }"
  />
</template>
