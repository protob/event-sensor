<script setup lang="ts">
import { computed, onMounted, ref } from "vue";
import { useAuthStore } from "@/stores/auth";
import { useEventsStore } from "@/stores/events";
import { useSettingsStore, SETTING_KEYS } from "@/stores/settings";
import { useToast } from "@/composables";
import { countryName, REGION_PRESETS } from "@/utils";
import { Surface, Btn, TextField, Mono, Flag } from "@/components/ui";

const auth = useAuthStore();
const events = useEventsStore();
const settings = useSettingsStore();
const toast = useToast();

const user = computed(() => auth.user);
const initials = computed(() => (user.value?.username || "?").slice(0, 2).toUpperCase());

// --- profile ---
const editingProfile = ref(false);
const profileUsername = ref("");
const profileEmail = ref("");
function startEditProfile() {
  profileUsername.value = user.value?.username ?? "";
  profileEmail.value = user.value?.email ?? "";
  editingProfile.value = true;
}
async function saveProfile() {
  try {
    await auth.updateProfile({
      username: profileUsername.value,
      email: profileEmail.value,
    });
    editingProfile.value = false;
    toast.success("Profile updated");
  } catch {
    toast.error("Failed to update profile");
  }
}

// --- password ---
const pwCurrent = ref("");
const pwNew = ref("");
const pwConfirm = ref("");
const pwSaving = ref(false);
async function changePassword() {
  if (pwNew.value !== pwConfirm.value) {
    toast.error("New passwords do not match");
    return;
  }
  pwSaving.value = true;
  try {
    await auth.changePassword({
      current_password: pwCurrent.value,
      new_password: pwNew.value,
    });
    pwCurrent.value = pwNew.value = pwConfirm.value = "";
    toast.success("Password changed");
  } catch {
    toast.error("Failed to change password");
  } finally {
    pwSaving.value = false;
  }
}

// --- API key ---
const apiKey = ref("");
const apiKeySaving = ref(false);
async function saveApiKey() {
  apiKeySaving.value = true;
  try {
    await settings.set(SETTING_KEYS.tmApiKey, apiKey.value.trim());
    toast.success("API key saved");
  } catch {
    toast.error("Failed to save API key");
  } finally {
    apiKeySaving.value = false;
  }
}

// --- region filter ---
const regionPreset = REGION_PRESETS[0];
const selectedCodes = ref<string[]>([]);
function isSelected(code: string) {
  return selectedCodes.value.includes(code);
}
async function toggleCode(code: string) {
  // An empty list saves as "" and the backend reads that as "unset" - i.e. the full default
  // region. Refusing the last removal keeps the UI honest about what it can express.
  if (isSelected(code) && selectedCodes.value.length === 1) {
    toast.error("Keep at least one country - an empty region falls back to the default");
    return;
  }
  if (isSelected(code)) selectedCodes.value = selectedCodes.value.filter((c) => c !== code);
  else selectedCodes.value = [...selectedCodes.value, code];
  try {
    await settings.setList(SETTING_KEYS.regionCodes, selectedCodes.value);
  } catch {
    toast.error("Failed to save region");
  }
}

// --- data management ---
const pruning = ref(false);
async function prunePastTm() {
  if (!confirm("Delete all past Ticketmaster events you haven't claimed? This can't be undone."))
    return;
  pruning.value = true;
  try {
    await events.prunePastTm();
    toast.success("Past unclaimed Ticketmaster events deleted");
  } catch {
    toast.error("Failed to prune past events");
  } finally {
    pruning.value = false;
  }
}

onMounted(async () => {
  await settings.load();
  if (apiKey.value === "") apiKey.value = settings.map[SETTING_KEYS.tmApiKey] ?? "";
  const saved = settings.getList(SETTING_KEYS.regionCodes);
  selectedCodes.value = saved.length ? saved : [...regionPreset.codes];
});
</script>

