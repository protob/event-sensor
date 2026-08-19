import { ref } from "vue";

export interface Toast {
  id: number;
  message: string;
  type: "success" | "error" | "info" | "warning";
}

const toasts = ref<Toast[]>([]);
let nextId = 0;

// Toast state is a module-level ref, shared across every useToast() caller.
export function useToast() {
  function show(message: string, type: Toast["type"] = "info", duration = 4000) {
    const id = nextId++;
    toasts.value.push({ id, message, type });
    if (duration > 0) {
      setTimeout(() => dismiss(id), duration);
    }
  }

  function success(message: string, duration?: number) {
    show(message, "success", duration);
  }

  function error(message: string, duration?: number) {
    show(message, "error", duration);
  }

  function warning(message: string, duration?: number) {
    show(message, "warning", duration);
  }

  function info(message: string, duration?: number) {
    show(message, "info", duration);
  }

  function dismiss(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id);
  }

  return {
    toasts,
    show,
    success,
    error,
    warning,
    info,
    dismiss,
  };
}
