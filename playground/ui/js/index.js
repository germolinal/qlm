
//////////////////////////
// MESSAGES STATE
//////////////////////////
const messagesDisplay = document.getElementById('messages-display')
if (!messagesDisplay) {
    throw new Error("could find main container of messages")
}
const messages = [
    {
        content: "This is a resolved message",
        id: '0',
        veredict: { wow_msg: 'Yes, very good' }
    },
    {
        content: "Pending messsage",
        id: '1',
    }
]
/**
 * Adds a new message to the list of messages
 * @param {Object} m the message to add to the rendering
 */
function appendMsg(m) {
    let row = document.createElement('tr')

    let content = document.createElement('td')
    content.innerText = m.content
    row.appendChild(content)

    let id = document.createElement('td')
    id.innerText = m.id
    row.appendChild(id)

    let veredict = document.createElement('td')
    if (m.veredict) {
        veredict.innerText = JSON.stringify(m.veredict)
    } else {
        veredict.innerHTML = '<small style="opacity:0.6">... Pending response</small>'
    }
    row.appendChild(veredict)

    messagesDisplay.appendChild(row)
}
/**
 * Re renders the messages on the UI
 */
function reRenderMessages() {
    // Clean
    messagesDisplay.innerHTML = ''
    // Append messages
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
/**
 * sends a message to the queue
 * @param {string} msg the message to send
 */
async function submit(msg) {
    // return if empty
    if (msg.trim().length === 0) {
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
        prompt: msg,
        id,
        webhook: 'http://localhost:3001/webhook',
        format: {
            type: 'object',
            properties: {
                inappropriate: {
                    type: 'boolean',
                    description:
                        'is this email inappropriate for a professional situation?'
                },
                contains_pii: {
                    type: 'boolean',
                    description:
                        'does the email contain Personally Identifiable Information or client data?'
                }
            },
            required: ['inappropriate', 'contains_pii']
        }
    }
    // queue message
    console.log(m)
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

