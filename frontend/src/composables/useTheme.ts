import { ref, watchEffect } from "vue";

type Theme = "dark" | "light";
const KEY = "es-theme";
const theme = ref<Theme>((localStorage.getItem(KEY) as Theme) || "dark");

watchEffect(() => {
  const el = document.documentElement;
  el.classList.toggle("light", theme.value === "light");
  localStorage.setItem(KEY, theme.value);
});

export function useTheme() {
  const toggle = () => (theme.value = theme.value === "dark" ? "light" : "dark");
  const set = (t: Theme) => (theme.value = t);
  return { theme, toggle, set };
}
