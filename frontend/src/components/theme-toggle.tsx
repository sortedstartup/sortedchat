import { Moon, Sun } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useTheme } from "@/components/theme-provider";

export function ThemeToggle() {
  const { theme, setTheme } = useTheme();

  const toggleTheme = () => {
    setTheme(theme === "light" ? "dark" : "light");
  };

  return (
    <Button
      variant="outline"
      size="default"
      onClick={toggleTheme}
      className="gap-2"
    >
      {theme === "light" ? (
        <>
          <Sun className="h-[1.2rem] w-[1.2rem]" />
          <span>Light</span>
        </>
      ) : (
        <>
          <Moon className="h-[1.2rem] w-[1.2rem]" />
          <span>Dark</span>
        </>
      )}
    </Button>
  );
}