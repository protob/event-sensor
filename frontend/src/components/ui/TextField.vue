<script setup lang="ts">
import { ref } from "vue";
import IconEye from "~icons/mdi/eye";
import IconEyeOff from "~icons/mdi/eye-off";

withDefaults(
  defineProps<{
    label?: string;
    type?: string;
    placeholder?: string;
    autocomplete?: string;
    required?: boolean;
    revealable?: boolean;
  }>(),
  { type: "text" },
);

const model = defineModel<string>({ default: "" });
const show = ref(false);

// The root is the <label>; without this a data-testid or an id set by the caller would
// land there instead of on the field it names.
defineOptions({ inheritAttrs: false });
</script>

<template>
  <label class="block">
    <span v-if="label" class="block mb-1 font-mono text-meta uppercase tracking-wide text-faint">
      {{ label }}
    </span>
    <div class="relative">
      <input
        v-model="model"
        v-bind="$attrs"
        :type="revealable && show ? 'text' : type"
        :placeholder="placeholder"
        :autocomplete="autocomplete"
        :required="required"
        class="block w-full bg-surface-2 border border-line-2 rounded-sm px-2.5 py-1.5 text-sm text-body placeholder:text-faint focus:outline-none focus:border-accent-bright"
        :class="revealable ? 'pr-9' : ''"
      />
      <button
        v-if="revealable"
        type="button"
        class="absolute inset-y-0 right-0 flex items-center pr-2.5 text-faint hover:text-muted"
        :aria-label="show ? 'Hide' : 'Show'"
        @click="show = !show"
      >
        <IconEye v-if="show" class="h-4 w-4" />
        <IconEyeOff v-else class="h-4 w-4" />
      </button>
    </div>
  </label>
</template>
