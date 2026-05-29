import { ref } from 'vue';

interface Toast {
  id: number;
  severity: 'success' | 'error' | 'info' | 'warn';
  summary: string;
  life: number;
}

const toasts = ref<Toast[]>([]);
let idCounter = 0;

export function useToast() {
  function add(opts: { severity: Toast['severity']; summary: string; life?: number }) {
    const id = ++idCounter;
    const toast: Toast = { id, severity: opts.severity, summary: opts.summary, life: opts.life ?? 3000 };
    toasts.value.push(toast);
    setTimeout(() => {
      toasts.value = toasts.value.filter(t => t.id !== id);
    }, toast.life);
  }
  return { add, toasts };
}
