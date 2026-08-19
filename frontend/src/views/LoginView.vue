<script setup lang="ts">
import { ref } from "vue";
import { useRouter, useRoute, RouterLink } from "vue-router";
import { storeToRefs } from "pinia";
import { useAuthStore } from "@/stores/auth";
import { Btn, TextField } from "@/components/ui";
import AuthCard from "@/components/auth/AuthCard.vue";

const route = useRoute();
const router = useRouter();
const auth = useAuthStore();
const { loading, error } = storeToRefs(auth);

const username = ref("");
const password = ref("");

async function onSubmit() {
  // login throws after setting error.value - the throw is for callers that want it
  await auth.login({ username: username.value, password: password.value }).catch(() => {});
  if (!error.value) {
    const redirect = (route.query.redirect as string) || "/";
    router.push(redirect);
  }
}
</script>

<template>
  <AuthCard title="Login">
    <form class="flex flex-col gap-4" @submit.prevent="onSubmit">
      <TextField v-model="username" label="Username" autocomplete="username" required />
      <TextField
        v-model="password"
        label="Password"
        type="password"
        autocomplete="current-password"
        revealable
        required
      />
      <p v-if="error" class="text-xs text-danger">{{ error }}</p>
      <Btn type="submit" tone="accent" :loading="loading" class="w-full justify-center">
        {{ loading ? "Logging in…" : "Login" }}
      </Btn>
      <div class="text-center text-xs text-muted">
        <RouterLink to="/reset-password" class="text-accent-text">Forgot password?</RouterLink>
      </div>
    </form>
  </AuthCard>
</template>
