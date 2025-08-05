// src/console-bridge.js
import { LogInfo, LogDebug, LogError } from "../wailsjs/runtime/runtime";

// Preserve originals
const oldLog   = console.log.bind(console);
const oldDebug = console.debug.bind(console);
const oldError = console.error.bind(console);

// Helper: flatten any arguments into a single string
const fmt = (...args: any[]) =>
  args.map(a => (typeof a === "object" ? JSON.stringify(a) : String(a))).join(" ");

// Override console.log ➜ LogInfo
console.log = (...args) => {
  oldLog(...args);               // still see it in DevTools
  LogInfo(fmt(...args));         // send to Wails logger[1]
};

// Override console.debug ➜ LogDebug
console.debug = (...args) => {
  oldDebug(...args);
  LogDebug(fmt(...args));        // Wails debug log[1]
};

// Override console.error ➜ LogError
console.error = (...args) => {
  oldError(...args);
  LogError(fmt(...args));        // Wails error log[1]
};