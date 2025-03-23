function getById(name) {
    const v = document.getElementById(name)
    if (!v) {
        throw new Error(`could find '${name}'`)
    }
    return v
}

function onError(e) {
    console.error(e)
}

const messagesDisplay = getById('messages-display')
const messagesHeader = getById('messages-header')
const addColumnButton = getById("add-column-button")
const fieldName = getById("field-name")
const fieldFormat = getById("field-format")
const fieldRequired = getById("field-required")
const fieldDescription = getById("field-description")
const modelName = getById("model-name")

var format = [{
    name: "is_spanish",
    type: "boolean",
    description: "indicates whether the input was in spanish",
    required: true
}]
// var format = null

const messages = [
    {
        content: "This is a resolved message",
        id: '0',
        veredict: "{ wow_msg: 'Yes, very good' }"
    },
    {
        content: "Pending messsage",
        id: '1',
    }
]

//////////////////////////
// SCHEMA HANDLING
//////////////////////////
function addFieldToFormat() {
    const newFieldName = fieldName.value.trim()
    if (newFieldName.length === 0) {
        onError("A field name is required for adding a new field")
    }
    const newFieldFormat = fieldFormat.value
    const newFieldDescription = fieldDescription.value
    const required = fieldRequired.value === 'required'

    if (!format || format.length === 0) {
        format = []
    }
    const index = format.findIndex(f => f.name === newFieldName);
    if (index !== -1) {
        onError(`Field named ${newFieldName} already exists`)
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
                onError(`Message with ID ${payload.id} not found.`);
            }
        } else {
            onError(`Invalid payload received: ${payload}`);
        }
    } catch (error) {
        onError("Error parsing JSON:" + error);
    }
};

socket.onclose = function (event) {
    if (event.wasClean) {
        console.log(`Connection closed cleanly, code=${event.code} reason=${event.reason}`);
    } else {
        onError('Connection died');
    }
};

socket.onerror = function (error) {
    onError("WebSocket error:" + error);
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
            if (m.veredict && m.veredict[f.name]) {
                td.innerText = m.veredict[f.name]
            } else {
                td.innerHTML = pendingMsg
            }
            row.appendChild(td)

        })
    } else {
        let td = document.createElement('td')
        if (m.veredict) {
            td.innerHTML = m.veredict
        } else {
            td.innerHTML = pendingMsg
        }
        row.appendChild(td)
    }

    messagesDisplay.appendChild(row)




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
        onError("A model name should be provided")
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
    let m = {
        model,
        prompt: msg,
        id,
        webhook: protocol + '//' + window.location.host + '/webhook',
        format: structuredOutput(),
    }

    // queue message
    const ret = await fetch("/msg", {
        method: "POST",
        body: JSON.stringify(m),
        headers: {
            'Content-Type': 'application/json'
        }
    })
    if (!ret.ok) {
        onError(ret)
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

