const INFORM_ICON = `<svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 512 512" height="1em" width="1em" xmlns="http://www.w3.org/2000/svg"><path d="M256 90c44.3 0 86 17.3 117.4 48.6C404.7 170 422 211.7 422 256s-17.3 86-48.6 117.4C342 404.7 300.3 422 256 422s-86-17.3-117.4-48.6C107.3 342 90 300.3 90 256s17.3-86 48.6-117.4C170 107.3 211.7 90 256 90m0-42C141.1 48 48 141.1 48 256s93.1 208 208 208 208-93.1 208-208S370.9 48 256 48z"></path><path d="M277 360h-42V235h42v125zm0-166h-42v-42h42v42z"></path></svg>`
const WARNING_ICON = `<svg stroke="currentColor" fill="currentColor" stroke-width="0" viewBox="0 0 24 24" height="1em" width="1em" xmlns="http://www.w3.org/2000/svg"><path fill="none" d="M0 0h24v24H0z"></path><path d="M12 5.99 19.53 19H4.47L12 5.99M12 2 1 21h22L12 2z"></path><path d="M13 16h-2v2h2zM13 10h-2v5h2z"></path></svg>`

function getById(name) {
    const v = document.getElementById(name)
    if (!v) {
        throw new Error(`could find '${name}'`)
    }
    return v
}


const toast = getById("toast")
const toastIcon = getById("toast-icon")
const toastMsg = getById("toast-msg")
function showToast(icon, msg) {
    toast.style.bottom = "1em";
    const t = typeof msg
    const aux = msg
    if (t !== 'string' && t !== 'number') {
        aux = JSON.stringify(msg)
    }
    toastIcon.innerHTML = icon
    toastMsg.innerHTML = aux

    setTimeout(() => {
        toast.style.bottom = "-5em";
    }, 1000)
}

function log(msg) {
    const border = 'rgb(71,170,123)'
    const bg = 'rgb(236,252,244)'
    toastIcon.style.color = border
    toast.style.backgroundColor = bg
    toast.style.borderColor = border;
    console.log(msg)
    showToast(INFORM_ICON, msg)
}

function error(e) {
    const border = 'rgb(224,120,115)'
    const bg = 'rgb(252,241,240)'
    toastIcon.style.color = border
    toast.style.backgroundColor = bg
    toast.style.borderColor = border;
    console.error(e)
    showToast(WARNING_ICON, e)
}

const messagesDisplay = getById('messages-display')
const messagesHeader = getById('messages-header')
const addColumnButton = getById("add-column-button")
const fieldName = getById("field-name")
const fieldFormat = getById("field-format")
const fieldRequired = getById("field-required")
const fieldDescription = getById("field-description")
const modelName = getById("model-name")
const systemPrompt = getById("system-prompt")

var format = null


const messages = []


//////////////////////////
// SCHEMA HANDLING
//////////////////////////
function addFieldToFormat() {
    const newFieldName = fieldName.value.trim()
    if (newFieldName.length === 0) {
        error("A field name is required for adding a new field")
        return
    }
    const newFieldFormat = fieldFormat.value
    const newFieldDescription = fieldDescription.value
    const required = fieldRequired.value === 'required'

    if (!format || format.length === 0) {
        format = []
    }
    const index = format.findIndex(f => f.name === newFieldName);
    if (index !== -1) {
        error(`Field named '${newFieldName}' already exists`)
        return
    }

    format.push({
        name: newFieldName,
        type: newFieldFormat,
        description: newFieldDescription,
        required,
    })


    fieldName.value = ""
    fieldDescription.value = ""
    reRenderMessages()
}
function deleteField(name) {
    if (!format || format.length === 0) {
        return
    }
    const index = format.findIndex((f) => f.name === name);
    if (index !== -1) {
        format.splice(index, 1);
    }

}
addColumnButton.onclick = addFieldToFormat


//////////////////////////
// END OF SCHEMA HANDLING
//////////////////////////

//////////////////////////
// WEB SOCKET CONNECTION
//////////////////////////
let protocol = window.location.protocol
let ws = "ws"
if (protocol === "https:") {
    ws = "wss"
}
const socket = new WebSocket(ws + "://" + window.location.host + "/socket");
socket.onopen = function (event) {
    console.log("WebSocket connection established!");

    // Example: Send a message to the server
    socket.send("Hello from the client!");
};
socket.onmessage = function (event) {
    try {
        const payload = JSON.parse(event.data);
        console.log(JSON.stringify(payload, null, 2))
        if (payload && payload.id && payload.response) {
            // Find the message with the matching ID
            const messageIndex = messages.findIndex((msg) => msg.id === payload.id);
            if (messageIndex !== -1) {
                if (format && format.length > 0) {
                    messages[messageIndex].veredict = JSON.parse(payload.response);
                } else {
                    messages[messageIndex].veredict = payload.response
                }
                reRenderMessages()
            } else {
                error(`Message with ID ${payload.id} not found.`);
            }
        } else {
            error(`Invalid payload received: ${payload}`);
        }
    } catch (error) {
        error("Error parsing JSON:" + error);
    }
};

