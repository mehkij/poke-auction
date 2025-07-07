import { MdLightMode, MdDarkMode } from "react-icons/md";
import { useTheme } from "../hooks/useTheme";

function ThemeToggle() {
  const { theme, toggle } = useTheme();

  return (
    <button onClick={toggle}>
      {theme === "dark" ? (
        <MdLightMode className="size-6 cursor-pointer" />
      ) : (
        <MdDarkMode className="size-6 cursor-pointer" />
      )}
    </button>
  );
}

export default ThemeToggle;
