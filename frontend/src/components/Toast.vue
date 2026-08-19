<script setup lang="ts">
import { useToast } from "@/composables/useToast";
import IconCheck from "~icons/mdi/check";
import IconClose from "~icons/mdi/close";
import IconAlert from "~icons/mdi/alert";
import IconInfo from "~icons/mdi/information";

const { toasts, dismiss } = useToast();

const typeClasses: Record<string, string> = {
  success: "bg-ok-bg border-ok-border text-ok-text",
  error: "bg-danger-chip border-danger-chip-border text-danger",
  warning: "bg-surface-2 border-line-2 text-body",
  info: "bg-surface-2 border-line-2 text-body",
};

const typeIcons = {
  success: IconCheck,
  error: IconClose,
  warning: IconAlert,
  info: IconInfo,
} as const;
</script>

<template>
  <Teleport to="body">
    <div class="pointer-events-none fixed right-4 top-4 z-50 flex flex-col gap-2">
      <TransitionGroup name="toast">
        <div
          v-for="toast in toasts"
          :key="toast.id"
          class="pointer-events-auto flex min-w-72 items-center gap-3 rounded-lg border px-4 py-3 shadow-lg"
          :class="typeClasses[toast.type]"
          data-testid="toast"
        >
          <component :is="typeIcons[toast.type]" class="h-5 w-5" />
          <span class="flex-1 text-sm font-medium">{{ toast.message }}</span>
          <button class="ml-2 opacity-60 hover:opacity-100" @click="dismiss(toast.id)">
            <IconClose class="h-4 w-4" />
          </button>
        </div>
      </TransitionGroup>
    </div>
  </Teleport>
</template>

<style scoped>
.toast-enter-active,
.toast-leave-active {
  transition: all 0.3s ease;
}
.toast-enter-from {
  opacity: 0;
  transform: translateX(100%);
}
.toast-leave-to {
  opacity: 0;
  transform: translateX(100%);
}
</style>