socket.onclose = function (event) {
    if (event.wasClean) {
        console.log(`Connection closed cleanly, code=${event.code} reason=${event.reason}`);
    } else {
        error('Connection died');
    }
};

socket.onerror = function (error) {
    error("WebSocket error:" + error);
};

//////////////////////////
// END OF WEB SOCKETS
//////////////////////////


//////////////////////////
// MESSAGES STATE
//////////////////////////


const pendingMsg = '<small style="opacity:0.6">...pending</small>'
/**
 * Adds a new message to the list of messages
 * @param {Object} m the message to add to the rendering
 */
function appendMsg(m) {
    let row = document.createElement('tr')

    let id = document.createElement('td')
    id.innerText = m.id
    row.appendChild(id)

    let content = document.createElement('td')
    content.innerText = m.content
    row.appendChild(content)


    if (format && format.length > 0) {
        format.forEach((f) => {
            let td = document.createElement('td')
            if (m.veredict !== undefined && m.veredict[f.name] !== undefined) {
                const v = m.veredict[f.name]
                console.log(typeof v)
                switch (typeof v) {
                    case 'boolean':
                        td.innerHTML = `<code>${v}</code>`
                        break
                    default:
                        td.innerText = v
                }

            } else {
                td.innerHTML = pendingMsg
            }
            row.appendChild(td)

        })
    } else {
        let td = document.createElement('td')
        if (m.veredict) {
            td.innerHTML = JSON.stringify(m.veredict)
        } else {
            td.innerHTML = pendingMsg
        }
        row.appendChild(td)
    }

    messagesDisplay.appendChild(row)




}

function getSystemPrompt() {
    const v = systemPrompt.value.trim()
    if (v.length === 0) {
        return "You are a helpful assistant."
    }
    return v
}

/**
 * Re renders the messages on the UI
 */
function reRenderMessages() {
    // Header
    messagesHeader.innerHTML = '<th><div>ID</div></th><th><div>Content</div></th>'
    if (format && format.length > 0) {
        // Add fields
        format.forEach((f) => {
            let th = document.createElement('th')
            th.innerHTML = `<div>
                <span>
                    ${f.name}<sup class='warning'>${f.required ? '*' : ''}</sup>
                </span>
            <code>${f.type}</codes></div>`
            messagesHeader.appendChild(th)
        })
    } else {
        let th = document.createElement('th')
        th.innerHTML = "Response"
        messagesHeader.appendChild(th)
    }
    // Body
    messagesDisplay.innerHTML = ''
    messages.forEach(m => appendMsg(m))
}
reRenderMessages()
//////////////////////////
// END OF MESSAGES STATE
//////////////////////////

//////////////////////////
// SEND A MESSAGE
//////////////////////////
const writer = document.getElementById("msgwriter")
if (!writer) {
    throw new Error("could find a textarea to write messages in")
}

function structuredOutput() {
    if (!format || format.length === 0) {
        return null
    }

    let ret = {
        type: 'object',
        properties: {},
        required: []
    }
    format.forEach((f) => {
        ret.properties[f.name] = {
            type: f.type,
            description: f.description
        }
        if (f.required) {
            ret.required.push(f.name)
        }
    })
    return ret
}

/**
 * sends a message to the queue
 * @param {string} msg the message to send
 */
async function submit(msg) {
    // return if empty
    if (msg.trim().length === 0) {
        return
    }
    let model = modelName.value.trim()
    if (model.length === 0) {
        error("A model name should be provided")
        return
    }

    // Append as pending
    let id = `${messages.length}`
    let aux = {
        content: msg.trim(),
        id
    }
    messages.push(aux)
    appendMsg(aux)




    // Shape the message    
    let body = JSON.stringify({
        model,
        prompt: msg,
        id,
        webhook: protocol + '//' + "playground:3000" + '/webhook',
        format: structuredOutput(),
        system: getSystemPrompt(),
    });
    // queue message
    const ret = await fetch("/msg", {
        method: "POST",
        body,
        headers: {
            'Content-Type': 'application/json'
        }
    })
    if (!ret.ok) {
        error(ret)
    }
}

writer.onkeyup = function (e) {
    if (e.key === 'Enter') {
        let v = e.target.value.trim()
        submit(v)
        e.target.value = ''
    }
}
//////////////////////////
// END OF SEND MESSAGE
//////////////////////////

