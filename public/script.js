function activeRoom() {
    return sessionStorage.getItem('room');
}

function setActiveRoom(room) {
    sessionStorage.setItem('room', room);
}

function username() {
    return sessionStorage.getItem('username');
}

function setUsername(username) {
    sessionStorage.setItem('username', username);
}

function connect() {
    const ws = new WebSocket(`ws://${location.host}/chat`);

    ws.addEventListener('close', ev => {
        if (ev.wasClean) {
            showNotification(`connection closed. reason: ${ev.reason}`);
        } else {
            showNotification('connection closed');
        }
    });

    ws.addEventListener('message', e => {
        const msg = decodeData(e.data);

        switch (msg.Type) {
            case 'login':
                const {Username, Rooms, Users, Messages} = msg.Data;

                setUsername(Username);
                setActiveRoom("");

                reload(roomsContainer, Rooms, roomRenderer);
                reload(usersContainer, Users, userRenderer);
                reload(historyContainer, Messages.reverse(), messageRenderer);

                select(find(roomsContainer, ""));

                setVisible(logInSection, false);
                setVisible(chatSection, true);

                sendForm.elements['message'].focus();
                break;
            case 'message':
                const message = msg.Data;
                let room = message.Recipient
                    ? (username() === message.Author ? message.Recipient : message.Author)
                    : message.Recipient;

                if (room === activeRoom()) {
                    messageRenderer(historyContainer, message);
                    historyContainer.scrollTop = historyContainer.scrollHeight;
                } else {
                    const el = find(roomsContainer, room) || roomRenderer(roomsContainer, room);
                    el.classList.add('notify');
                }
                break;
            case 'join':
                userRenderer(usersContainer, msg.Data.Name);
                break;
            case 'leave':
                const el = find(usersContainer, msg.Data.Name);
                el && el.remove();
                break;
            case 'history':
                reload(historyContainer, msg.Data.Messages.reverse(), messageRenderer);
                break;
            case 'error':
                showNotification(msg.Data.Message);
                break;
        }
    });

    return ws;
}

function encodeData(data) {
    return JSON.stringify(data);
}

function decodeData(data) {
    return JSON.parse(data);
}

function showNotification(message) {
    const notification = document.getElementById('notification');
    setVisible(notification, true);
    notification.firstElementChild.innerHTML = message;
    setTimeout(() => {
        setVisible(notification, false);
    }, 4000);
}

function render(parent, template, data) {
    const node = template.content.cloneNode(true);
    const nodeContent = node.querySelector('[data-content]');
    nodeContent.innerHTML = data['text'];
    nodeContent.dataset.id = data['id'];
    parent.appendChild(node);
    return parent.lastElementChild;
}

function createRenderer(template, mapper) {
    return function (parent, item) {
        return render(parent, template, mapper(item))
    }
}

function clear(element) {
    element.innerHTML = '';
}

function reload(parent, items, renderer) {
    clear(parent);

    if (!items) return;

    for (let item of items) {
        renderer(parent, item);
    }
}

function setVisible(element, visible) {
    element && (element.style.display = !!visible ? '' : 'none');
}

function select(element) {
    const selectedSibling = element.parentNode.querySelector('.is-active');
    selectedSibling && selectedSibling.classList.remove('is-active');
    element.classList.add('is-active');
}

function selected(element) {
    return element.classList.contains('is-active');
}

function find(container, content) {
    return container.querySelector(`[data-id="${content}"]`);
}

const historyContainer = document.getElementById('history');
const messageTemplate = document.getElementById('message-template');
const messageRenderer = createRenderer(messageTemplate, msg => {
    return {id: null, text: `${msg.Author === username() ? 'I' : msg.Author}: ${msg.Body}`};
});

const roomsContainer = document.getElementById('rooms');
const roomTemplate = document.getElementById('room-template');
const roomRenderer = createRenderer(roomTemplate, room => {
    return {id: room, text: room || 'Public room'}
});

const usersContainer = document.getElementById('users');
const userTemplate = document.getElementById('user-template');
const userRenderer = createRenderer(userTemplate, user => {
    return {id: user, text: user}
});

const logInSection = document.getElementById('login');
const chatSection = document.getElementById('chat');

const logInForm = document.forms['login'];
const sendForm = document.forms['send'];

let webSocket = connect();

logInForm.addEventListener('submit', e => {
    e.preventDefault();

    webSocket.send(encodeData({
        Type: 'login',
        Data: {
            Name: logInForm.elements['username'].value
        }
    }));
});

sendForm.addEventListener('submit', e => {
    e.preventDefault();

    webSocket.send(encodeData({
        Type: 'message',
        Data: {
            Author: sessionStorage.getItem('username'),
            Recipient: sessionStorage.getItem('room'),
            Body: sendForm.elements['message'].value,
        }
    }));

    sendForm.reset();
});

document.addEventListener('click', e => {
    const el = e.target;

    if (el.matches('.room') && !selected(el)) {
        const room = el.dataset.id;

        select(el);
        setActiveRoom(room);
        el.classList.remove('notify');

        webSocket.send(encodeData({
            Type: 'history',
            Data: {
                User: username(),
                Room: activeRoom()
            }
        }));

        return;
    }

    if (el.matches('.user')) {
        const room = el.dataset.id;
        const roomEl = find(roomsContainer, room) || roomRenderer(roomsContainer, room);

        select(roomEl);
        setActiveRoom(room);

        webSocket.send(encodeData({
            Type: 'history',
            Data: {
                User: username(),
                Room: activeRoom()
            }
        }));
    }
});
