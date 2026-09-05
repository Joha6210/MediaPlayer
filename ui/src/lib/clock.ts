// src/lib/clock.ts
import { readable } from 'svelte/store';

export const clock = readable(new Date(), (set) => {
    const interval = setInterval(() => set(new Date()), 1000);
    return () => clearInterval(interval);
});