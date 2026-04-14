import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useI18n } from "@/lib/i18n";

interface AutoFetchFieldProps {
  value?: boolean | null;
  onChange: (value?: boolean | null) => void;
  variant?: "default" | "compact";
  className?: string;
}

export function AutoFetchField({
  value,
  onChange,
  variant = "default",
  className,
}: AutoFetchFieldProps) {
  const { t } = useI18n();

  const stringValue =
    value === null ? "null" : value === true ? "true" : value === false ? "false" : "null";

  const handleChange = (v: string) => {
    onChange(v === "null" ? null : v === "true" ? true : v === "false" ? false : undefined);
  };

  if (variant === "compact") {
    return (
      <Select value={stringValue} onValueChange={handleChange}>
        <SelectTrigger className={className}>
          <SelectValue
            placeholder={`${t("settings.auto_fetch.label")}: ${t("settings.auto_fetch.inherit")}`}
          />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="null">
            {t("settings.auto_fetch.label")}: {t("settings.auto_fetch.inherit")}
          </SelectItem>
          <SelectItem value="true">
            {t("settings.auto_fetch.label")}: {t("settings.auto_fetch.enabled")}
          </SelectItem>
          <SelectItem value="false">
            {t("settings.auto_fetch.label")}: {t("settings.auto_fetch.disabled")}
          </SelectItem>
        </SelectContent>
      </Select>
    );
  }

  return (
    <div className="space-y-1.5">
      <label className="text-[13px] font-medium">
        {t("settings.auto_fetch.label")}
      </label>
      <Select value={stringValue} onValueChange={handleChange}>
        <SelectTrigger className={className ?? "h-10"}>
          <SelectValue placeholder={t("settings.auto_fetch.inherit_from_group")} />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="null">
            {t("settings.auto_fetch.inherit_from_group")}
          </SelectItem>
          <SelectItem value="true">
            {t("settings.auto_fetch.enabled")}
          </SelectItem>
          <SelectItem value="false">
            {t("settings.auto_fetch.disabled")}
          </SelectItem>
        </SelectContent>
      </Select>
    </div>
  );
}
