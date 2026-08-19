import { api } from "./client";
import type {
  Category,
  CreateCategoryPayload,
  UpdateCategoryPayload,
  Artist,
  MessageResponse,
} from "@/types";

export const categoriesApi = {
  list(): Promise<Category[]> {
    return api.get<Category[]>("/categories");
  },

  create(payload: CreateCategoryPayload): Promise<Category> {
    return api.post<Category>("/categories", payload);
  },

  update(id: string, payload: UpdateCategoryPayload): Promise<Category> {
    return api.put<Category>(`/categories/${id}`, payload);
  },

  remove(id: string): Promise<MessageResponse> {
    return api.delete<MessageResponse>(`/categories/${id}`);
  },

  listArtists(categoryId: string): Promise<Artist[]> {
    return api.get<Artist[]>(`/categories/${categoryId}/artists`);
  },

  addArtist(categoryId: string, artistId: string): Promise<MessageResponse> {
    return api.post<MessageResponse>(`/categories/${categoryId}/artists`, {
      artist_id: artistId,
    });
  },

  removeArtist(categoryId: string, artistId: string): Promise<MessageResponse> {
    return api.delete<MessageResponse>(`/categories/${categoryId}/artists/${artistId}`);
  },

  // Bulk add/remove artists ↔ categories in one transaction.
  assign(artist_ids: string[], add?: string[], remove?: string[]): Promise<{ artists: number }> {
    return api.post<{ artists: number }>("/categories/assign", { artist_ids, add, remove });
  },
};
