// this script runs after window.onload
// const traces = {}

// trace example:
///   {"FuncInfo":{"Pkg":"github.com/xhd2015/xgo","IdentityName":"TestHelloWorld","Name":"TestHelloWorld","RecvType":"","RecvPtr":false,"Generic":false,"RecvName":"","ArgNames":["t"],"ResNames":[],"FirstArgCtx":false,"LastResultErr":false},"Begin":0,"End":0,"Args":{"t":{}},"Results":{},"Children":null}

let selectedID = ""

function getHeadID(id) {
    return `head_${id}`
}
function getToggleID(id) {
    return `toggle_${id}`
}
function getTraceListID(id) {
    return `trace_list_${id}`
}

function onClickHead(id) {
    const headID = getHeadID(id)
    const el = document.getElementById(headID)
    if (!el) {
        return
    }
    const classList = el.classList
    if (classList.contains("selected")) {
        classList.remove("selected")
    } else {
        classList.add("selected")
    }

    if (selectedID !== headID) {
        const prev = document.getElementById(selectedID)
        if (prev) {
            prev.classList.remove("selected")
        }
        selectedID = headID
        const info = document.getElementById("detail-info")
        const infoPkg = document.getElementById("detail-info-pkg")
        const infoFunc = document.getElementById("detail-info-func")
        const req = document.getElementById("detail-request")
        const resp = document.getElementById("detail-response")
        const traceData = traces[id]

        if (traceData.error) {
            infoPkg.innerText = "<unknown>"
            infoFunc.innerText = "<unknown>"
            req.value = traceData.error
            resp.value = ''
            return
        }

        infoPkg.innerText = traceData.FuncInfo?.Pkg || ""
        infoFunc.innerText = traceData.FuncInfo?.IdentityName || ""
        req.value = JSON.stringify(traceData.Args, null, "    ")
        if (traceData.Error) {
            let msg = traceData.Error
            if (!msg.includes("err")) {
                msg = "error:" + msg
            }
            resp.value = msg
        } else {
            resp.value = JSON.stringify(traceData.Results, null, "    ")
        }
    }
}

function onClickToggle(e, id) {
    e.stopPropagation()

    setToggle(id, true)
    toggleTraceList(id, true)
}

function setToggle(id, toggle, collapsed) {
    const el = document.getElementById(getToggleID(id))
    if (!el) {
        return
    }
    const cl = el.classList
    if (toggle) {
        if (cl.contains("down")) {
            collapsed = true
        } else if (cl.contains("right")) {
            collapsed = false
        }
    }
    if (collapsed) {
        cl.remove("down")
        cl.add("right")
    } else {
        cl.add("down")
        cl.remove("right")
    }
}

function onClickExpandAll(e) {
    const el = document.getElementById("toolbar")
    const toggleAllOn = "toggle-all-on"
    if (el.classList.contains(toggleAllOn)) {
        el.classList.remove(toggleAllOn)
        // collapse all
        for (const id of ids) {
            toggleTraceList(id, false, true)
            setToggle(id, false, true)
        }
    } else {
        // expand all
        el.classList.add(toggleAllOn)
        for (const id of ids) {
            toggleTraceList(id, false, false)
            setToggle(id, false, false)
        }
    }
}

function toggleTraceList(id, toggle, collapsed) {
    const traceList = document.getElementById(getTraceListID(id))
    if (!traceList) {
        return
    }
    if (toggle) {
        traceList.classList.toggle("collapsed")
        return
    }
    if (collapsed) {
        traceList.classList.add("collapsed")
        return
    }
    traceList.classList.remove("collapsed")
}

// for debugging
window.traces = traces
window.onClickHead = onClickHead
window.onClickToggle = onClickToggle
window.onClickExpandAll = onClickExpandAll

window.shit = function () {
    debugger
    alert("shit")
}
onClickHead("2")