<template>
  <div class="h-full overflow-y-auto">
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-5 p-gutter max-w-[1200px]">
      <!-- left column -->
      <div class="flex flex-col gap-5">
        <Surface header="Profile">
          <div class="flex items-center gap-4">
            <div
              class="h-12 w-12 rounded-full bg-accent-chip border border-accent-chip-border text-accent-text font-mono font-semibold flex items-center justify-center"
            >
              {{ initials }}
            </div>
            <div v-if="!editingProfile" class="min-w-0 flex-1">
              <div class="text-heading font-semibold">{{ user?.username }}</div>
              <Mono size="11" class="text-muted">{{ user?.email }}</Mono>
            </div>
            <div v-else class="flex-1 flex flex-col gap-2">
              <TextField v-model="profileUsername" label="Username" />
              <TextField v-model="profileEmail" label="Email" type="email" />
            </div>
            <div class="shrink-0">
              <Btn v-if="!editingProfile" tone="neutral" size="sm" @click="startEditProfile">
                ✎ Edit
              </Btn>
              <div v-else class="flex gap-1.5">
                <Btn tone="accent" size="sm" @click="saveProfile">Save</Btn>
                <Btn tone="neutral" size="sm" @click="editingProfile = false">Cancel</Btn>
              </div>
            </div>
          </div>
        </Surface>

        <Surface header="Change Password">
          <div class="flex flex-col gap-3">
            <TextField v-model="pwCurrent" label="Current Password" type="password" revealable />
            <TextField v-model="pwNew" label="New Password" type="password" revealable />
            <TextField
              v-model="pwConfirm"
              label="Confirm New Password"
              type="password"
              revealable
            />
            <div>
              <Btn tone="accent" size="md" :loading="pwSaving" @click="changePassword">
                Change Password
              </Btn>
            </div>
          </div>
        </Surface>

        <Surface header="API Configuration">
          <div class="flex flex-col gap-3">
            <TextField
              v-model="apiKey"
              label="Ticketmaster API Key"
              type="password"
              revealable
              placeholder="overrides server default"
            />
            <Mono size="10" class="text-faint">
              Stored per-user; overrides the server env/agenix key at fetch time. Leave blank to use
              the server default.
            </Mono>
            <div>
              <Btn
                tone="accent"
                size="md"
                :loading="apiKeySaving"
                data-testid="settings-save"
                @click="saveApiKey"
              >
                Save Key
              </Btn>
            </div>
          </div>
        </Surface>
      </div>

      <!-- right column -->
      <div class="flex flex-col gap-5">
        <Surface header="Data Management">
          <div class="flex flex-col gap-3">
            <div>
              <div class="text-sm text-body mb-0.5">Clear past Ticketmaster events</div>
              <Mono size="10" class="text-faint block mb-2">
                Deletes past TM events you never claimed. Claimed events (interested / going /
                attended / missed) are kept permanently.
              </Mono>
              <Btn tone="danger" size="sm" :loading="pruning" @click="prunePastTm">
                Clear past TM events
              </Btn>
            </div>
          </div>
        </Surface>

        <Surface header="Region Filter">
          <Mono size="10" class="text-faint block mb-2.5">
            Countries events are fetched and stored from.
          </Mono>
          <div class="flex flex-wrap gap-1.5">
            <button
              v-for="code in regionPreset.codes"
              :key="code"
              class="inline-flex items-center gap-1.5 px-2 py-1 rounded-sm border font-mono text-label transition-colors"
              :class="
                isSelected(code)
                  ? 'bg-accent-strong border-accent-bright text-accent-text'
                  : 'bg-surface-2 border-line-2 text-muted hover:text-body'
              "
              :data-testid="'settings-region-' + code"
              @click="toggleCode(code)"
            >
              <Flag :code="code" :w="16" />
              {{ countryName(code, code.toUpperCase()) }}
              <span v-if="isSelected(code)">✓</span>
            </button>
          </div>
        </Surface>
      </div>
    </div>
  </div>
</template>
