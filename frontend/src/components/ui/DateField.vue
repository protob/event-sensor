<script setup lang="ts">
// Text-based ISO-8601 date input (YYYY-MM-DD): deterministic, never the OS/browser locale a
// native date input renders in, and no calendar popup by design. v-model is the ISO string.
const model = defineModel<string>({ default: "" });

// Digits and dashes only; typed dashes are never blocked. Bare digits get dashes
// auto-inserted (20070629 -> 2007-06-29); once a dash is present the input is left alone.
function onInput(e: Event) {
  const el = e.target as HTMLInputElement;
  let v = el.value.replace(/[^\d-]/g, "").slice(0, 10);
  if (!v.includes("-")) {
    const d = v.replace(/\D/g, "").slice(0, 8); // YYYYMMDD
    if (d.length > 6) v = `${d.slice(0, 4)}-${d.slice(4, 6)}-${d.slice(6)}`;
    else if (d.length > 4) v = `${d.slice(0, 4)}-${d.slice(4)}`;
  }
  el.value = v;
  model.value = v;
}
</script>

<template>
  <input
    :value="model"
    type="text"
    inputmode="numeric"
    placeholder="YYYY-MM-DD"
    maxlength="10"
    spellcheck="false"
    autocomplete="off"
    @input="onInput"
  />
</template>
