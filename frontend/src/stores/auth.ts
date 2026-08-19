import { ref, computed } from "vue";
import { defineStore } from "pinia";
import type {
  User,
  LoginPayload,
  UpdateProfilePayload,
  ChangePasswordPayload,
  ResetPasswordPayload,
} from "@/types";
import { authApi } from "@/api";
import { errMessage } from "@/utils/apiError";

export const useAuthStore = defineStore("auth", () => {
  const user = ref<User | null>(null);
  const token = ref<string | null>(localStorage.getItem("auth_token"));
  const loading = ref(false);
  const error = ref<string | null>(null);

  const isAuthenticated = computed(() => !!token.value);

  async function initialize() {
    const savedToken = localStorage.getItem("auth_token");
    if (!savedToken) {
      token.value = null;
      user.value = null;
      return;
    }

    token.value = savedToken;
    try {
      const me = await authApi.getMe();
      user.value = me;
    } catch {
      // Token is invalid or expired - clear it
      token.value = null;
      user.value = null;
      localStorage.removeItem("auth_token");
    }
  }

  async function login(payload: LoginPayload) {
    loading.value = true;
    error.value = null;
    try {
      const response = await authApi.login(payload);
      user.value = response.user;
      token.value = response.token;
      localStorage.setItem("auth_token", response.token);
    } catch (err) {
      error.value = errMessage(err, "Login failed");
      throw err;
    } finally {
      loading.value = false;
    }
  }

  function logout() {
    localStorage.removeItem("auth_token");
    user.value = null;
    token.value = null;
  }

  async function updateProfile(payload: UpdateProfilePayload) {
    const updated = await authApi.updateProfile(payload);
    user.value = updated;
    return updated;
  }

  async function changePassword(payload: ChangePasswordPayload) {
    await authApi.changePassword(payload);
  }

  async function resetPassword(payload: ResetPasswordPayload) {
    await authApi.resetPassword(payload);
  }

  return {
    user,
    token,
    loading,
    error,
    isAuthenticated,
    initialize,
    login,
    logout,
    updateProfile,
    changePassword,
    resetPassword,
  };
});
