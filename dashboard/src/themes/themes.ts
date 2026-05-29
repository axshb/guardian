export const THEME_IDS = [
  'gruvbox', 'nord', 'catppuccin', 'rose-pine', 'one-dark',
  'everblush', 'tokyo-night', 'dracula', 'carbon-fox',
  'light', 'catppuccin-latte', 'amoled',
];

export function themeLabel(id: string): string {
  return id
    .split('-')
    .map((w) => w.charAt(0).toUpperCase() + w.slice(1))
    .join(' ');
}

export function themePreview(id: string) {
  const prev = document.documentElement.dataset.theme;
  document.documentElement.dataset.theme = id;
  const style = getComputedStyle(document.documentElement);
  const result = {
    primary: style.getPropertyValue('--primary').trim() || '#ccc',
    bg: style.getPropertyValue('--background').trim() || '#111',
    accent: style.getPropertyValue('--success').trim() || '#888',
  };
  if (prev) document.documentElement.dataset.theme = prev;
  return result;
}
