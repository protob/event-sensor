import { ref } from "vue";
import { defineStore } from "pinia";
import type { Category, CreateCategoryPayload, UpdateCategoryPayload, Artist } from "@/types";
import { categoriesApi } from "@/api";
import { errMessage } from "@/utils/apiError";

export const useCategoriesStore = defineStore("categories", () => {
  const categories = ref<Category[]>([]);
  const categoryArtists = ref<Record<string, Artist[]>>({});
  const loading = ref(false);
  const error = ref<string | null>(null);

  async function fetchCategories() {
    loading.value = true;
    error.value = null;
    try {
      categories.value = await categoriesApi.list();
    } catch (err) {
      error.value = errMessage(err, "Failed to fetch categories");
    } finally {
      loading.value = false;
    }
  }

  async function createCategory(payload: CreateCategoryPayload) {
    error.value = null;
    try {
      const category = await categoriesApi.create(payload);
      categories.value.push(category);
      return category;
    } catch (err) {
      error.value = errMessage(err, "Failed to create category");
      return null;
    }
  }

  async function updateCategory(id: string, payload: UpdateCategoryPayload) {
    error.value = null;
    try {
      const updated = await categoriesApi.update(id, payload);
      const idx = categories.value.findIndex((c) => c.id === id);
      if (idx !== -1) {
        categories.value[idx] = updated;
      }
      return updated;
    } catch (err) {
      error.value = errMessage(err, "Failed to update category");
      return null;
    }
  }

  async function deleteCategory(id: string) {
    error.value = null;
    try {
      await categoriesApi.remove(id);
      categories.value = categories.value.filter((c) => c.id !== id);
      delete categoryArtists.value[id];
    } catch (err) {
      error.value = errMessage(err, "Failed to delete category");
    }
  }

  async function fetchCategoryArtists(categoryId: string) {
    error.value = null;
    try {
      const artists = await categoriesApi.listArtists(categoryId);
      categoryArtists.value[categoryId] = artists;
      return artists;
    } catch (err) {
      error.value = errMessage(err, "Failed to fetch category artists");
      return [];
    }
  }

  async function addArtistToCategory(categoryId: string, artistId: string) {
    error.value = null;
    try {
      await categoriesApi.addArtist(categoryId, artistId);
      await fetchCategoryArtists(categoryId);
      const idx = categories.value.findIndex((c) => c.id === categoryId);
      if (idx !== -1) {
        categories.value[idx].artist_count += 1;
      }
    } catch (err) {
      error.value = errMessage(err, "Failed to add artist to category");
    }
  }

  async function removeArtistFromCategory(categoryId: string, artistId: string) {
    error.value = null;
    try {
      await categoriesApi.removeArtist(categoryId, artistId);
      if (categoryArtists.value[categoryId]) {
        categoryArtists.value[categoryId] = categoryArtists.value[categoryId].filter(
          (a) => a.id !== artistId,
        );
      }
      const idx = categories.value.findIndex((c) => c.id === categoryId);
      if (idx !== -1 && categories.value[idx].artist_count > 0) {
        categories.value[idx].artist_count -= 1;
      }
    } catch (err) {
      error.value = errMessage(err, "Failed to remove artist from category");
    }
  }

  return {
    categories,
    categoryArtists,
    loading,
    error,
    fetchCategories,
    createCategory,
    updateCategory,
    deleteCategory,
    fetchCategoryArtists,
    addArtistToCategory,
    removeArtistFromCategory,
  };
});
