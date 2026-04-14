import { useMutation, useQueryClient } from "@tanstack/react-query";
import { groupAPI, type Group } from "@/lib/api";
import { queryKeys } from "./keys";

export function useUpdateGroupExtended() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: async ({
      id,
      name,
      auto_fetch_full_content,
    }: {
      id: number;
      name?: string;
      auto_fetch_full_content?: boolean | null;
    }) => {
      await groupAPI.update(id, { name, auto_fetch_full_content });
      return { id, name, auto_fetch_full_content };
    },
    onSuccess: ({ id, name, auto_fetch_full_content }) => {
      qc.setQueryData(queryKeys.groups.list(), (old: Group[] | undefined) =>
        old?.map((g) =>
          g.id === id
            ? {
                ...g,
                ...(name !== undefined && { name }),
                ...(auto_fetch_full_content !== undefined && {
                  auto_fetch_full_content,
                }),
              }
            : g,
        ),
      );
    },
  });
}
