import { DiffComputer } from "monaco-editor/esm/vs/editor/common/diff/diffComputer"

export interface Request {
    OldLines: string[]
    NewLines: string[]

    Pretty?: boolean // default true
    ComputeChar?: boolean // default false
}
// export interface Result {
//     QuitEarly: boolean
//     Changes: LineChange[]
// }

// export interface LineChange {
//     OriginalStartLineNumber: number
//     OriginalEndLineNumber: number
//     ModifiedStartLineNumber: number
//     ModifiedEndLineNumber: number
// }

function run() {
    // take request from global
    const request: Request = globalThis.request
    const { OldLines, NewLines, ComputeChar, Pretty } = request || {}

    // create the diff computer
    const computer = new DiffComputer(OldLines || [], NewLines || [], {
        shouldMakePrettyDiff: Pretty !== false,
        shouldComputeCharChanges: ComputeChar,
        shouldIgnoreTrimWhitespace: true,
        maxComputationTime: Number.MAX_VALUE, // to unlimit the cost used, return all meaningful changes
    })

    // run the diff
    const res = computer.computeDiff()

    // now we need to make res to upper case names
    const changes = []
    res?.changes?.forEach?.(ch => {
        changes.push({
            OriginalStartLineNumber: ch.originalStartLineNumber,
            OriginalEndLineNumber: ch.originalEndLineNumber,
            ModifiedStartLineNumber: ch.modifiedStartLineNumber,
            ModifiedEndLineNumber: ch.modifiedEndLineNumber,
        })
    })
    return { QuitEarly: !!res?.quitEarly, Changes: changes }
}

globalThis.run = run