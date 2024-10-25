import * as readline from "node:readline/promises"

// see https://nodejs.org/docs/v0.4.7/api/process.html#process.stdin
process.stdin.resume()

const rl = readline.createInterface({ input: process.stdin });

const callbacks: ((line: string, e: any, parseErr: Error) => void)[] = []
const onCloseCallbacks: (() => void)[] = []

rl.on("line", (line) => {
  try {
    const object = JSON.parse(line);
    callbacks?.forEach?.(fn => fn(line, object, undefined))
  } catch (e) {
    callbacks?.forEach?.(fn => fn(line, undefined, e))
  }
});

// NOTE: in nodejs,  when process meets ^D(stdin closed), it naturelly exits.
rl.on("close", () => {
  onCloseCallbacks?.forEach?.(e => e())
})

process.stdout.on('end', function () {
  // console.error("stdout end")
  // process.exit();
});

export default function onObject(fn: (line: string, e: any, parseErr: Error) => void) {
  callbacks.push(fn)
}

export function onClose(fn: () => void) {
  onCloseCallbacks.push(fn)
}