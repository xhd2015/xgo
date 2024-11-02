import { DiffComputer } from "monaco-editor/esm/vs/editor/common/diff/diffComputer"
import * as fs from "fs/promises"

const oldFile = process.argv[2]
const newFile = process.argv[3]

if (!oldFile || !newFile) {
    console.error(`usage: node ${process.argv[1]} oldFile newFile`)
    process.exit(1)
}

// const pretty = process.env["DIFF_PRETTY"] === 'true'
async function run() {
    const oldContent = await fs.readFile(oldFile, { encoding: "utf-8" })
    const newContent = await fs.readFile(newFile, { encoding: "utf-8" })
    const oldLines = oldContent.split("\n")
    const newLines = newContent.split("\n")

    const computer = new DiffComputer(oldLines, newLines, {
        shouldMakePrettyDiff: true, // pretty is easier to understand, without pretty, the original myers diff output is not easy to understand
        shouldComputeCharChanges: true,
        // shouldComputeCharChanges: false,
        // shouldPostProcessCharChanges: true,
        shouldIgnoreTrimWhitespace: true,
        maxComputationTime: Number.MAX_VALUE, // to unlimit the cost used, return all meaningful changes
    })

    const res = computer.computeDiff()
    console.log(JSON.stringify(res, null, 4))
}

run()