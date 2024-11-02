import { DiffComputer } from "monaco-editor/esm/vs/editor/common/diff/diffComputer"

export interface Request {
    oldLines: string[]
    newLines: string[]

    pretty?: boolean // default true
    computeChar?: boolean // default false
}

async function run() {
    let data = ''
    process.stdin.resume()
    process.stdin.on('data', e => {
        data += e
    })
    process.stdin.on('end', () => {
        const req = JSON.parse(data) as Request

        const { oldLines, newLines, computeChar, pretty } = req || {}

        const computer = new DiffComputer(oldLines || [], newLines || [], {
            shouldMakePrettyDiff: pretty !== false,
            shouldComputeCharChanges: computeChar,
            shouldIgnoreTrimWhitespace: true,
            maxComputationTime: Number.MAX_VALUE, // to unlimit the cost used, return all meaningful changes
        })

        const res = computer.computeDiff()
        process.stdout.write(JSON.stringify(res))
        process.exit(0)
    })
}
run()