const extMap: Record<string, string> = {
  'image/jpeg': 'jpg',
  'image/png': 'png',
  'image/gif': 'gif',
  'image/webp': 'webp',
  'image/svg+xml': 'svg',
  'image/avif': 'avif',
  'image/bmp': 'bmp',
  'image/tiff': 'tiff',
};

export function extFromContentType(ct: string | undefined): string {
  if (!ct) return '';
  const ext = extMap[ct.toLowerCase()];
  return ext ? `.${ext}` : '';
}

export function assetUrl(id: string, ct: string | undefined): string {
  return `/assets/${id}${extFromContentType(ct)}`;
}

export function artUrl(id: string, ct: string | undefined): string {
  return `/art/${id}${extFromContentType(ct)}`;
}
