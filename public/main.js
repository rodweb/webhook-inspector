class State {
    constructor() {
        this.eventSource = new EventSource('/sse');
        this.requests = {};
        this.notify = false;
    }

    addRequest(req) {
        this.requests[req.id] = req;
    }

    get requestCount() {
        return Object.keys(this.requests).length;
    }
}

const state = new State()

document.addEventListener('DOMContentLoaded', function () {
    document.getElementById('clear-requests').onclick = clearRequests

    setupListeners(state.eventSource)

    setInterval(updateConnectionStatus, 1000)
    setInterval(updateRelativeTime, 5000)
    setTimeout(askNotificationPermission, 1000)
})

function askNotificationPermission() {
    if (Notification.permission === 'default') {
        Notification.requestPermission().then(permission => {
            if (permission === 'granted') {
                console.log('Notification permission granted')
                state.enableNotifications()
            } else {
                console.log('Notification permission denied:', permission)
            }
        })
    }
}

function updateRelativeTime() {
    for (const req of Object.values(state.requests)) {
        const time = document.getElementById(req.id).querySelector('.relative-time');
        updateIfChanged(time, relativeTime(req.timestamp))
    }
}

function updateIfChanged(element, value) {
    if (element.innerHTML !== value) {
        element.innerHTML = value;
    }
}

function setupListeners(eventSource) {
    eventSource.addEventListener('message', function (event) {
        console.log('Message', event.data)
        const req = JSON.parse(event.data);
        addRequest(req);
    })

    eventSource.addEventListener('open', function (event) {
        console.log('Connection opened');
    })

    eventSource.addEventListener('close', function (event) {
        console.log('Connection closed');
    })

    eventSource.addEventListener('error', function (event) {
        console.log('Error', event);
    })
}

function addRequest(req) {
    const list = document.getElementById('request-list');
    const request = document.createElement('button');
    request.classList.add('inline-flex', 'items-center', 'justify-center', 'whitespace-nowrap', 'text-sm', 'ring-offset-background', 'transition-colors', 'focus-visible:outline-none', 'focus-visible:ring-2', 'focus-visible:ring-ring', 'focus-visible:ring-offset-2', 'disabled:pointer-events-none', 'disabled:opacity-50', 'hover:bg-accent', 'hover:text-accent-foreground', 'h-9', 'px-3', 'text-left', 'font-normal', 'gap-2', 'rounded-none');
    request.innerHTML = req.method + ' ' + req.endpoint;
    request.setAttribute('id', req.id);

    request.onclick = function (event) {
        showRequest(req);
    }

    const time = document.createElement('span');
    time.classList.add('ml-auto', 'text-xs', 'text-gray-500', 'dark:text-gray-400', 'relative-time');
    time.innerHTML = relativeTime(req.timestamp)

    request.appendChild(time);

    const first = list.firstChild;
    if (first) {
        list.insertBefore(request, first);
    } else {
        list.appendChild(request);
    }

    state.addRequest(req)
    updateRequestCount();

    if (Notification.permission === 'granted') {
        sendNotification(req.method + ' ' + req.endpoint)
    }
}

function sendNotification(body) {
    new Notification("Webhook Inspector", { body })
}

function updateConnectionStatus() {
    const readyState = state.eventSource.readyState
    let status = 'disconnected'
    switch (readyState) {
        case EventSource.CONNECTING:
            status = 'connecting';
            break;
        case EventSource.OPEN:
            status = 'connected';
            break;
        case EventSource.CLOSED:
            status = 'disconnected';
            break;
    }

    const statusDiv = document.getElementById('connection-status');
    updateIfChanged(statusDiv, status)
}

function updateRequestCount() {
    const count = document.getElementById('request-count');
    count.innerHTML = state.requestCount.toString()
}

function updateHeaders(headers) {
    const list = document.getElementById('request-headers');
    list.innerHTML = '';

    for (const [key, value] of Object.entries(headers)) {
        const item = document.createElement('div');
        item.classList.add('grid', 'gap-0.5');

        const keyDiv = document.createElement('div');
        keyDiv.classList.add('font-semibold');
        keyDiv.innerHTML = key;

        const valueDiv = document.createElement('div');
        valueDiv.innerHTML = value.toString();

        item.appendChild(keyDiv);
        item.appendChild(valueDiv);

        list.appendChild(item);
    }
}

function updateBody(content) {
    const body = document.getElementById('request-body');
    body.innerHTML = content;

    if (content === '') {
        body.parentElement.classList.add('hidden')
    } else {
        body.parentElement.classList.remove('hidden')
    }
}

function showRequest(req) {
    updateHeaders(req.headers)
    updateBody(indentJSON(req.body))
}

function indentJSON(content) {
    try {
        const obj = JSON.parse(content);
        return JSON.stringify(obj, null, 2);
    } catch (e) {
        return content;
    }
}

function clearRequests() {
    const list = document.getElementById('request-list');
    list.innerHTML = '';

    updateHeaders({})
    updateBody('')

    state.requests = {};
    updateRequestCount();
}

function relativeTime(ts) {
    const now = new Date();
    const diff = now - new Date(ts);

    const seconds = Math.floor(diff / 1000);
    const minutes = Math.floor(seconds / 60);
    const hours = Math.floor(minutes / 60);
    const days = Math.floor(hours / 24);

    if (days > 0) {
        return `${days} days ago`
    } else if (hours > 0) {
        return `${hours} hours ago`
    } else if (minutes > 0) {
        return `${minutes} minutes ago`
    } else {
        return `${seconds} seconds ago`
    }
}
