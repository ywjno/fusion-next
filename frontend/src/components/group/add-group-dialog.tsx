import { useState } from "react";
import { FolderPlus } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { useIsMobile } from "@/hooks/use-mobile";
import { useI18n } from "@/lib/i18n";
import { useUIStore } from "@/store";
import { useCreateGroupExtended } from "@/queries/groups.ext";
import { AutoFetchField } from "@/components/feed/auto-fetch-field";
import { toast } from "sonner";

export function AddGroupDialog() {
  const { t } = useI18n();
  const isMobile = useIsMobile();
  const { isAddGroupOpen, setAddGroupOpen } = useUIStore();
  const createGroup = useCreateGroupExtended();

  const [name, setName] = useState("");
  const [autoFetchValue, setAutoFetchValue] = useState<boolean | null | undefined>(undefined);

  const handleCreate = async () => {
    const trimmed = name.trim();
    if (!trimmed) return;

    try {
      await createGroup.mutateAsync({
        name: trimmed,
        auto_fetch_full_content: autoFetchValue,
      });
      setName("");
      setAutoFetchValue(undefined);
      setAddGroupOpen(false);
      toast.success(t("group.toast.created"));
    } catch {
      toast.error(t("group.toast.createFailed"));
    }
  };

  return (
    <Dialog
      open={isAddGroupOpen}
      onOpenChange={(open) => {
        setAddGroupOpen(open);
        if (!open) {
          setName("");
          setAutoFetchValue(undefined);
        }
      }}
    >
      <DialogContent className="sm:max-w-[400px]">
        <DialogHeader>
          <DialogTitle>{t("group.add.title")}</DialogTitle>
          <DialogDescription>
            {t("group.add.description")}
          </DialogDescription>
        </DialogHeader>
        <div className="space-y-4">
          <div className="relative">
            <label htmlFor="add-group-name" className="sr-only">
              {t("group.add.title")}
            </label>
            <FolderPlus className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              id="add-group-name"
              name="group-name"
              placeholder={t("group.add.placeholder")}
              value={name}
              onChange={(e) => setName(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleCreate()}
              className="pl-10"
              autoComplete="off"
              aria-label={t("group.add.placeholder")}
              autoFocus={!isMobile}
            />
          </div>
          <AutoFetchField
            value={autoFetchValue}
            onChange={setAutoFetchValue}
          />
        </div>
        <DialogFooter>
          <Button
            variant="outline"
            onClick={() => setAddGroupOpen(false)}
            disabled={createGroup.isPending}
          >
            {t("common.cancel")}
          </Button>
          <Button
            onClick={handleCreate}
            disabled={!name.trim() || createGroup.isPending}
          >
            {createGroup.isPending ? t("common.creating") : t("common.create")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
