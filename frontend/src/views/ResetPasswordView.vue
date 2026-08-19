<script setup lang="ts">
import { ref } from "vue";
import { useRouter, RouterLink } from "vue-router";
import { useAuthStore } from "@/stores/auth";
import { errMessage } from "@/utils/apiError";
import { Btn, TextField } from "@/components/ui";
import AuthCard from "@/components/auth/AuthCard.vue";

const router = useRouter();
const auth = useAuthStore();

const username = ref("");
const email = ref("");
const newPassword = ref("");
const confirmPassword = ref("");
const loading = ref(false);
const error = ref<string | null>(null);
const success = ref(false);

async function onSubmit() {
  error.value = null;
  if (newPassword.value.length < 6) {
    error.value = "Password must be at least 6 characters";
    return;
  }
  if (newPassword.value !== confirmPassword.value) {
    error.value = "Passwords do not match";
    return;
  }
  loading.value = true;
  try {
    await auth.resetPassword({
      username: username.value,
      email: email.value,
      new_password: newPassword.value,
    });
    success.value = true;
  } catch (err) {
    error.value = errMessage(err, "Password reset failed. Please check your username and email.");
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <AuthCard title="Reset password">
    <div v-if="success" class="flex flex-col gap-4">
      <div class="rounded-sm border border-ok-border bg-ok-bg p-3 text-xs text-ok-text">
        Your password has been reset successfully.
      </div>
      <Btn tone="accent" class="w-full justify-center" @click="router.push('/login')">
        Go to Login
      </Btn>
    </div>

    <form v-else class="flex flex-col gap-4" @submit.prevent="onSubmit">
      <p class="text-xs text-muted">
        Enter your username and email to verify your identity, then set a new password.
      </p>
      <TextField v-model="username" label="Username" autocomplete="username" required />
      <TextField v-model="email" label="Email" type="email" autocomplete="email" required />
      <TextField
        v-model="newPassword"
        label="New Password"
        type="password"
        autocomplete="new-password"
        revealable
        required
      />
      <TextField
        v-model="confirmPassword"
        label="Confirm New Password"
        type="password"
        autocomplete="new-password"
        revealable
        required
      />
      <p v-if="error" class="text-xs text-danger">{{ error }}</p>
      <Btn type="submit" tone="accent" :loading="loading" class="w-full justify-center">
        {{ loading ? "Resetting…" : "Reset Password" }}
      </Btn>
      <div class="text-center text-xs text-muted">
        Remember your password?
        <RouterLink to="/login" class="text-accent-text">Login</RouterLink>
      </div>
    </form>
  </AuthCard>
</template>
