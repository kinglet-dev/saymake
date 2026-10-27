// latest wraps fn so a burst of calls runs it once, with the last call's
// arguments, delay ms after the burst (for rebuilding while a slider moves).
export function latest(fn, delay, timers = globalThis) {
  let id = null;
  return (...args) => {
    if (id !== null) timers.clearTimeout(id);
    id = timers.setTimeout(() => {
      id = null;
      fn(...args);
    }, delay);
  };
}
