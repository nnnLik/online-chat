const socket = new WebSocket("ws://localhost:8080/ws");
console.log("Attempting WebSocket connection...");

socket.onopen = () => {
    console.log("Success!");
};

socket.onclose = (event) => {
    console.log("Connection closed");
};

socket.onmessage = (event) => {
    const message = event.data;
    const chatBox = document.getElementById("chat-box");
    chatBox.innerHTML += `<p>${message}</p>`;
    chatBox.scrollTop = chatBox.scrollHeight;
};

const chatInput = document.getElementById("chat-input");
chatInput.addEventListener("keypress", (event) => {
    if (event.keyCode === 13) {
        const message = event.target.value;
        if (message != "") {
            socket.send(message);
            event.target.value = "";
        }
    }
});