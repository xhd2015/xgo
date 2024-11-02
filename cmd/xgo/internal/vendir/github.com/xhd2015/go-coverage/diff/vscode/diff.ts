import { DiffComputer } from "monaco-editor/esm/vs/editor/common/diff/diffComputer"
import onObject from "./stream"
import * as fs from "fs/promises"

export interface Request {
    id: string
    ping?: boolean // is this a ping request?
    oldLines: string[]
    newLines: string[]

    exit?: boolean

    pretty?: boolean // default true
    computeChar?: boolean // default false
}
export interface Response {
    error?: string
}

// not working
// process.on('disconnect', function () {
// process.stderr.write("parent exited \n")
// process.exit();
// });


const exitAfterPingTimeout = process.env["EXIT_AFTER_PING_TIMEOUT"] === 'true'
const disableDebugLog = process.env["DISABLE_DEBUG_LOG"] === 'true'

let lastPingTime: number
if (exitAfterPingTimeout) {
    setInterval(() => {
        if (!lastPingTime || Date.now() - lastPingTime > 10 * 1000) {
            if (!disableDebugLog) {
                process.stderr.write("ping lost for 10s, will exit now.\n", () => {
                    process.exit()
                })
            } else {
                process.exit()
            }
        }
    }, 10 * 1000)
}

function debugLog(...log: any[]) {
    if (disableDebugLog) {
        return
    }
    console.error(...log)
}

const responseIDPrefix = process.env["RESPONSE_ID_PREFIX"] === 'true'
onObject((line: string, req: Request, err: Error) => {
    if (!disableDebugLog) {
        debugLog("received line:", line, err)
    }
    if (process.env["ENABLE_LOG"] === 'true') {
        fs.writeFile("./debug.log", "req: " + JSON.stringify(req) + "\n")
    }
    lastPingTime = Date.now()
    if (err) {
        const resp: Response = { error: err.message }
        if (responseIDPrefix) {
            // for id unknown, the parent can discard it, and the request can abort after 1s timeout
            process.stdout.write("unknown:")
        }
        process.stdout.write(JSON.stringify(resp))
        process.stdout.write("\n")
        return
    }
    const { id, ping, oldLines, newLines, computeChar, pretty, exit } = req || {}
    if (exit) {
        process.exit(1)
    }
    if (ping) {
        // no response on ping
        if (!disableDebugLog) {
            debugLog("on ping:", id)
        }
        return
    }

    const computer = new DiffComputer(oldLines || [], newLines || [], {
        shouldMakePrettyDiff: pretty !== false,
        shouldComputeCharChanges: computeChar,
        shouldIgnoreTrimWhitespace: true,
        maxComputationTime: Number.MAX_VALUE, // to unlimit the cost used, return all meaningful changes
    })

    const res = computer.computeDiff()
    if (responseIDPrefix) {
        process.stdout.write(id || "unknown")
        process.stdout.write(":")
    }
    process.stdout.write(JSON.stringify(res))
    process.stdout.write("\n")
    // avoid console.log, which may contain colors
})