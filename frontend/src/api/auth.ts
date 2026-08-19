import { api } from "./client";
import type {
  LoginPayload,
  AuthResponse,
  User,
  ChangePasswordPayload,
  UpdateProfilePayload,
  ResetPasswordPayload,
  MessageResponse,
} from "@/types";

export const authApi = {
  login(payload: LoginPayload): Promise<AuthResponse> {
    return api.post<AuthResponse>("/auth/login", payload);
  },

  getMe(): Promise<User> {
    return api.get<User>("/auth/me");
  },

  changePassword(payload: ChangePasswordPayload): Promise<MessageResponse> {
    return api.post<MessageResponse>("/auth/change-password", payload);
  },

  updateProfile(payload: UpdateProfilePayload): Promise<User> {
    return api.put<User>("/auth/profile", payload);
  },

  resetPassword(payload: ResetPasswordPayload): Promise<MessageResponse> {
    return api.post<MessageResponse>("/auth/reset-password", payload);
  },
};
