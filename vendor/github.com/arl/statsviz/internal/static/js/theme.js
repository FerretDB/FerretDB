/**
 * Get color theme based on previous user choice or browser theme
 */
export const getThemeMode = () => {
  let themeMode = localStorage.getItem("theme-mode");

  if (themeMode === null) {
    const isDark =
      window.matchMedia &&
      window.matchMedia("(prefers-color-scheme: dark)").matches;
    themeMode = (isDark && "dark") || "light";

    localStorage.setItem("theme-mode", themeMode);
  }

  return themeMode;
};

/**
 * Set light or dark theme
 */
export const updateThemeMode = () => {
  if (getThemeMode() === "dark") {
    document.body.classList.add("dark-theme");
    document
      .getElementById("navbar")
      .classList.replace("navbar-light", "navbar-dark");
    document.getElementById("navbar").classList.replace("bg-light", "bg-dark");
  } else {
    document.body.classList.remove("dark-theme");
    document
      .getElementById("navbar")
      .classList.replace("navbar-dark", "navbar-light");
    document.getElementById("navbar").classList.replace("bg-dark", "bg-light");
  }
};
