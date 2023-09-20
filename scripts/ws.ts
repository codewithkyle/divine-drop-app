import alerts from "./alerts";

let socket:WebSocket;
let connected = false;
let wasReconnection = false;

async function connect() {
    if (connected){
        return;
    }
    // @ts-expect-error
    const { SOCKET_URL, ENV } = await import("/static/config.js");
    try{
        socket = new WebSocket(SOCKET_URL);
        await new Promise((resolve) => {
            if (wasReconnection){
                alerts.success("Reconnected", "We've reconnected with the server.");
            }
            wasReconnection = false;
            socket.onopen = () => {
                console.log("Connected to server.");
                connected = true;
                resolve(null);
            };
        });
    } catch (e) {
        console.error(e);
    }
    socket.addEventListener("message", async (event) => {
        try {
            if (ENV === "dev"){
                console.log(event.data);
            }
        } catch (e) {
            console.error(e, event);
        }
    });
    socket.addEventListener("close", () => {
        disconnect();
    });
}

function disconnect() {
    if (connected) {
        alerts.warn("Connection Lost", "Hang tight we've lost the server connection. Any changes you make will be synced when you've reconnected.");
        connected = false;
        wasReconnection = true;
    }
    setTimeout(() => {
        connect();
    }, 5000);
}

function send(msg:any):void{
    if (connected){
        console.log("Sending: ", msg);
        socket.send(msg);
    }
    else if (!connected && wasReconnection){
        // We lost connection, check if we were in a game
        console.log("Not connected, but we were reconnected.");
    }
    else {
        // Do... something... maybe?
        console.log("Not connected, and we weren't reconnected.");
    }
}

function close():void{
    if (connected){
        socket.close();
    }
}

export { connected, disconnect, connect, send, close };
